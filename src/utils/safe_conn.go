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
	hasUdp  bool
	session *Session

	readMutex  sync.Mutex
	writeMutex sync.Mutex
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

func encodePacket(packetId int, data []byte) []byte {
	packet := []byte{}
	length := []byte{}
	datatypes.AppendVarInt(&packet, packetId)
	packet = append(packet, data...)
	datatypes.AppendVarInt(&length, len(packet))

	return append(length, packet...)
}

func (sconn *SafeConn) WriteUDP(conn net.UDPConn, data []byte) (int, error) {
	if !sconn.hasUdp { return 0, nil }
	value, err := conn.WriteToUDP(data, &sconn.udpAddr)
	return value, err
}

func (sconn *SafeConn) WriteTCP(data []byte) (int, error) {
    value, err := sconn.tcpConn.Write(data)
	return value, err
}

func (sconn *SafeConn) ReadTCP(data []byte) (int, error) {
	value, err := sconn.tcpConn.Read(data)

	return value, err
}

func (sconn *SafeConn) Close() error {
	return sconn.tcpConn.Close()
}

func (sconn *SafeConn) SendPacketUDP(conn net.UDPConn, packetId int, data []byte) error {
	if *globals.OnlySendTCP {
		return sconn.SendPacketTCP(packetId, data)
	}

	sconn.writeMutex.Lock(); defer sconn.writeMutex.Unlock()

	if *globals.DebugShowOutgoing {
		fmt.Printf("sending udp packet id %d with data %v\n", packetId, data)
	}

	_, err := sconn.WriteUDP(conn, encodePacket(packetId, data))
	return err
}

func (sconn *SafeConn) SendPacketTCP(packetId int, data []byte) error {
	sconn.writeMutex.Lock(); defer sconn.writeMutex.Unlock()

	if *globals.DebugShowOutgoing {
		fmt.Printf("sending tcp packet id %d with data %v\n", packetId, data)
	}

	_, err := sconn.WriteTCP(encodePacket(packetId, data))
	return err
}

func (sconn *SafeConn) ReadPacketTCP() (int, []byte, error) {
	sconn.readMutex.Lock(); defer sconn.readMutex.Unlock()

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
    sconn.writeMutex.Lock(); defer sconn.writeMutex.Unlock()
	sconn.udpAddr = addr
	sconn.hasUdp = true
}

func NewSafeConn(tcpConn net.Conn, session *Session) *SafeConn {
	return &SafeConn{
		tcpConn: tcpConn,
		session: session,
		hasUdp: false,
	}
}