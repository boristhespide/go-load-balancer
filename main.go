package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

type simpleServer struct {
	addr  string
	proxy *httputil.ReverseProxy
}

type Server interface {
	Addres() string
	IsAlive() bool
	Serve(w http.ResponseWriter, r *http.Request)
}

type LoadBalancer struct {
	port           string
	roadRobinCount int
	servers        []Server
}

func newSimpleServer(addr string) *simpleServer {
	serverUrl, err := url.Parse(addr)
	handleError(err)

	return &simpleServer{
		addr:  addr,
		proxy: httputil.NewSingleHostReverseProxy(serverUrl),
	}
}

func NewLoadBalancer(port string, servers []Server) *LoadBalancer {
	return &LoadBalancer{
		port:           port,
		roadRobinCount: 0,
		servers:        servers,
	}
}

func handleError(err error) {
	if err != nil {
		fmt.Printf("ErrorL %v\n", err)
		os.Exit(1)
	}
}

func (s *simpleServer) Addres() string { return s.addr }

func (s *simpleServer) IsAlive() bool { return true } // for demo only

func (s *simpleServer) Serve(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func (lb *LoadBalancer) getNextAvailableServer() Server {
	server := lb.servers[lb.roadRobinCount%len(lb.servers)]
	for !server.IsAlive() {
		lb.roadRobinCount++
		server = lb.servers[lb.roadRobinCount%len(lb.servers)]
	}
	lb.roadRobinCount++ // for demo only
	return server
}

func (lb *LoadBalancer) serveProxy(w http.ResponseWriter, r *http.Request) {
	targetServer := lb.getNextAvailableServer()
	fmt.Printf("Forwarding request to adress %v\n", targetServer.Addres())
	targetServer.Serve(w, r)
}

func main() {
	servers := []Server{
		newSimpleServer("https://facebook.com"),
		newSimpleServer("https://google.com"),
		newSimpleServer("https://github.com"),
	}
	lb := NewLoadBalancer("8000", servers)
	handleRedirect := func(w http.ResponseWriter, r *http.Request) {
		lb.serveProxy(w, r)
	}
	http.HandleFunc("/", handleRedirect)
	fmt.Printf("serving requests at port %v\n", lb.port)
	http.ListenAndServe(":"+lb.port, nil)
}
