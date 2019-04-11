# Simple HTTP Servers

There are two simple test servers:
- a simple HTTP server (see simpleHTTPServer.go), with port 6666
- a simple HTTPS server (see simpleHTTPSServer.gp), with port 7777

## Simple HTTP server

### Start The Server
```
$ go run test_servers/simpleHTTPServer.go
```

### Interact With The Server
There is a simple api called 'ping' which can be used to interact with the server. 
After you call the api, the server will sleep for a short while (2 seconds or less), 
then it responses with 'pong!'
```
$ curl http://localhost:6666/ping
pong!
```

## Simple HTTPS server

## Start The Server
```
$ go run test_servers/simpleHTTPSServer.go
```

### Interact With The Server
There is also an api called 'ping' which functions exactly the same way as 'api' in 
the simple HTTP server. Because our test server uses a self-signed certificate, you 
need disable the verification of the peer by using `curl -k` or `curl --insecure`:
```
$ curl -k https://localhost:7777/ping
pong!
```



