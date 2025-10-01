package networking

import (
	"potato-bones/src/globals"
	"potato-bones/src/utils"
	"potato-bones/src/environment/entities"
	"fmt"
	"sync"
)

type lobbyID uint32
var lobbyMutex sync.RWMutex
var lobbies = make(map[lobbyID]*Lobby)
var joinedLobbies = make(map[clientID]*Lobby)
var lobbyIdManager *utils.IDManager[lobbyID]

type Lobby struct{
	name string
	host clientID
	id lobbyID

	players map[clientID]*entities.Player
	connections map[clientID]*utils.SafeConn

	mutex sync.RWMutex
}

func initLobby() {
	lobbyIdManager = utils.NewIDManager(lobbyID(*globals.MaxLobbies))
}

func (lobby *Lobby) Rename(name string) {
	lobby.mutex.Lock(); defer lobby.mutex.Unlock()

	lobby.name = name
}

func (lobby *Lobby) Release() {
	lobbyIdManager.Release(lobbyID(lobby.id))
}

func (lobby *Lobby) RemovePlayer(clientId clientID) {
	lobby.mutex.Lock(); defer lobby.mutex.Unlock()

	delete(lobby.players, clientId)
	delete(lobby.connections, clientId)
	delete(joinedLobbies, clientId)

	if len(lobby.players) == 0 {
		delete(lobbies, lobby.id)
	} else if lobby.host == clientId {
		for newHost := range lobby.players {
			lobby.host = newHost
			break
		}

		//TODO sync lobby
	}
	//TODO sync players

	if *globals.DebugLobbyInfo {
		fmt.Printf("[DEBUG] Removed client %d from lobby %d\n", clientId, lobby.id)
	}
}

func (lobby *Lobby) AddPlayer(clientId clientID, name string, sconn *utils.SafeConn) error {
	lobby.mutex.Lock(); defer lobby.mutex.Unlock()

	_, ok := joinedLobbies[clientId]
	if ok {
		return fmt.Errorf("cannot add client %d as they are already in a lobby", clientId)
	}

	lobby.players[clientId] = entities.NewPlayer(name)
	lobby.connections[clientId] = sconn
	joinedLobbies[clientId] = lobby

	if *globals.DebugLobbyInfo {
		fmt.Printf("[DEBUG] Added client %d to lobby %d\n", clientId, lobby.id)
	}

	//TODO sync players
	return nil
}

func (lobby *Lobby) GetClientData(clientId clientID) (*entities.Player, *utils.SafeConn, error) {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock()

	player, ok := lobby.players[clientId]
	if !ok {
		return nil, nil, fmt.Errorf("could not find player from client id %d in lobby %s", clientId, lobby.name)
	}
	sconn, ok := lobby.connections[clientId]
	if !ok {
		return nil, nil, fmt.Errorf("could not find connection from client id %d in lobby %s", clientId, lobby.name)
	}

	return player, sconn, nil
}

func (lobby *Lobby) HasClient(clientId clientID) (bool) {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock()

	_, ok := lobby.connections[clientId]
	return ok
}

func (lobby *Lobby) SendPacketToAll(packetID int, data []byte) []error {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock()

	errList := []error{}

	for _, sconn := range(lobby.connections) {
		err := sconn.SendPacket(packetID, data)
		if err != nil { errList = append(errList, err) }
	}

	return errList
}

func GetLobbyFromClient(clientId clientID) (*Lobby, bool) {
	lobbyMutex.RLock(); defer lobbyMutex.RUnlock()

	lobby, ok := joinedLobbies[clientId]
	return lobby, ok
}

func CreateLobby(name string, host clientID) (*Lobby, error) {
	lobbyMutex.RLock(); defer lobbyMutex.RUnlock()

	lobbyID, ok := lobbyIdManager.Get()
	if !ok {
		return nil, fmt.Errorf("max lobby limited reached")
	}

	for {
		_, ok := lobbies[lobbyID]
		if !ok { break }
		lobbyID++
	}

	lobby := &Lobby{
		name: name,
		host: host,
		id: lobbyID,
		players: make(map[clientID]*entities.Player),
		connections: make(map[clientID]*utils.SafeConn),
	}

	lobbies[lobbyID] = lobby
	if *globals.DebugLobbyInfo {
		fmt.Printf("[DEBUG] Created lobby %s %d with host %d\n", name, lobbyID, host)
	}

	return lobby, nil
}