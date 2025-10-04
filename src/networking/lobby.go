package networking

import (
	"potato-bones/src/globals"
	"potato-bones/src/utils"
	"potato-bones/src/environment/entities"
	"fmt"
	"sync"
)

type lobbyID uint32
var lobbies = make(map[lobbyID]*Lobby)
var joinedLobbies = make(map[clientID]*Lobby)
var lobbyIdManager *utils.IDManager[lobbyID]

var globalLobbyMutex sync.RWMutex

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

func (lobby *Lobby) Release() {
	lobbyIdManager.Release(lobbyID(lobby.id))
}

func (lobby *Lobby) GetPlayers() map[clientID]*entities.Player {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock();
	return lobby.players
}

func (lobby *Lobby) GetName() string {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock();
	return lobby.name
}

func (lobby *Lobby) SwapName(name string) {
	lobby.mutex.Lock(); defer lobby.mutex.Unlock()
	lobby.name = name
}

func (lobby *Lobby) GetID() lobbyID {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock();
	return lobby.id
}

func (lobby *Lobby) GetHost() clientID {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock();
	return lobby.host
}

func (lobby *Lobby) SwapHost(clientId clientID) (bool) {
	if !lobby.HasClient(clientId) { return false }
	lobby.mutex.Lock()
	lobby.host = clientId
	lobby.mutex.Unlock()

	scSyncEntireLobby(lobby)
	return true
}

func (lobby *Lobby) HasClient(clientId clientID) (bool) {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock()
	_, ok := lobby.connections[clientId]
	return ok
}

func (lobby *Lobby) AddPlayer(clientId clientID, name string, sconn *utils.SafeConn) error {
	lobby.mutex.Lock()

	_, ok := joinedLobbies[clientId]
	if ok {
		lobby.mutex.Unlock();
		return fmt.Errorf("cannot add client %d as player %s they are already in a lobby", clientId, name)
	}

	lobby.players[clientId] = entities.NewPlayer(name)
	lobby.connections[clientId] = sconn
	joinedLobbies[clientId] = lobby

	if *globals.DebugLobbyInfo {
		fmt.Printf("added client %d as player %s to lobby %d\n", clientId, name, lobby.id)
	}

	lobby.mutex.Unlock();
	scSyncLobbyPlayers(lobby)

	return nil
}

func (lobby *Lobby) RemovePlayer(clientId clientID) {
	lobby.mutex.Lock();

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

		lobby.mutex.Unlock();
		scSyncEntireLobby(lobby)
		lobby.mutex.Lock();
	}
	lobby.mutex.Unlock();
	scSyncLobbyPlayers(lobby)
	lobby.mutex.Lock();

	if *globals.DebugLobbyInfo {
		fmt.Printf("removed client %d from lobby %d\n", clientId, lobby.id)
	}

	lobby.mutex.Unlock();
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


func (lobby *Lobby) SendPacketToAllUDP(packetId int, data []byte) {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock()
	for _, sconn := range(lobby.connections) {
		sconn.SendPacketUDP(*udpConn, packetId, data)
	}
}

func (lobby *Lobby) SendPacketToAllTCP(packetId int, data []byte) {
	lobby.mutex.RLock(); defer lobby.mutex.RUnlock()

	for _, sconn := range(lobby.connections) {
		sconn.SendPacketTCP(packetId, data)
	}
}

func GetLobbyFromClient(clientId clientID) (*Lobby, bool) {
	globalLobbyMutex.RLock(); defer globalLobbyMutex.RUnlock()

	lobby, ok := joinedLobbies[clientId]
	return lobby, ok
}

func GetLobbyFromId(lobbyId lobbyID) (*Lobby, bool) {
	globalLobbyMutex.RLock(); defer globalLobbyMutex.RUnlock()

	lobby, ok := lobbies[lobbyId]
	return lobby, ok
}

func CreateLobby(name string, host clientID) (*Lobby, error) {
	globalLobbyMutex.Lock(); defer globalLobbyMutex.Unlock()

	lobbyID, ok := lobbyIdManager.Get()
	if !ok {
		return nil, fmt.Errorf("max lobby limited reached")
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
		fmt.Printf("created lobby %s %d with host %d\n", name, lobbyID, host)
	}

	return lobby, nil
}