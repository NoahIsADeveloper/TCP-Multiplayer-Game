package networking

import (
	"potato-bones/src/networking/datatypes"
	"potato-bones/src/utils"
)

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