package main

import (
	"potato-bones/src/globals"
	"potato-bones/src/networking"
	"flag"
	"net"
	"fmt"
)

func parseFlags() {
	globals.Port = flag.Int("port", 30000, "Port to run the server on")
	globals.Host = flag.String("host", "0.0.0.0", "Host address to bind to")
	globals.Tickrate = flag.Int("tickrate", 20, "Server TPS")
	globals.MaxEntities = flag.Int("max-entities", 255, "Max Entities per environment")
	globals.GameSpeed = flag.Float64("gamespeed", 1, "Game speed multiplier")

	globals.DebugShowOutgoing = flag.Bool("debug-outgoing", false, "Print outgoing packets")
	globals.DebugShowIncoming = flag.Bool("debug-incoming", false, "Print incoming packets")
	globals.DebugLobbyInfo = flag.Bool("debug-lobby", false, "Print lobby updates")

	flag.Parse()
}

func main() {
	parseFlags()

	ln, err := net.Listen("tcp", *globals.Host + ":" + fmt.Sprint(*globals.Port))
	if (err != nil) { panic(err) }
	fmt.Println("Server running on " + *globals.Host + ":" + fmt.Sprint(*globals.Port))

	networking.StartUpdateLoop(*globals.Tickrate)

	for {
		conn, err := ln.Accept()
		if err != nil { return }
		go networking.HandleClient(conn)
	}
}