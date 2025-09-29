package main

import (
	"game/src/networking"
	"flag"
	"net"
	"fmt"
)

func main() {
	port := flag.Int("port", 30000, "Port to run the server on")
	flag.Parse()

	ln, err := net.Listen("tcp", "0.0.0.0:" + fmt.Sprint(*port))
	if (err != nil) { panic(err) }
	fmt.Println("Server running on localhost:" + fmt.Sprint(*port))

	go networking.StartUpdateLoop()

	for {
		conn, err := ln.Accept()
		if err != nil { return }
		go networking.HandleClient(conn)
	}
}