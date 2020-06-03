package main

import (
        "fmt"
        "net/http"
        "net/http/fcgi"
        "net"
)

type FastCGIServer struct{}

func (s FastCGIServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
        w.Write([]byte("This is a FastCGI example server.\n"))
}

func main() {
        fmt.Println("Starting server...")
        l, _ := net.Listen("unix", "/var/run/fcgiwrap.socket")
        b := new(FastCGIServer)
        fcgi.Serve(l, b)
}
