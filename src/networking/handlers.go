package networking

import (
	"fmt"
	"net"
	"game/src/environment/entities"
)

var players = make(map[clientId]entities.Player)

func scJoinAccept(conn net.Conn) {
	sendPacket(conn, 0x00, []byte{0x00})
}

func scJoinDeny(conn net.Conn, reason string) {
	var data []byte
	appendString(&data, reason)
	sendPacket(conn, 0x01, data)
}

func scUpdatePlayers() {
	var data []byte
	appendVarInt(&data, len(players))

	for _, player := range players {
		appendString(&data, player.Name)
		x, y := player.GetPosition()
		appendPosition(&data, x, y)
		fmt.Println(player.Name, x, y)
	}

	for _, conn := range connections {
		sendPacket(conn, 0x04, data)
	}
}

func csJoinRequest(conn net.Conn, clientId clientId, packetData []byte) error {
	if (true) {
		var offset int = 0
		name, err := readString(packetData, &offset)
		if err != nil { return err }
		scJoinAccept(conn)

		players[clientId] = *entities.CreatePlayer(name)
		fmt.Printf("Client %d joined the game\n", clientId)
	} else {
		scJoinDeny(conn, "Join request denied")
	}

	return nil
}

func csMove(clientId clientId, packetData []byte) error {
	var offset int = 0
	x, y, err := readPosition(packetData, &offset)
	if err != nil { return err }
	
	player, ok := players[clientId]
	if !ok { return fmt.Errorf("couldn't move client %d: no player object found", clientId)}
	player.Move(x, y)

	return nil
}

func csPing(conn net.Conn, data []byte) error {
	sendPacket(conn, 0x03, data)

	return nil
}

func handlePacket(conn net.Conn, clientId clientId, packetID int, packetData []byte) error {
	switch packetID {
	case 0x00: // Join (fields ignored for now)
		return csJoinRequest(conn, clientId, packetData)
	case 0x01: // Move
		return csMove(clientId, packetData)
	case 0x02: // Ping
		return csPing(conn, packetData)
	default:
		return fmt.Errorf("received unknown packet id %d from client %d", packetID, clientId)
	}
}