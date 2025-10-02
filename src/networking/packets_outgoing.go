package networking

import (
	"fmt"
	"potato-bones/src/networking/datatypes"
	"potato-bones/src/utils"
)

func scUpdatePlayers(lobby *Lobby) error {
	var data []byte
	var array []byte
	var arraySize int = 0

	lobby.mutex.RLock()
	for clientId, player := range(lobby.players) {
		// check if player actually needs update
		datatypes.AppendVarInt(&array, int(clientId))
		x, y, rotation := player.GetPosition()
		datatypes.AppendPosition(&array, x, y)
		datatypes.AppendRotation(&array, rotation)
		arraySize++
	}
	lobby.mutex.RUnlock()

	datatypes.AppendVarInt(&data, arraySize)
	data = append(data, array...)

	lobby.SendPacketToAllUDP(SC_UPDATE_PLAYERS, data)
	return nil
}

//
func scJoinAccept(sconn *utils.SafeConn, clientId clientID, lobby *Lobby) error {
	var data []byte
	lobby.mutex.RLock()
	datatypes.AppendVarInt(&data, int(clientId))
	datatypes.AppendVarInt(&data, int(lobby.id))
	datatypes.AppendString(&data, lobby.name)
	lobby.mutex.RUnlock()

	return sconn.SendPacketTCP(SC_LOBBY_JOIN_ACCEPT, data)
}

func scJoinDeny(sconn *utils.SafeConn, reason string) error {
	var data []byte
	datatypes.AppendString(&data, reason)
	return sconn.SendPacketTCP(SC_LOBBY_JOIN_ACCEPT, data)
}

// Requests
func scSyncPlayer(sconn *utils.SafeConn, clientId clientID) error {
	lobby, ok := GetLobbyFromClient(clientId)
	if !ok {
		return fmt.Errorf("scSyncPlayer cannot sync client %d as they are not in a lobby", clientId)
	}

	data := getSyncData(lobby)
	return sconn.SendPacketTCP(SC_SYNC_PLAYER, data)
}

func scSyncClientId(sconn *utils.SafeConn, clientId clientID) error {
	data := []byte{}
	datatypes.AppendVarInt(&data, int(clientId))
	return sconn.SendPacketTCP(SC_SYNC_CLIENT_ID, data)
}

func scSyncSessionId(sconn *utils.SafeConn, clientId clientID) error {
	data := []byte{}

	clientMutex.RLock()
	datatypes.AppendString(&data, sessionIdFromClientId[clientId])
	clientMutex.RUnlock()

	return sconn.SendPacketTCP(SC_SYNC_SESSION_ID, data)
}

func scSyncLobby(sconn *utils.SafeConn, clientId clientID) error {
	lobby, ok := GetLobbyFromClient(clientId)
	if !ok {
		return fmt.Errorf("scSyncLobby cannot sync client %d as they are not in a lobby", clientId)
	}

	data := getSyncData(lobby)
	return sconn.SendPacketTCP(SC_SYNC_PLAYER, data)
}

func scSyncEntireLobby(lobby *Lobby) error {
	data := getSyncData(lobby)
	lobby.SendPacketToAllTCP(SC_SYNC_PLAYER, data)

	return nil
}


func scLobbyList(sconn *utils.SafeConn) error {
	lobbyMutex.RLock();

	data := []byte{}

	datatypes.AppendVarInt(&data, len(lobbies))
	for _, lobby := range(lobbies) {
		lobby.mutex.RLock()

		datatypes.AppendVarInt(&data, int(lobby.id))
		datatypes.AppendString(&data, lobby.name)
		datatypes.AppendVarInt(&data, int(lobby.host))

		datatypes.AppendVarInt(&data, len(lobby.players))
		for _, player := range(lobby.players) {
			datatypes.AppendString(&data, player.GetName())
		}

		lobby.mutex.RUnlock()
	}

	lobbyMutex.RUnlock();
	return sconn.SendPacketTCP(SC_LOBBY_LIST, data)
}

func scKickPlayer(sconn *utils.SafeConn, reason string) error {
	var data []byte
	datatypes.AppendString(&data, reason)
	err := sconn.SendPacketTCP(SC_LOBBY_KICK, data)
	return err
}