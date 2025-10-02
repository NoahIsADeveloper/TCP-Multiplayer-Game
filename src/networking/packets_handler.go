package networking

import (
	"potato-bones/src/networking/datatypes"
	"potato-bones/src/utils"
)

const (
	// Handshake
	CS_PING = 0x02
	CS_PONG = 0x03

	// Request data
	CS_REQUEST_CLIENT_ID = 0x05
	CS_REQUEST_SESSION_ID = 0x0C
	CS_REQUEST_PLAYER_SYNC = 0x04
	CS_REQUEST_LOBBY_LIST = 0x06
	CS_REQUEST_LOBBY_SYNC = 0x0B

	// Lobby Controls
	CS_LOBBY_JOIN = 0x00
	CS_LOBBY_LEAVE = 0x0A
	CS_LOBBY_CREATE = 0x07
	CS_LOBBY_KICK = 0x08
	CS_LOBBY_PROMOTE = 0x09

	// Playing
	CS_MOVE = 0x01
)

const (
	// Handshake
	SC_PING = 0x02
	SC_PONG = 0x03

	// Lobby stuff
	SC_LOBBY_LIST = 0x08
	SC_LOBBY_KICK = 0x06
	SC_LOBBY_JOIN_ACCEPT = 0x00
	SC_LOBBY_JOIN_DENY = 0x01

	// Sync
	SC_SYNC_CLIENT_ID = 0x07
	SC_SYNC_SESSION_ID = 0x0A
	SC_SYNC_PLAYER = 0x05
	SC_SYNC_LOBBY = 0x09

	// Updates
	SC_UPDATE_PLAYERS = 0x04
)

func getSyncData(lobby *Lobby) []byte {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock()

	var data []byte

	players := lobby.players
	datatypes.AppendVarInt(&data, len(players))

	for clientId, player := range players {
		datatypes.AppendVarInt(&data, int(clientId))
		datatypes.AppendString(&data, player.GetName())
	}

	return data
}

func HandlePacket(sconn *utils.SafeConn, clientId clientID, packetId int, packetData []byte) error {
	switch packetId {
		// Handshake
	case CS_PING:
		return sconn.SendPacketTCP(SC_PONG, packetData)
	case CS_PONG:
		return nil // TODO timeout type shit

		// Requests
	case CS_REQUEST_PLAYER_SYNC:
		return scSyncPlayer(sconn, clientId)
	case CS_REQUEST_CLIENT_ID:
		return scSyncClientId(sconn, clientId)
	case CS_REQUEST_SESSION_ID:
		return scSyncSessionId(sconn, clientId)
	case CS_REQUEST_LOBBY_SYNC:
		return scSyncLobby(sconn, clientId)
	case CS_REQUEST_LOBBY_LIST:
		return scLobbyList(sconn)

		// Lobby controls
	case CS_LOBBY_JOIN:
		return csLobbyJoin(sconn, clientId, packetData)
	case CS_LOBBY_CREATE:
		return csLobbyCreate(sconn, clientId, packetData)

		// Playing
	case CS_MOVE:
		return csMove(clientId, packetData)
	}

	return nil
}