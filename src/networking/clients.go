package networking

import (
	"net"
	"fmt"
	"sync"
	"potato-bones/src/utils"
	"potato-bones/src/globals"
)

type clientID uint32
var clientIdManager *utils.IDManager[clientID]
var connections = make(map[clientID]*utils.SafeConn)
var clientMutex sync.RWMutex

func InitNetworking() {
	clientIdManager = utils.NewIDManager(clientID(*globals.MaxClients))

	initLobby()
}

func addConnection(clientId clientID, sconn *utils.SafeConn) {
	clientMutex.Lock(); defer clientMutex.Unlock()
	connections[clientId] = sconn
}

func removeConnection(clientId clientID) {
	clientMutex.Lock(); defer clientMutex.Unlock()
	delete(connections, clientId)
}

func StartUpdateLoop(tickrate int) {
	// Nothing for now
}

func HandleClient(conn net.Conn) error {
	sconn := utils.NewSafeConn(conn)
	clientId, ok := clientIdManager.Get()
	if !ok {
		conn.Close();
		return fmt.Errorf("couldn't get a client id for %s", conn.RemoteAddr())
	}
	addConnection(clientId, sconn)

	defer fmt.Printf("Client %d connection closed.\n", clientId)
	defer sconn.Close()
	defer clientIdManager.Release(clientId)
	defer removeConnection(clientId)
	defer func() {
		lobby, ok := GetLobbyFromClient(clientId)
		if ok {
			lobby.RemovePlayer(clientId)
		}
	}()
	defer func() {
		if panic := recover(); panic != nil {
			fmt.Printf("Recovered from panic for client %d: %v\n", clientId, panic)
		}
	}()

	fmt.Printf("Client %d connected from %s.\n", clientId, conn.RemoteAddr().String())

	for {
		packetID, packetData, err := sconn.ReadPacket()
		if err != nil {
			fmt.Printf("Couldn't read packet from client %d, encountered error %v.\n", clientId, err)
			return err
		}

		err = HandlePacket(sconn, clientId, packetID, packetData)
		if err != nil {
			fmt.Printf("Couldn't handle packet from client %d id %d, encountered error %v.\n", clientId, packetID, err)
			return err
		}
	}
}