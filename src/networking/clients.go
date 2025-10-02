package networking

import (
	"errors"
	"fmt"
	"io"
	"net"
	"potato-bones/src/globals"
	"potato-bones/src/networking/datatypes"
	"potato-bones/src/utils"
	"sync"
)

type clientID uint32

var sessionManager *utils.SessionManager
var clientIdManager *utils.IDManager[clientID]

var connectionsFromClientId = make(map[clientID]*utils.SafeConn)
var clientIdFromSessionId = make(map[string]clientID)
var clientMutex sync.RWMutex

func InitNetworking() {
	clientIdManager = utils.NewIDManager(clientID(*globals.MaxClients))
	sessionManager = utils.NewSessionManager()

	initLobby()
}

func addClient(sconn *utils.SafeConn, session *utils.Session, clientId clientID) {
	clientMutex.Lock(); defer clientMutex.Unlock()
	connectionsFromClientId[clientId] = sconn
	clientIdFromSessionId[session.ID] = clientId
}

func removeClient(clientId clientID, session *utils.Session) {
	clientMutex.Lock(); defer clientMutex.Unlock()
	delete(connectionsFromClientId, clientId)
	delete(clientIdFromSessionId, session.ID)
}

func StartUpdateLoop(tickrate int) {
	// Nothing for now
}

func HandleUDPPacket(addr *net.UDPAddr, data []byte) {
	defer func() {
		if panic := recover(); panic != nil {
			fmt.Printf("Recovered from panic from udp %s, encountered error %v.\n", addr.Network(), panic)
		}
	}()

	var offset int = 0

	// TODO errors here get voided right now after they get returned, please fixy :)
	sessionId, err := datatypes.ReadString(data, &offset)
	if err != nil {
		fmt.Printf("error reading session id %v\n", err)
		return
	}

	_, ok := sessionManager.GetSession(sessionId)
	if !ok {
		fmt.Printf("client attempted to use invalid session %s\n", sessionId)
		return
	}

	clientId, ok := clientIdFromSessionId[sessionId]
	if !ok {
		fmt.Println("server invalid session id")
		return
	}
	sconn, ok := connectionsFromClientId[clientId]
	if !ok {
		fmt.Println("server invalid client id")
		return
	}

	length, err := datatypes.ReadVarInt(data, &offset)
	if err != nil {
		fmt.Printf("error reading udp packet length %v\n", err)
		return
	}

	packetData:= data[offset:offset + length]

	length = 0
	packetId, err := datatypes.ReadVarInt(packetData, &length)
	if err != nil {
		fmt.Printf("error reading udp packet id %v\n", err)
		return
	}
	packetData = packetData[length:]

	err = HandlePacket(sconn, clientId, packetId, packetData)
	if err != nil {
		fmt.Printf("error handling udp packet %v\n", err)
		return
	}
}

func HandleTCPClient(conn net.Conn) error {
	session := sessionManager.CreateSession()
	sconn := utils.NewSafeConn(conn, session)
	clientId, ok := clientIdManager.Get()
	if !ok {
		conn.Close();
		return fmt.Errorf("couldn't get a client id for %s", conn.RemoteAddr())
	}
	addClient(sconn, session, clientId)

	defer sessionManager.KillSession(session.ID)
	defer fmt.Printf("Client %d connection closed.\n", clientId)
	defer sconn.Close()
	defer clientIdManager.Release(clientId)
	defer removeClient(clientId, session)
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
		packetID, packetData, err := sconn.ReadPacketTCP()
		if err != nil {
			fmt.Printf("Couldn't read packet from client %d, encountered error %v.\n", clientId, err)

			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return err
			} else { continue }
		}

		err = HandlePacket(sconn, clientId, packetID, packetData)
		if err != nil {
			fmt.Printf("Couldn't handle packet from client %d id %d, encountered error %v.\n", clientId, packetID, err)

			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return err
			} else { continue }
		}
	}
}