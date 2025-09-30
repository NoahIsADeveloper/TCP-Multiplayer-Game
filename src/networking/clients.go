package networking

import (
	"net"
	"fmt"
	"time"
)

type clientId uint8
var nextClientID clientId = 0
var freeClientIDs = []clientId{}
var connections = make(map[clientId]net.Conn)

func getID() clientId {
	if (len(freeClientIDs) > 0) {
		id := freeClientIDs[len(freeClientIDs)-1]
		freeClientIDs = freeClientIDs[:len(freeClientIDs)-1]
		return id
	}
	nextClientID++
	return nextClientID - 1
}

func releaseID(id clientId) {
	freeClientIDs = append(freeClientIDs, id)
	GetLobbyFromClient(id).RemovePlayer(id)
	delete(connections, id)
	delete(toUpdate, id)
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