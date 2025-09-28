package networking

import (
	"fmt"
	"net"
)

func scJoinAccept(conn net.Conn) {
	sendPacket(conn, 0x00, []byte{0x00})
}

func scJoinDeny(conn net.Conn, reason string) {
	var data []byte
	appendString(&data, reason)
	sendPacket(conn, 0x01, data)
}

func handlePacket(conn net.Conn, clientId clientId, packetID int, data []byte) {

	switch packetID {
	case 0x00: // Join (fields ignored for now)
		if (true) {
			fmt.Printf("Client %d joined the game\n", clientId)
			scJoinAccept(conn)
		} else {
			scJoinDeny(conn, "Join request denied")
		}
	default:
		fmt.Printf("Received unknown packet id %d from client %d\n", packetID, clientId)
	}
}