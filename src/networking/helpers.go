package networking

import (
	"fmt"
	"net"
	"encoding/binary"
)

const PACKET_MAX_SIZE = 1_000_000

// // // // // // //
// Packet format: //
// // // // // // //
// Packet length: VarInt
// Packet ID: VarInt
// Data: any

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

func sendPacket(conn net.Conn, packetID int, data []byte) error {
	encodedID := encodeVarInt(packetID)
	packet := append(encodedID, data...)
	length := encodeVarInt(len(packet))
	_, err := conn.Write(append(length, packet...))
	return err
}

func sendPacketToAll(packetID int, data []byte) error {
	for _, conn := range(connections) {
		err := sendPacket(conn, packetID, data)
		if err != nil { return err }
	}

	return nil
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

func appendVarInt(list *[]byte, value int) {
	*list = append(*list, encodeVarInt(value)...)
}

func appendString(list *[]byte, str string) {
	data := []byte(str)
	appendVarInt(list, len(data))
	*list = append(*list, data...)
}

func appendPosition(list *[]byte, x uint16, y uint16) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[0:2], uint16(x))
	binary.BigEndian.PutUint16(buf[2:4], uint16(y))
	*list = append(*list, buf...)
}

func readItemArray(data []byte, offset *int, read func(data []byte, offset *int) any) ([]any, error) {
	length, err := readVarInt(data, offset)
	if err != nil {
		return nil, err
	}

	result := make([]any, 0, length)

	for range length {
		item := read(data, offset)
		result = append(result, item)
	}

	return result, nil
}

func readVarInt(data []byte, offset *int) (int, error) {
	var result int
	var shift uint

	for {
		if *offset >= len(data) {
			return 0, fmt.Errorf("varint: unexpected end of data")
		}
		b := data[*offset]
		*offset++

		result |= int(b & 0x7F) << shift

		if b & 0x80 == 0 {
			break
		}
		shift += 7
		if shift > 35 {
			return 0, fmt.Errorf("varint too big")
		}
	}
	return result, nil
}

func readString(data []byte, offset *int) (string, error) {
	length, err := readVarInt(data, offset)
	if err != nil {
		return "", err
	}
	if length < 0 || *offset+length > len(data) {
		return "", fmt.Errorf("string: invalid length %d", length)
	}
	str := string(data[*offset : *offset+length])
	*offset += length
	return str, nil
}

func readPosition(data []byte, offset *int) (uint16, uint16, error) {
	if *offset + 4 > len(data) {
		return 0, 0, fmt.Errorf("position: not enough data")
	}
	x := binary.BigEndian.Uint16(data[*offset : *offset+2])
	y := binary.BigEndian.Uint16(data[*offset+2 : *offset+4])
	*offset += 4
	return x, y, nil
}