package main

import (
	"flag"
	"fmt"
	"net"
	"potato-bones/src/globals"
	"potato-bones/src/networking"
)

func parseFlags() {
	// Potatoes are 4th most grown crop, and there's 206 bones in a human body!
	globals.Port = flag.Int("port", 4206, "Port to run the server on")

	globals.Host = flag.String("host", "0.0.0.0", "Host address to bind to")
	globals.Tickrate = flag.Int("tickrate", 50, "Server update rate (in miliseconds)")
	globals.MaxEntities = flag.Int("max-entities", 255, "Max Entities per environment")
	globals.MaxClients = flag.Int("max-clients", 255, "Max Clients per environment")
	globals.MaxLobbies = flag.Int("max-lobbies", 255, "Max Lobbies")
	globals.MaxPacketSize = flag.Int("max-packet-size", 1024, "Max incoming packet size in bytes")
	globals.GameSpeed = flag.Float64("gamespeed", 1, "Game speed multiplier")
	globals.DebugShowOutgoing = flag.Bool("debug-outgoing", false, "Print outgoing packets")
	globals.DebugShowIncoming = flag.Bool("debug-incoming", false, "Print incoming packets")
	globals.DebugLobbyInfo = flag.Bool("debug-lobby", false, "Print lobby updates")
	globals.SessionLength = flag.Int("session-length", 1440, "How long before a session id expires in minutes")

	flag.Parse()
}

func HandleTCPServer() {
	ln, err := net.Listen("tcp", *globals.Host + ":" + fmt.Sprint(*globals.Port))
	if (err != nil) { panic(err) }
	fmt.Println("TCP Server running on " + *globals.Host + ":" + fmt.Sprint(*globals.Port))

	for {
		conn, err := ln.Accept()
		if err != nil { continue }
		go networking.HandleTCPClient(conn)
	}
}

func HandleUDPServer() {
	addr, err := net.ResolveUDPAddr("udp", *globals.Host + ":" + fmt.Sprint(*globals.Port))
	if err != nil { panic(err) }

	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil { panic(err) }
	defer udpConn.Close()
	networking.SetUDPConn(udpConn)

	fmt.Println("UDP Server running on " + *globals.Host + ":" + fmt.Sprint(*globals.Port))

	buffer := make([]byte, *globals.MaxPacketSize)
	for {
		length, clientAddr, err := udpConn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading:", err)
			continue
		}

		data := buffer[:length]
		go networking.HandleUDPPacket(clientAddr, data)
	}
}

func main() {
	parseFlags()

	networking.InitNetworking()
	networking.StartUpdateLoop(*globals.Tickrate)

	fmt.Println("Starting servers...")
	go HandleUDPServer()
	HandleTCPServer()
}