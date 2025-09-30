package networking

import (
	"potato-bones/src/environment/entities"
	"fmt"
	"net"
)

var Lobbies = make(map[int]*Lobby)
var JoinedLobbies = make(map[clientId]*Lobby)

type Lobby struct{
	Name string
	Host clientId
	ID int

	players map[clientId]*entities.Player
	connections map[clientId]net.Conn
}

func (lobby *Lobby) Rename(name string) {
	lobby.Name = name
}

func (lobby *Lobby) RemovePlayer(clientId clientId) {
	delete(lobby.players, clientId)
	delete(lobby.connections, clientId)
	delete(JoinedLobbies, clientId)
}

func (lobby *Lobby) AddPlayer(clientId clientId, name string, conn net.Conn) {
	lobby.players[clientId] = entities.CreatePlayer(name)
	lobby.connections[clientId] = conn
	JoinedLobbies[clientId] = lobby
}

func (lobby *Lobby) GetClientData(clientId clientId) (*entities.Player, net.Conn, error) {
	player, ok := lobby.players[clientId]
	if !ok {
		return nil, nil, fmt.Errorf("could not find player from client id %d in lobby %s", clientId, lobby.Name)
	}
	conn, ok := lobby.connections[clientId]
	if !ok {
		return nil, nil, fmt.Errorf("could not find connection from client id %d in lobby %s", clientId, lobby.Name)
	}

	return player, conn, nil
}

func (lobby *Lobby) SendPacketToAll(packetID int, data []byte) []error {
	errList := []error{}

	for _, conn := range(lobby.connections) {
		err := sendPacket(conn, packetID, data)
		if err != nil { errList = append(errList, err) }
	}

	return errList
}

func GetLobbyFromClient(clientId clientId) (*Lobby, bool) {
	lobby, ok := JoinedLobbies[clientId]
	return lobby, ok
}

func CreateLobby(name string, host clientId) *Lobby {
	lobbyID := int(host)

	for {
		_, ok := Lobbies[lobbyID]
		if !ok { break }
		lobbyID++
	}

	lobby := &Lobby{
		Name: name,
		Host: host,
		ID: lobbyID,
		players: make(map[clientId]*entities.Player),
		connections: make(map[clientId]net.Conn),
	}

	Lobbies[lobbyID] = lobby

	return lobby
}