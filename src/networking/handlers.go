package networking

import (
	"fmt"
	"net"
	"potato-bones/src/globals"
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
	CS_KICK_PLAYER = 0x08
	CS_CHANGE_HOST = 0x09
	CS_LEAVE_LOBBY = 0x0A
	CS_LOBBY_INFO = 0x0B
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
	SC_LOBBY_LIST = 0x08
	SC_LOBBY_INFO = 0x09
)

var toUpdate = make(map[clientId]bool)

func scJoinAccept(conn net.Conn, clientId clientId, lobby *Lobby) error {
	var data []byte
	appendVarInt(&data, int(clientId))
	appendString(&data, lobby.Name)
	return sendPacket(conn, SC_JOIN_ACCEPT, data)
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

	errs := lobby.SendPacketToAll(SC_UPDATE_PLAYERS, data)
	if len(errs) > 0 { return errs[0] }

	return nil
}

func getSyncData(lobby *Lobby) []byte {
	var data []byte

	appendVarInt(&data, int(lobby.Host))
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
	_, ok := JoinedLobbies[clientId]
	if ok {
		return scJoinDeny(conn, "Cannot join as you're already in a lobby.")
	}

	if (true) {
		var offset int = 0
		name, err := readString(packetData, &offset)
		if err != nil { return err }

		lobbyId, err:= readVarInt(packetData, &offset)
		if err != nil { return err }

		lobby, ok := Lobbies[lobbyId]
		if !ok {
			return scJoinDeny(conn, "Cannot join lobby as it does not exist")
		}

		err = lobby.AddPlayer(clientId, name, conn)
		if err != nil { return err}

		toUpdate[clientId] = false

		err = scJoinAccept(conn, clientId, lobby)
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

	lobby, ok := GetLobbyFromClient(clientId)
	if !ok { return fmt.Errorf("cannot find client %d in a lobby", clientId) }

	player, _, err := lobby.GetClientData(clientId)
	if err != nil { return err }
	plrX, plrY := player.GetPosition()
	if x == plrX && y == plrY { return nil }

	player.Move(x, y)
	toUpdate[clientId] = true

	return nil
}

func scKickPlayer(conn net.Conn, reason string) error {
	var data []byte
	appendString(&data, reason)
	err := sendPacket(conn, SC_KICK_PLAYER, data)
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

	return sendPacket(conn, SC_LOBBY_LIST, data)
}

func csCreateLobby(conn net.Conn, clientId clientId, packetData []byte) error {
	var offset int = 0

	// TODO: This sequence can probably be it's own fuction (same with csJoinRequest)
	username, err := readString(packetData, &offset)
	if err != nil { return err }

	lobbyName, err := readString(packetData, &offset)
	if err != nil { return err }

	lobby := CreateLobby(lobbyName, clientId)
	err = lobby.AddPlayer(clientId, username, conn)
	if err != nil { return err }

	err = scJoinAccept(conn, clientId, lobby)
	if err != nil { return err }

	return nil
}

func csKickPlayer(host clientId, packetData []byte) error {
	lobby, ok := GetLobbyFromClient(host)

	// TODO: Send client message on fail
	if !ok { return nil }
	if lobby.Host != host { return nil }

	var offset int = 0
	victimId, err := readVarInt(packetData, &offset)
	if err != nil { return err }

	toKick := clientId(victimId)
	if !lobby.HasClient(toKick) { return nil }

	reason, err := readString(packetData, &offset)
	if err != nil { return err }

	scKickPlayer(connections[toKick], reason)
	lobby.RemovePlayer(toKick)

	return nil
}

func csChangeHost(host clientId, packetData []byte) error {
	lobby, ok := GetLobbyFromClient(host)

	// TODO: Send client message on fail
	if !ok { return nil }
	if lobby.Host != host { return nil }

	var offset int = 0
	victimId, err := readVarInt(packetData, &offset)
	if err != nil { return err }
	toPromote := clientId(victimId)
	if !lobby.HasClient(toPromote) { return nil }

	lobby.Host = toPromote
	return scLobbyInfoToAll(lobby)
}

func csLeaveLobby(clientId clientId) error {
	lobby, ok := GetLobbyFromClient(clientId)

	// TODO: Send client message on fail
	if !ok { return nil }
	if !lobby.HasClient(clientId) { return nil }

	lobby.RemovePlayer(clientId)

	return nil
}

func getLobbyData(lobby *Lobby) []byte {
	data := []byte{}
	appendVarInt(&data, lobby.ID)
	appendString(&data, lobby.Name)
	appendVarInt(&data, int(lobby.Host))

	return data
}

func scLobbyInfo(conn net.Conn, clientId clientId) error {
	lobby, ok := GetLobbyFromClient(clientId)
	if !ok {
		sendPacket(conn, SC_LOBBY_INFO, []byte{0x00})
		return nil
	}

	data := []byte{0x01}
	data = append(data, getLobbyData(lobby)...)

	return sendPacket(conn, SC_LOBBY_INFO, data)
}

func scLobbyInfoToAll(lobby *Lobby) error {
	data := []byte{0x01}
	data = append(data, getLobbyData(lobby)...)

	errs := lobby.SendPacketToAll(SC_LOBBY_INFO, data)
	if len(errs) > 0 { return errs[0] }

	return nil
}

func handlePacket(conn net.Conn, clientId clientId, packetID int, packetData []byte) error {
	if *globals.DebugShowIncoming {
		fmt.Printf("[DEBUG] Incomming packet ID %d\n", packetID)
	}

	switch packetID {
	case CS_JOIN: // Join (fields ignored for now)
		return csJoinRequest(conn, clientId, packetData)
	case CS_MOVE: // Move
		return csMove(clientId, packetData)
	case CS_PING: // Incomming ping
		return scPong(conn, packetData)
	case CS_PONG: // Outgoing ping response packet
		return nil
	case CS_REQUEST_SYNC: // Request a sync
		return scSyncPlayers(conn, clientId)
	case CS_REQUEST_CLIENT_ID: // Request the client id
		return csRequestClientId(conn, clientId)
	case CS_REQUEST_LOBBY_LIST: // Request a list of lobbies
		return scLobbyList(conn)
	case CS_CREATE_LOBBY: // Request a lobby be created
		return csCreateLobby(conn, clientId, packetData)
	case CS_KICK_PLAYER: // Kick a player
		return csKickPlayer(clientId, packetData)
	case CS_CHANGE_HOST: // Change the host
		return csChangeHost(clientId, packetData)
	case CS_LEAVE_LOBBY: // Leave the lobby
		return csLeaveLobby(clientId)
	case CS_LOBBY_INFO: // Request lobby info
		return scLobbyInfo(conn, clientId)
	default:
		return fmt.Errorf("received unknown packet id %d from client %d", packetID, clientId)
	}
}