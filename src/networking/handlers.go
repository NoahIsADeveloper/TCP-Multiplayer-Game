package networking

import (
	"fmt"
	"net"
)

const (
	CS_JOIN = 0x00
	CS_MOVE = 0x01
	CS_PING = 0x02
	CS_PONG = 0x03
	CS_REQUEST_SYNC = 0x04
	CS_REQUEST_CLIENT_ID = 0x05
	CS_REQUEST_LOBBY_LIST = 0x06
	CS_CREATE_LOBBY = 0x07
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

var toUpdate = make(map[clientId]bool)

func scJoinAccept(conn net.Conn, clientId clientId) error {
	return sendPacket(conn, SC_JOIN_ACCEPT, encodeVarInt(int(clientId)))
}

func scJoinDeny(conn net.Conn, reason string) error {
	var data []byte
	appendString(&data, reason)
	return sendPacket(conn, SC_JOIN_DENY, data)
}

func scUpdatePlayers(lobby *Lobby) error {
	var data []byte
	var array []byte
	var arraySize int = 0
	players := lobby.players

	for clientId, player := range players {
		value, ok := toUpdate[clientId]
		if !ok || !value { continue }

		appendVarInt(&array, int(clientId))
		x, y := player.GetPosition()
		appendPosition(&array, x, y)
		arraySize++
		toUpdate[clientId] = false
	}

	appendVarInt(&data, arraySize)
	data = append(data, array...)

	return sendPacketToAll(SC_UPDATE_PLAYERS, data)
}

func getSyncData(lobby *Lobby) []byte {
	var data []byte
	players := lobby.players
	appendVarInt(&data, len(players))

	for clientId, player := range players {
		appendVarInt(&data, int(clientId))
		appendString(&data, player.Name)
		x, y := player.GetPosition()
		appendPosition(&data, x, y)
	}

	return data
}

func scSyncPlayers(conn net.Conn, clientId clientId) error {
	lobby, ok := JoinedLobbies[clientId]
	if !ok { return fmt.Errorf("couldn't find lobby client %d is connected to", clientId) }
	return sendPacket(conn, SC_SYNC_PLAYERS, getSyncData(lobby))
}

func scSyncAllPlayers(lobby *Lobby) []error {
	data := getSyncData(lobby)
	return lobby.SendPacketToAll(SC_SYNC_PLAYERS, data)
}

func csJoinRequest(conn net.Conn, clientId clientId, packetData []byte) error {
	lobby, ok := JoinedLobbies[clientId]
	if ok {
		return fmt.Errorf("cannot accept join request from client %d as it's already in a lobby", clientId)
	}

	if (true) {
		var offset int = 0
		name, err := readString(packetData, &offset)
		if err != nil { return err }
		lobby.AddPlayer(clientId, name, conn)
		toUpdate[clientId] = false

		err = scJoinAccept(conn, clientId)
		if err != nil { return err }
		errs := scSyncAllPlayers(lobby)
		if len(errs) > 0 { return errs[1] }

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

	lobby, ok := GetLobbyFromClient(clientId)
	if !ok { return fmt.Errorf("cannot find client %d in a lobby", clientId) }

	player, _, err := lobby.GetClientData(clientId)
	if err != nil { return err }
	plrX, plrY := player.GetPosition()
	if x == plrX && y == plrY { return nil}

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

func scLobbyList(conn net.Conn) error {
	var data []byte
	appendVarInt(&data, len(Lobbies))
	for _, lobby := range(Lobbies) {
		appendVarInt(&data, lobby.ID)
		appendString(&data, lobby.Name)
		appendVarInt(&data, int(lobby.Host))

		appendVarInt(&data, len(lobby.players))
		for _, player := range(lobby.players) {
			appendString(&data, player.Name)
		}
	}

	return sendPacket(conn, 0x08, data)
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
		return scSyncPlayers(conn, clientId)
	case CS_REQUEST_CLIENT_ID: // Request the client id
		return csRequestClientId(conn, clientId)
	case CS_REQUEST_LOBBY_LIST: // Request a list of lobbies
		return scLobbyList(conn)
	case CS_CREATE_LOBBY: // Request a lobby be created
		var offset int = 0
		lobbyName, err := readString(packetData, &offset)
		if err != nil { return err }
		CreateLobby(lobbyName, clientId)

		return nil
	default:
		return fmt.Errorf("received unknown packet id %d from client %d", packetID, clientId)
	}
}