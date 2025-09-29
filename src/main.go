package main

import (
	"flag"
	"net"
	"fmt"
	"game/src/networking"
)

func main() {
	port := flag.Int("port", 30000, "Port to run the server on")
	flag.Parse()

	ln, err := net.Listen("tcp", "0.0.0.0:" + fmt.Sprint(*port))
	if (err != nil) { panic(err) }
	fmt.Println("Server running on localhost:" + fmt.Sprint(*port))

	for {
		conn, err := ln.Accept()
		if err != nil { return }
		go networking.HandleClient(conn)
	}
}