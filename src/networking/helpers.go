package networking

import (
	"fmt"
	"net"
)

const PACKET_MAX_SIZE = 1_000_000

// // // // // // //
// Packet format: //
// // // // // // //
// Packet length: VarInt
// Packet ID: VarInt
// Data: any

func decodeVarInt(conn net.Conn) (int, error) {
	var result int
	var shift uint
	bfr := make([]byte, 1)
	for {
		_, err := conn.Read(bfr)
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
			return 0, fmt.Errorf("VarInt too big")
		}
	}
	return result, nil
}

func decodeVarIntFromBytes(data []byte) (int, int) {
	var result int
	var shift uint
	var bytesRead int
	
	for index, reading := range data {
		result |= int(reading & 0x7F) << shift
		bytesRead = index + 1
		if reading & 0x80 == 0 {
			break
		}
		shift += 7
		if shift > 35 {
			return 0, 0
		}
	}
	return result, bytesRead
}

func encodeVarInt(value int) []byte {
	var result []byte
	for {
		msb := byte(value & 0x7F)
		value >>= 7
		if value != 0 {
			msb |= 0x80
		}
		result = append(result, msb)
		if value == 0 {
			break
		}
	}
	return result
}

func readPacket(conn net.Conn) (int, []byte, error) {
	length, err := decodeVarInt(conn)
	if err != nil {
		return 0, nil, err
	}

	if length < 1 {
		return 0, nil, fmt.Errorf("packet too short")
	}

	if length > PACKET_MAX_SIZE {
		return 0, nil, fmt.Errorf("packet too large: %d", length)
	}

	packetData := make([]byte, length)
	totalRead := 0
	for totalRead < length {
		read, err := conn.Read(packetData[totalRead:])
		if err != nil {
			return 0, nil, err
		}
		totalRead += read
	}
	
	packetID, idBytes := decodeVarIntFromBytes(packetData)
	if idBytes == 0 {
		return 0, nil, fmt.Errorf("invalid packet ID")
	}
	
	data := packetData[idBytes:]
	return packetID, data, nil
}

func sendPacket(conn net.Conn, packetID int, data []byte) error {
	encodedID := encodeVarInt(packetID)
	packet := append(encodedID, data...)
	length := encodeVarInt(len(packet))
	_, err := conn.Write(append(length, packet...))
	return err
}

func appendString(list *[]byte, str string) {
	data:= []byte(str)
	*list = append(*list, encodeVarInt(len(data))...)
	*list = append(*list, data...)
}