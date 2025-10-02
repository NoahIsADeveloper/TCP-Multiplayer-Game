package networking

import (
	"fmt"
	"potato-bones/src/networking/datatypes"
	"potato-bones/src/utils"
)

func csLobbyJoin(sconn *utils.SafeConn, clientId clientID, packetData []byte) error {
	var offset int = 0

	username, err := datatypes.ReadString(packetData, &offset)
	if err != nil { return err }

	lobbyId, err := datatypes.ReadVarInt(packetData, &offset)
	if err != nil { return err }

	lobbyMutex.RLock()
	lobby, ok := lobbies[lobbyID(lobbyId)]
	lobbyMutex.RUnlock()
	if !ok {
		return scJoinDeny(sconn, "Requested lobby doesn't exist")
	}

	lobby.AddPlayer(clientId, username, sconn)
	return scJoinAccept(sconn, clientId, lobby)
}

func csLobbyCreate(sconn *utils.SafeConn, clientId clientID, packetData []byte) error {
	var offset int = 0

	username, err := datatypes.ReadString(packetData, &offset)
	if err != nil { return err }

	lobbyName, err := datatypes.ReadString(packetData, &offset)
	if err != nil { return err }

	lobby, err := CreateLobby(lobbyName, clientId)
	if err != nil { return err }

	lobby.AddPlayer(clientId, username, sconn)
	scJoinAccept(sconn, clientId, lobby)

	return nil
}

func csMove(clientId clientID, packetData []byte) error {
	var offset int = 0

	lobby, ok := GetLobbyFromClient(clientId)
	if !ok {
		return fmt.Errorf("client %d can not move as they're not in a lobby", clientId)
	}

	player, _, err := lobby.GetClientData(clientId)
	if err != nil { return err }

	sequence, err := datatypes.ReadVarInt(packetData, &offset)
	if err != nil { return err }

	x, y, err := datatypes.ReadPosition(packetData, &offset)
	if err != nil { return err }

	rotation, err := datatypes.ReadRotation(packetData, &offset)
	if err != nil { return err }

	player.Move(x, y, rotation, sequence)

	return nil
}