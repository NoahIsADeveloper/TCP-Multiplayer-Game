package networking

import (
	"fmt"
	"potato-bones/src/networking/datatypes"
	"potato-bones/src/utils"
)

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

func scSyncLobby(sconn *utils.SafeConn, clientId clientID) error {
	lobby, ok := GetLobbyFromClient(clientId)
	if !ok {
		return fmt.Errorf("scSyncLobby cannot sync client %d as they are not in a lobby", clientId)
	}

	data := getSyncData(lobby)
	return sconn.SendPacketTCP(SC_SYNC_PLAYER, data)
}

func scLobbyList(sconn *utils.SafeConn) error {
	lobbyMutex.RLock(); defer lobbyMutex.Unlock();

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

	return sconn.SendPacketTCP(SC_LOBBY_LIST, data)
}