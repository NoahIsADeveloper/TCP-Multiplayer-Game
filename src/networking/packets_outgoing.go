package networking

import (
	"fmt"
	"potato-bones/src/networking/datatypes"
	"potato-bones/src/utils"
)

func scResetSequenceCount(sconn *utils.SafeConn) {
	sconn.SendPacketTCP(SC_RESET_SEQUENCE_COUNT, make([]byte, 0))
}

func scUpdatePlayers(lobby *Lobby, sequence int) error {
	var data []byte
	var array []byte
	var arraySize int = 0

	players := lobby.GetPlayers()
	for clientId, player := range(players) {
		if !player.DoUpdate() { continue }
		datatypes.AppendVarInt(&array, int(clientId))
		x, y, rotation := player.GetPosition()
		datatypes.AppendPosition(&array, x, y)
		datatypes.AppendRotation(&array, rotation)
		arraySize++
	}

	if arraySize == 0 { return nil }

	datatypes.AppendVarInt(&data, sequence)
	datatypes.AppendVarInt(&data, arraySize)
	data = append(data, array...)

	lobby.SendPacketToAllUDP(SC_UPDATE_PLAYERS, data)
	return nil
}

//
func scJoinAccept(sconn *utils.SafeConn, clientId clientID, lobby *Lobby) error {
	var data []byte
	datatypes.AppendVarInt(&data, int(clientId))
	datatypes.AppendVarInt(&data, int(lobby.GetID()))
	datatypes.AppendString(&data, lobby.GetName())

	return sconn.SendPacketTCP(SC_LOBBY_JOIN_ACCEPT, data)
}

func scJoinDeny(sconn *utils.SafeConn, reason string) error {
	var data []byte
	datatypes.AppendString(&data, reason)
	return sconn.SendPacketTCP(SC_LOBBY_JOIN_ACCEPT, data)
}

// Requests
func scSyncClientId(sconn *utils.SafeConn, clientId clientID) error {
	data := []byte{}
	datatypes.AppendVarInt(&data, int(clientId))
	return sconn.SendPacketTCP(SC_SYNC_CLIENT_ID, data)
}

func scSyncSessionId(sconn *utils.SafeConn, clientId clientID) error {
	data := []byte{}

	sessionId, ok := getSessionId(clientId)
	if !ok {
		return fmt.Errorf("session id not found for client %d", clientId)
	}
	datatypes.AppendString(&data, sessionId)

	return sconn.SendPacketTCP(SC_SYNC_SESSION_ID, data)
}

func scSyncPlayer(sconn *utils.SafeConn, clientId clientID) error {
	lobby, ok := GetLobbyFromClient(clientId)
	if !ok {
		return fmt.Errorf("scSyncPlayer cannot sync client %d as they are not in a lobby", clientId)
	}

	data := getPlayerSyncData(lobby)
	return sconn.SendPacketTCP(SC_SYNC_PLAYER, data)
}

func scSyncLobbyPlayers(lobby *Lobby) error {
	data := getPlayerSyncData(lobby)
	lobby.SendPacketToAllTCP(SC_SYNC_PLAYER, data)
	return nil
}

func scSyncLobby(sconn *utils.SafeConn, clientId clientID) error {
	lobby, ok := GetLobbyFromClient(clientId)

	if !ok {
		sconn.SendPacketTCP(SC_SYNC_LOBBY, []byte{0x00})
	}

	data := []byte{0x01}
	data = append(data, getLobbySyncData(lobby)...)
	return sconn.SendPacketTCP(SC_SYNC_LOBBY, data)
}

func scSyncEntireLobby(lobby *Lobby) error {
	data := []byte{0x01}
	data = append(data, getLobbySyncData(lobby)...)
	lobby.SendPacketToAllTCP(SC_SYNC_PLAYER, data)
	return nil
}

func scLobbyList(sconn *utils.SafeConn) error {
	globalLobbyMutex.RLock(); defer globalLobbyMutex.RUnlock()

	data := []byte{}

	datatypes.AppendVarInt(&data, len(lobbies))

	for _, lobby := range(lobbies) {
		datatypes.AppendVarInt(&data, int(lobby.GetID()))
		datatypes.AppendString(&data, lobby.GetName())
		datatypes.AppendVarInt(&data, int(lobby.GetHost()))

		players := lobby.GetPlayers()
		datatypes.AppendVarInt(&data, len(players))
		for _, player := range(players) {
			datatypes.AppendString(&data, player.GetName())
		}
	}

	return sconn.SendPacketTCP(SC_LOBBY_LIST, data)
}

func scKickPlayer(sconn *utils.SafeConn, reason string) error {
	var data []byte
	datatypes.AppendString(&data, reason)
	err := sconn.SendPacketTCP(SC_LOBBY_KICK, data)
	return err
}