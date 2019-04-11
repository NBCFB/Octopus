package Octopus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	DefaultEnvVar = "OCTOPUS_LISTENER"
	DefaultNetwork = "tcp"
	DefaultAwaitTimeout =  5 * time.Second
)
/*
listenerDescriptor defines a listener descriptor file. A descriptor file is created when a child process is forked.
 */
type listenerDescriptor struct {
	Addr	string	`json:"addr"`
	FD		int		`json:"FD"`
	Name	string	`json:"Name"`
}

/*
GracefulServer defines a HTTP server. The reason we do not make something like

	type GracefulServer struct {
		http.Server
		...
	}
is because we want to give user freedom to make their own and we can serve it as long as it is a http.Server.
 */
type GracefulServer struct {
	Addr 		string
	PID  		int
	Server		*http.Server
	Listener	net.Listener
}

/*
GracefulServe starts a HTTP server. It receives a http.Server server passed by user and an indicators killMaster.
It first create a listener (either a new one or a imported one). Then it starts a goroutine for the server to start
accepting connections. Any hooked signals will be handled in handleSignals(...).
 */
func GracefulServe(server *http.Server, killMaster bool) (gs *GracefulServer, err error) {

	srv := &GracefulServer{
		Addr:   server.Addr,
		Server: server,
	}

	err = srv.createListener()
	if err != nil {
		log.Fatalf("[ERR] Unable to create a listener: %v.\n", err)
	}

	go srv.Server.Serve(srv.Listener)

	pid := syscall.Getpid()
	srv.PID = pid
	log.Printf("[INFO] The server has started (%d).\n", srv.PID)

	err = srv.handleSignals(killMaster, srv.PID)
	if err != nil {
		log.Fatalf("[ERR] The server has shut down: %v\n", err)
	}

	log.Printf("[INFO] The server has shut down.\n")

	return gs, nil
}

/*
GracefulServeTLS starts a HTTPS server. It receives a http.Server server passed by user and an indicators killMaster.
It first create a listener (either a new one or a imported one). Then it starts a goroutine for the server to start
accepting connections. Any hooked signals will be handled in handleSignals(...). Certificate and key are compulsory
for starting a HTTPS server.
*/
func GracefulServeTLS(server *http.Server, killMaster bool, certFile, keyFile string) (err error) {

	srv := &GracefulServer{
		Addr:   server.Addr,
		Server: server,
	}

	err = srv.createListener()
	if err != nil {
		log.Fatalf("[ERR] Unable to create a listener: %v.\n", err)
	}

	go srv.Server.ServeTLS(srv.Listener, certFile, keyFile)

	server.Close()

	pid := syscall.Getpid()
	srv.PID = pid
	log.Printf("[INFO] The server has started (%d).\n", srv.PID)

	err = srv.handleSignals(killMaster, srv.PID)
	if err != nil {
		log.Fatalf("[ERR] The server has shut down: %v\n", err)
	}

	log.Printf("[INFO] The server has shut down.\n")

	return
}

/*
GracefulShutDown shuts down a server with a timeout. In most case, you will not need use it.
 */
func GracefulShutDown(server *http.Server) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAwaitTimeout)
	expired, _ := ctx.Deadline()
	log.Printf("[INFO] Digesting requests will be timed out at %v", expired.Format("2006-01-02 15:04:05"))
	defer cancel()
	return server.Shutdown(ctx)
}

/*
createListener creates a listener on a given address. If a descriptor file is found, it creates a listener out
of it (importListener); otherwise, it creates a new one (newListener).
 */
func (srv *GracefulServer) createListener() (err error) {
	// Check environment variables
	env := os.Getenv(DefaultEnvVar)
	if env != "" {
		// If it is not empty, try to create a listener using this information
		err = srv.importListener(env)
		if err == nil {
			log.Println("[INFO] Imported a listener from file.")
			return
		} else {
			log.Printf("[INFO] Unable to import a listener from file: %v. Trying to create a new one.", err)
		}
	}

	// If env is empty or unable to create a listener out of a descriptor file (e.g., file not found, broken),
	// create a new one.
	err = srv.newListener()
	if err != nil {
		return
	}
	log.Printf("[INFO] Created a new listener on %s.", srv.Addr)

	return
}

/*
importListener imports a listener from a descriptor file.
 */
func (srv *GracefulServer) importListener(env string) (err error) {
	var fl listenerDescriptor

	err = json.Unmarshal([]byte(env), &fl)
	if err != nil {
		return fmt.Errorf("unable to unmarsh [%s] environment variable", env)
	}

	if fl.Addr != srv.Addr  {
		return fmt.Errorf("unable to find listener on %s", srv.Addr)
	}

	f := os.NewFile(uintptr(fl.FD), fl.Name)
	if f == nil {
		return fmt.Errorf("unable to create listener file %s", fl.Name)
	}

	defer f.Close()

	srv.Listener, err = net.FileListener(f)
	if err != nil {
		return err
	}

	return nil
}

/*
newListener creates a brand new listener.
 */
func (srv *GracefulServer) newListener() (err error) {
	srv.Listener, err = net.Listen(DefaultNetwork, srv.Addr)
	if err != nil {
		return err
	}

	return nil
}

/*
handleSignals handles OS signals. It receives two arguments killMaster and mpid. killMaster controls the killing
behaviour after a child is forked:
	- if it is true, kill the master (using mpid) after a child is successfully forked
	- if it is false, keep the master alive.
 */
func (srv *GracefulServer) handleSignals(killMaster bool, mpid int) error {
	sigChan := make(chan os.Signal, 1024)
	sigHooks := []os.Signal{syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM}
	signal.Notify(sigChan, sigHooks...)

	for {
		sig := <-sigChan
		log.Printf("[INFO] Server (%d) received signal %q.\n", mpid, sig)
		switch sig {
		case syscall.SIGHUP:
			err := srv.forkChild(killMaster, mpid)
			if err != nil {
				log.Printf("[ERR] Unable to fork a child: %v.\n", err)
				continue
			}
			return srv.shutDown()
		case syscall.SIGUSR1,  syscall.SIGUSR2:
			err := srv.forkChild(killMaster, mpid)
			if err != nil {
				log.Printf("[ERR] Unable to fork a child: %v.\n", err)
				continue
			}
		case syscall.SIGINT, syscall.SIGTERM:
			return srv.shutDown()

		default:
			log.Printf("[INFO] The signal %q is not a hooked one, ignored!\n", sig)
		}
	}
}

/*
forkChild forks a child. If killMaster is true, the master who has forked the child will be killed (using mpid).
 */
func (srv *GracefulServer) forkChild(killMaster bool, mpid int) (error) {
	f, err := createListenerFile(srv.Listener)
	if err != nil {
		return err
	}
	defer f.Close()

	l := listenerDescriptor{
		Addr:	srv.Addr,
		FD:		3,
		Name:	f.Name(),
	}

	env, err := json.Marshal(l)
	if err != nil {
		return err
	}

	files := []*os.File{os.Stdin, os.Stdout, os.Stderr, f}

	environment := append(os.Environ(), fmt.Sprintf("%s=%s", DefaultEnvVar, string(env)))

	exec, err := os.Executable()
	if err != nil {
		return err
	}
	execDir := filepath.Dir(exec)

	p, err := os.StartProcess(exec, []string{exec}, &os.ProcAttr{
		Dir:   execDir,
		Env:   environment,
		Files: files,
		Sys:   &syscall.SysProcAttr{},
	})

	if err != nil {
		return err
	}

	log.Printf("[INFO] Forked child (%v).\n", p.Pid)

	if killMaster {
		err = syscall.Kill(mpid, syscall.SIGTERM)
		if err != nil {
			return errors.New(fmt.Sprintf("unable to kill the master (%d): %v", mpid, err))
		}
		log.Printf("[INFO] Master (%v) was killed.", mpid)
	} else {
		log.Printf("[INFO] Master (%v) is still alive.", mpid)
	}

	return nil
}

/*
shutDown shuts down a server. A context (expired in DefaultAwaitTimeout time) is created as a timeout to shut
down the server.
 */
func (srv *GracefulServer) shutDown() (err error){
	ctx, cancel := context.WithTimeout(context.Background(), DefaultAwaitTimeout)
	expired, _ := ctx.Deadline()
	log.Printf("[INFO] Digesting requests will be timed out at %v", expired.Format("2006-01-02 15:04:05"))
	defer cancel()
	return srv.Server.Shutdown(ctx)
}

/*
createListenerFile creates the listener file for a given listener based on the listener's type.
 */
func createListenerFile(l net.Listener) (*os.File, error) {
	switch t := l.(type) {
	case *net.TCPListener:
		return t.File()
	case *net.UnixListener:
		return t.File()
	}
	return nil, fmt.Errorf("unsupported listener: %T", l)
}

