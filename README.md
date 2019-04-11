# Octopus - Serve HTTP Server Gracefully

[![CircleCI](https://circleci.com/gh/NBCFB/Octopus/tree/develop.svg?style=svg)](https://circleci.com/gh/NBCFB/Octopus/tree/develop)

![](https://github.com/NBCFB/Octopus/blob/develop/octopus.jpeg)

Octopus is a tool that help serve HTTP server gracefully. User can:
- gracefully upgrade server (binary) without blocking incoming requests
- gracefully shutdown server so that incoming requests can be handled or time-outed.

## Credits
Octopus is greatly inspired by the article ["Gracefully Restarting a Go Program Without Downtime"](https://gravitational.com/blog/golang-ssh-bastion-graceful-restarts/) written by Russell Jones. A lot of credits should go to Russell. 

## Example - HTTP Server

### Create A Graceful HTTP Server

You simply call GracefulServe(...) by passing your http.Server. That's it. Here is an example.
```
package main

import (
	"flag"
	"fmt"
	"github.com/NBCFB/Octopus"
	"github.com/go-chi/chi"
	"net/http"
	"time"
)

func ping (w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong\n"))
}

func main() {
	var host string
	var port int

	// Handle command line flag
	flag.StringVar(&host, "h", "localhost", "specify host")
	flag.IntVar(&port, "p", 8080, "specify port")
	flag.Parse()

	// Make up addr with default settings
	addr := fmt.Sprintf("%s:%d", host, port)

	// Set up router
	r := chi.NewRouter()
	r.Get("/ping", ping)

	s := &http.Server{
		Addr: addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}

	// Start server
	Octopus.GracefulServe(s, false)
}
```

Now, we can start the server using **go run** from command line:
```
$ go run test_servers/simpleHTTPServer.go -h 172.18.1.239 -p 8080 &
[1] 52215
blackstar:Octopus xiali$ 2019/04/11 16:09:12 [INFO] Created a new listener on 172.18.1.239:8080.
2019/04/11 16:09:12 [INFO] The server has started (52233).

```
Now we can see that the server has started, listening on address **172.18.1.239:8080**, where the pid is **52233**. We can check if the server is running properly.
```
$ curl http://172.18.1.239:8080/ping
pong
```

### Upgrade(Restart) The Server
Assume we have updated (using **go build**) the server binary. The change is shown as below:
```
func ping (w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong pong\n"))
}
```

We do not need perform a 'stop-and-start' process. We can simply send a SIGUSR1 or SIGUSR2 to gracefully restart the server without blocking incoming requests:
```
$ kill -SIGUSR1 52233
blackstar:Octopus xiali$ 2019/04/11 16:18:11 [INFO] Server (52233) received signal "user defined signal 1".
2019/04/11 16:18:11 [INFO] Forked child (55344).
2019/04/11 16:18:11 [INFO] Master (52233) is still alive.
2019/04/11 16:18:11 [INFO] Unable to import a listener from file: unable to find listener on localhost:8080. Trying to create a new one.
2019/04/11 16:18:11 [INFO] Created a new listener on localhost:8080.
2019/04/11 16:18:11 [INFO] The server has started (55344).
```
Now we can see that a child was forked whose pid is **55344**. Since we have set **killMaster(the second arg in Octopus.GracefulServe(...)** to false, we keep the master alive. At this point, the incoming requests will be handled randomly by either the master or the child, as shown below:
```
$ curl http://172.18.1.239:8080/ping
pong pong
$ curl http://172.18.1.239:8080/ping
pong
```

You can set killMaster to true to kill the master after a child is successfully forked. 

Now, we can kill the master so that all the incoming requests will be handled based on the updated server binary.
```
$ kill -SIGTERM 52233
blackstar:Octopus xiali$ 2019/04/11 16:24:54 [INFO] Server (52233) received signal "terminated".
2019/04/11 16:24:54 [INFO] Digesting requests will be timed out at 2019-04-11 16:24:59
2019/04/11 16:24:54 [INFO] The server has shut down.
```

The above output shows that the arrived requests will be still processed till **2019-04-11 16:24:59**. Now, let's check if the server is updated.
```
$ curl http://172.18.1.239:8080/ping
pong pong
```
We can see that the server has been upgraded successfully without blocking any incoming request.

## Signals
Octopus listens only preset hooked signals including **syscall.SIGHUP, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT and syscall.SIGTERM**. The corresponding behaviors are shown as below:

| SIGNALS |  BEHAVIOR |
| ----    | ----  |
| kill -SIGNHUP <pid> | Fork --> shut down. |
| kill -SIGUSR1 <pid>, kill -SIGUSR2 <pid> | Fork |
| kill -SIGNINT <pid>, kill -SIGTERM <pid> | Shut down |
  
## Example Server Codes
There are two example servers in **/test_servers** folder:
- [A simple HTTP server](https://github.com/NBCFB/Octopus/blob/develop/test_servers/simpleHTTPServer.go)
- [A simple HTTPS server](https://github.com/NBCFB/Octopus/blob/develop/test_servers/simpleHTTPSServer.go)


