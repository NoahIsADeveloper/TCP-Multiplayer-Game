package networking

import (
	"fmt"
	"net"
	"game/src/environment/entities"
)

const (
	CS_JOIN = 0x00
	CS_MOVE = 0x01
	CS_PING = 0x02
	CS_PONG = 0x03
	CS_REQUEST_SYNC = 0x04
)

const (
	SC_JOIN_ACCEPT = 0x00
	SC_JOIN_DENY = 0x01
	SC_PING = 0x02
	SC_PONG = 0x03
	SC_UPDATE_PLAYERS = 0x04
	SC_SYNC_PLAYERS = 0x05
)

var players = make(map[clientId]entities.Player)

func scJoinAccept(conn net.Conn, clientId clientId) error {
	return sendPacket(conn, SC_JOIN_ACCEPT, encodeVarInt(int(clientId)))
}

func scJoinDeny(conn net.Conn, reason string) error {
	var data []byte
	appendString(&data, reason)
	return sendPacket(conn, SC_JOIN_DENY, data)
}

func scUpdatePlayers() {
	var data []byte
	appendVarInt(&data, len(players))

	for clientId, player := range players {
		appendVarInt(&data, int(clientId))
		x, y := player.GetPosition()
		appendPosition(&data, x, y)
	}

	sendPacketToAll(SC_UPDATE_PLAYERS, data)
}

func getSyncData() []byte {
	var data []byte
	appendVarInt(&data, len(players))

	for clientId, player := range players {
		appendVarInt(&data, int(clientId))
		appendString(&data, player.Name)
		x, y := player.GetPosition()
		appendPosition(&data, x, y)
	}

	return data
}

func scSyncAllPlayers() error {
	return sendPacketToAll(SC_SYNC_PLAYERS, getSyncData())
}

func scSyncPlayers(conn net.Conn) error {
	return sendPacket(conn, SC_SYNC_PLAYERS, getSyncData())
}

func csJoinRequest(conn net.Conn, clientId clientId, packetData []byte) error {
	_, ok := players[clientId]
	if ok {
		return fmt.Errorf("cannot accept join request from client %d as it's already in game", clientId)
	}

	if (true) {
		var offset int = 0
		name, err := readString(packetData, &offset)
		if err != nil { return err }
		players[clientId] = *entities.CreatePlayer(name)

		scJoinAccept(conn, clientId)
		scSyncAllPlayers()

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
	sendPacket(conn, SC_PONG, data)

	return nil
}

func handlePacket(conn net.Conn, clientId clientId, packetID int, packetData []byte) error {
	switch packetID {
	case CS_JOIN: // Join (fields ignored for now)
		return csJoinRequest(conn, clientId, packetData)
	case CS_MOVE: // Move
		return csMove(clientId, packetData)
	case CS_PING: // Ping
		return csPing(conn, packetData)
	case CS_PONG: // Pong
		return nil
	case CS_REQUEST_SYNC: // Request a sync
		return scSyncPlayers(conn)
	default:
		return fmt.Errorf("received unknown packet id %d from client %d", packetID, clientId)
	}
}