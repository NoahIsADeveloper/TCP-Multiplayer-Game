package utils

import (
	"fmt"
	"net"
	"potato-bones/src/globals"
	"potato-bones/src/networking/datatypes"
	"sync"
)

type SafeConn struct {
    tcpConn net.Conn
	udpAddr net.UDPAddr

	session *Session
    mutex sync.RWMutex
}

//TODO lpease find a better way to impliemnt this hoy fucking shit
func DecodeVarInt(sconn *SafeConn) (int, error) {
	var result int
	var shift uint
	bfr := make([]byte, 1)
	for {
		_, err := sconn.ReadTCP(bfr)
		if err != nil {
			return 0, err
		}
		value := bfr[0]
		result |= int(value & 0x7F) << shift
		if value & 0x80 == 0 {
			break
		}
		shift += 7
		if shift > 35 {
			return 0, fmt.Errorf("varint: too big")
		}
	}
	return result, nil
}

func (sconn *SafeConn) WriteTCP(data []byte) (int, error) {
    sconn.mutex.Lock(); defer sconn.mutex.Unlock()
    return sconn.tcpConn.Write(data)
}

func (sconn *SafeConn) ReadTCP(data []byte) (int, error) {
	sconn.mutex.RLock(); defer sconn.mutex.RUnlock()
	return sconn.tcpConn.Read(data)
}

func (sconn *SafeConn) Close() error {
	return sconn.tcpConn.Close()
}

func (sconn *SafeConn) SendPacketTCP(packetID int, data []byte) error {
	if *globals.DebugShowOutgoing {
		fmt.Printf("[DEBUG] Sending packet ID %d\n", packetID)
	}

	//TODO what the fuck
	packet := []byte{}
	length := []byte{}
	datatypes.AppendVarInt(&packet, packetID)
	packet = append(packet, data...)
	datatypes.AppendVarInt(&length, len(packet))
	_, err := sconn.WriteTCP(append(length, packet...))
	return err
}

func (sconn *SafeConn) ReadPacketTCP() (int, []byte, error) {
	length, err := DecodeVarInt(sconn)
	if err != nil {
		return 0, nil, err
	}

	if length < 1 {
		return 0, nil, fmt.Errorf("packet too short")
	}

	if length > *globals.MaxPacketSize {
		return 0, nil, fmt.Errorf("packet too large: %d", length)
	}

	packetData := make([]byte, length)
	totalRead := 0
	for totalRead < length {
		read, err := sconn.ReadTCP(packetData[totalRead:])
		if err != nil {
			return 0, nil, err
		}
		totalRead += read
	}

	offset := 0
	packetID, err := datatypes.ReadVarInt(packetData, &offset)
	if err != nil { return 0, nil, err }
	if offset == 0 {
		return 0, nil, fmt.Errorf("invalid packet ID")
	}

	data := packetData[offset:]
	return packetID, data, nil
}

func (sconn *SafeConn) AddUDPAddr(addr net.UDPAddr) {
    sconn.mutex.Lock(); defer sconn.mutex.Unlock()
	sconn.udpAddr = addr
}

func NewSafeConn(tcpConn net.Conn, session *Session) *SafeConn {
	return &SafeConn{
		tcpConn: tcpConn,
		session: session,
	}
}