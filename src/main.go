package main

import (
	"net"
	"fmt"
	"game/src/networking"
)

const PORT = 30000

func main() {
	ln, err := net.Listen("tcp", "0.0.0.0:" + fmt.Sprint(PORT))
	if (err != nil) { panic(err) }
	fmt.Println("Server running on localhost:" + fmt.Sprint(PORT))

	for {
		conn, err := ln.Accept()
		if err != nil { return }
		go networking.HandleClient(conn)
	}
}