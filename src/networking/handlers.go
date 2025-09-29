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
	CS_REQUEST_CLIENT_ID = 0x05
)

const (
	SC_JOIN_ACCEPT = 0x00
	SC_JOIN_DENY = 0x01
	SC_PING = 0x02
	SC_PONG = 0x03
	SC_UPDATE_PLAYERS = 0x04
	SC_SYNC_PLAYERS = 0x05
	SC_KICK_PLAYER = 0x06
	SC_CLIENT_ID = 0x07
)

var players = make(map[clientId]*entities.Player)
var toUpdate = make(map[clientId]bool)

func scJoinAccept(conn net.Conn, clientId clientId) error {
	return sendPacket(conn, SC_JOIN_ACCEPT, encodeVarInt(int(clientId)))
}

func scJoinDeny(conn net.Conn, reason string) error {
	var data []byte
	appendString(&data, reason)
	return sendPacket(conn, SC_JOIN_DENY, data)
}

func scUpdatePlayers() error {
	var data []byte
	appendVarInt(&data, len(players))

	// TODO: Only send recently moved players
	for clientId, player := range players {
		value, ok := toUpdate[clientId]
		if !ok || !value { continue }

		appendVarInt(&data, int(clientId))
		x, y := player.GetPosition()
		appendPosition(&data, x, y)

		toUpdate[clientId] = false
	}

	return sendPacketToAll(SC_UPDATE_PLAYERS, data)
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
		players[clientId] = entities.CreatePlayer(name)
		toUpdate[clientId] = false

		err = scJoinAccept(conn, clientId)
		if err != nil { return err }
		err = scSyncAllPlayers()
		if err != nil { return err }

		fmt.Printf("Client %d joined the game\n", clientId)
	} else {
		return scJoinDeny(conn, "Join request denied")
	}

	return nil
}

func csMove(clientId clientId, packetData []byte) error {
	var offset int = 0
	x, y, err := readPosition(packetData, &offset)
	if err != nil { return err }

	player, ok := players[clientId]
	if !ok { return fmt.Errorf("couldn't move client %d: no player object found", clientId)}
	if x == player.X && y == player.Y { return nil}
	if !player.InRange(x, y, 100) { // More than 100 units in a single packet
		scKickPlayer(connections[clientId], "You were moving too fast!")
		return fmt.Errorf("client %d was moving too fast", clientId)
	}

	player.Move(x, y)
	toUpdate[clientId] = true

	return nil
}

func scKickPlayer(conn net.Conn, reason string) error {
	var data []byte
	appendString(&data, reason)
	err := sendPacket(conn, SC_KICK_PLAYER, data)
	conn.Close()
	return err
}

func scPong(conn net.Conn, data []byte) error {
	return sendPacket(conn, SC_PONG, data)
}

func csRequestClientId(conn net.Conn, clientId clientId) error {
	return sendPacket(conn, SC_CLIENT_ID, encodeVarInt(int(clientId)))
}

func handlePacket(conn net.Conn, clientId clientId, packetID int, packetData []byte) error {
	switch packetID {
	case CS_JOIN: // Join (fields ignored for now)
		return csJoinRequest(conn, clientId, packetData)
	case CS_MOVE: // Move
		return csMove(clientId, packetData)
	case CS_PING: // Ping
		return scPong(conn, packetData)
	case CS_PONG: // Pong
		return nil
	case CS_REQUEST_SYNC: // Request a sync
		return scSyncPlayers(conn)
	case CS_REQUEST_CLIENT_ID: // Request the client id
		return csRequestClientId(conn, clientId)
	default:
		return fmt.Errorf("received unknown packet id %d from client %d", packetID, clientId)
	}
}