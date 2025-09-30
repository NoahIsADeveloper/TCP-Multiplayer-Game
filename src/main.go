package main

import (
	"potato-bones/src/networking"
	"flag"
	"net"
	"fmt"
)

func main() {
	port := flag.Int("port", 30000, "Port to run the server on")
	host := flag.String("host", "0.0.0.0", "Host address to bind to")
	tickrate := flag.Int("tickrate", 20, "Server TPS")
	flag.Parse()

	ln, err := net.Listen("tcp", *host + ":" + fmt.Sprint(*port))
	if (err != nil) { panic(err) }
	fmt.Println("Server running on " + *host + ":" + fmt.Sprint(*port))

	networking.StartUpdateLoop(*tickrate)

	for {
		conn, err := ln.Accept()
		if err != nil { return }
		go networking.HandleClient(conn)
	}
}