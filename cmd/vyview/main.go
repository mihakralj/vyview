package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/mihakralj/vyview/internal/demo"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address for the demo server")
	allowLAN := flag.Bool("allow-lan", false, "allow binding to non-loopback LAN/router addresses")
	flag.Parse()
	host, _, err := net.SplitHostPort(*addr)
	if err != nil {
		log.Fatalf("invalid listen address: %v", err)
	}
	if !*allowLAN {
		ip := net.ParseIP(host)
		if ip == nil || !ip.IsLoopback() {
			log.Fatal("demo server must listen on a loopback IP address (pass -allow-lan to bind to a LAN address)")
		}
	}
	handler, err := demo.Handler()
	if err != nil {
		log.Fatalf("initialize demo: %v", err)
	}
	server := &http.Server{Addr: *addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	log.Printf("Demo data only: http://%s", *addr)
	log.Fatal(server.ListenAndServe())
}
