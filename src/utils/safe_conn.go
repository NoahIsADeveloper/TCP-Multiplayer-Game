package utils

import (
	"fmt"
	"net"
	"potato-bones/src/globals"
	"potato-bones/src/networking/datatypes"
	"sync"
)

type SafeConn struct {
    conn net.Conn
    mutex sync.Mutex
}

//TODO lpease find a better way to impliemnt this hoy fucking shit
func DecodeVarInt(sconn *SafeConn) (int, error) {
	var result int
	var shift uint
	bfr := make([]byte, 1)
	for {
		_, err := sconn.Read(bfr)
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

func (sconn *SafeConn) Write(data []byte) (int, error) {
    sconn.mutex.Lock()
    defer sconn.mutex.Unlock()
    return sconn.conn.Write(data)
}

func (sconn *SafeConn) Read(data []byte) (int, error) {
	return sconn.conn.Read(data)
}

func (sconn *SafeConn) Close() error {
	return sconn.conn.Close()
}

func (sconn *SafeConn) SendPacket(packetID int, data []byte) error {
	if *globals.DebugShowOutgoing {
		fmt.Printf("[DEBUG] Sending packet ID %d\n", packetID)
	}

	//TODO what the fuck
	packet := []byte{}
	length := []byte{}
	datatypes.AppendVarInt(&packet, packetID)
	packet = append(packet, data...)
	datatypes.AppendVarInt(&length, len(packet))
	_, err := sconn.Write(append(length, packet...))
	return err
}

func (sconn *SafeConn) ReadPacket() (int, []byte, error) {
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
		read, err := sconn.Read(packetData[totalRead:])
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

func NewSafeConn(conn net.Conn) *SafeConn {
	return &SafeConn{conn: conn}
}