package networking

import (
	"net"
	"fmt"
	"time"
)

type clientId uint8
var nextID clientId = 0
var freeIDs = []clientId{}
var connections = make(map[clientId]net.Conn)

func getID() clientId {
	if (len(freeIDs) > 0) {
		id := freeIDs[len(freeIDs)-1]
		freeIDs = freeIDs[:len(freeIDs)-1]
		return id
	}
	nextID++
	return nextID - 1
}

func releaseID(id clientId) {
	freeIDs = append(freeIDs, id)
	delete(players, id)
	delete(connections, id)
}

func updatePlayersLoop(tickrate int) {
	ticker := time.NewTicker(time.Second / time.Duration(tickrate))
	defer ticker.Stop()

	for range ticker.C {
		scUpdatePlayers()
	}
}

func StartUpdateLoop(tickrate int) {
	go updatePlayersLoop(tickrate)
}

func HandleClient(conn net.Conn) {
	id := getID()
	connections[id] = conn

	defer fmt.Printf("Client %d connection closed.\n", id)
	defer conn.Close()
	defer releaseID(id)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered from panic for client %d: %v\n", id, r)
		}
	}()

	fmt.Printf("Client %d connected from %s.\n", id, conn.RemoteAddr().String())

	for {
		packetID, packetData, err := readPacket(conn)
		if err != nil {
			fmt.Printf("Couldn't read packet from client %d, encountered error %v.\n", id, err)
			return
		}

		err = handlePacket(conn, id, packetID, packetData)
		if err != nil {
			fmt.Printf("Couldn't handle packet from client %d id %d, encountered error %v.\n", id, packetID, err)
			return
		}
	}
}