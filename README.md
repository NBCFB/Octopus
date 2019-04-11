# Octopus - Serve HTTP Server Gracefully

![](https://github.com/NBCFB/Octopus/blob/develop/octopus.jpeg)

Octopus is a tool that help serve HTTP server gracefully. User can:
- gracefully upgrade server (binary) without blocking incoming requests
- gracefully shutdown server so that incoming requests can be handled or time-outed.

## Credits
Octopus is greatly inspired by the article ["Gracefully Restarting a Go Program Without Downtime"](https://gravitational.com/blog/golang-ssh-bastion-graceful-restarts/) written by Russell Jones. A lot of credits should go to Russell. 

<!-- 
- Octopus is a TCP listener. You can use **Octopus** to:
- Create a fresh TCP listener
- Create a TCP listener from a ***os.File**
- Fork a child process from main process

Octopus receives `kill` signals to interrupt or terminate a running process. Here is the expected results with their 
corresponding signals:

| SIGNALS |  DESC | IS MAIN (PARENT) ALIVE | IS CHILD ALIVE |
| ----    | ----  | ----      | ----       |
|`kill -SIGUSR2 <pid>` | Fork a child | Y | Y |
|`kill -SIGNHUP <pid>` | Fork a child and kill the parent | N | Y |
|`kill -SIGNINT <pid>` | Terminate a process | N if it is the target PROC | N if it is the target PROC | 
|`kill -SIGNQUIT <pid>` | Terminate a process | N if it is the target PROC | N if it is the target PROC | 
|`kill -SIGNTERM <pid>` | Terminate a process | N if it is the target PROC | N if it is the target PROC | 

When sending a `kill` signal to a PROC, the PROC will not die immediately. It can still process incoming HTTP requests 
within the next 5 seconds. Then it is doomed. This 5 seconds simply buys sometime for completing the forking procedure. 

To summarise:
- `kill -SIGNUSR2` and `kill -SIGHUP` can be used to restart/reload a service gracefully.
- `kill -SIGNINT`, `kill -SIGNQUIT`, `kill -SIGNTERM` can be used to terminate a service gracefully. -->


