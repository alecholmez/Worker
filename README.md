# Worker
A Go worker pool example using the dispatcher method

## Setup
To setup this project, clone and run the following:
```
make clean
make tools
make depend.install
```
## Running
To run the project:
```
make build
./worker
```
An HTTP server will listen on `0.0.0.0:8080` and react to `GET` requests at the `/` API route.

---
Credit for Idea/Code: [Marcio.IO](http://marcio.io/2015/07/handling-1-million-requests-per-minute-with-golang/)
