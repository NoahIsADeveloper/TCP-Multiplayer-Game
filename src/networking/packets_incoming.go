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
	_, ok := joinedLobbies[clientId]
	if ok {
		lobbyMutex.RUnlock()
		return scJoinDeny(sconn, "You are already in a lobby.")
	}
	lobby, ok := lobbies[lobbyID(lobbyId)]
	lobbyMutex.RUnlock()
	if !ok {
		return scJoinDeny(sconn, "Requested lobby doesn't exist")
	}

	err = lobby.AddPlayer(clientId, username, sconn)
	if err != nil { return err }

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

func csLobbyKick(host clientID, packetData []byte) error {
	lobby, ok := GetLobbyFromClient(host)
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock()

	// TODO: Send client message on fail
	if !ok { return nil }
	if lobby.host != host { return nil }

	var offset int = 0
	victimId, err := datatypes.ReadVarInt(packetData, &offset)
	if err != nil { return err }

	toKick := clientID(victimId)
	if !lobby.HasClient(toKick) { return nil }

	reason, err := datatypes.ReadString(packetData, &offset)
	if err != nil { return err }

	scKickPlayer(connectionsFromClientId[toKick], reason)

	lobby.RemovePlayer(toKick)

	return nil
}

func csLobbyPromote(host clientID, packetData []byte) error {
	lobby, ok := GetLobbyFromClient(host)
	lobby.mutex.Lock(); defer lobby.mutex.Unlock()

	// TODO: Send client message on fail
	if !ok { return nil }
	if lobby.host != host { return nil }

	var offset int = 0
	victimId, err := datatypes.ReadVarInt(packetData, &offset)
	if err != nil { return err }
	toPromote := clientID(victimId)
	lobby.mutex.Unlock()
	if !lobby.HasClient(toPromote) { lobby.mutex.Lock(); return nil }
	lobby.mutex.Lock()
	lobby.host = toPromote

	lobby.mutex.Unlock()
	scSyncEntireLobby(lobby)
	lobby.mutex.Lock()

	return nil
}

func csLeaveLobby(clientId clientID) error {
	lobby, ok := GetLobbyFromClient(clientId)

	if !ok { return nil }
	if !lobby.HasClient(clientId) { return nil }

	lobby.RemovePlayer(clientId)

	return nil
}