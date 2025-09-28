package networking

import (
	"net"
	"fmt"
)

type clientId uint8
var nextID clientId = 1
var freeIDs = []clientId{}

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
}

func HandleClient(conn net.Conn) {
	id := getID()

	defer conn.Close()
	defer releaseID(id)
	defer recover()

	fmt.Printf("Client %d connected from %s!\n", id, conn.RemoteAddr().String())

	for {
		packetID, packetData, err := readPacket(conn)
		if err != nil {
			fmt.Printf("Couldn't read packet from client %d, encountered error %s.\n", id, err)
			return
		}

		handlePacket(conn, id, packetID, packetData)
	}
}