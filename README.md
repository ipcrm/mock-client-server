# mock-client-server

This project is used to create a simple mock client/server application.

The server listens on port *:8080.  The client will send repeated messages to the server to stay connected. For control
over connections, duration, and quantity - see client help.

## Pre-reqs

```
Go >= 1.16
```

## building

```
make prepare
make ci
make build
```


## run

Client:

```
./bin/client-<os>-<arch> -h
```

Server:

```
./bin/server-<os>-<arch> -h
```
