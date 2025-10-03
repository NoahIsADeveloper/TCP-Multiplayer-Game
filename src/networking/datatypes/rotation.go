package datatypes

import (
	"encoding/binary"
	"fmt"
)

func ReadRotation(data []byte, offset *int) (float32, error) {
	if *offset+2 > len(data) {
		return 0, fmt.Errorf("rotation: not enough data")
	}

	raw := binary.BigEndian.Uint16(data[*offset : *offset+2])
	*offset += 2

	rotation := (float32(raw) / 65535.0) * 6.28318
	return rotation, nil
}

func AppendRotation(list *[]byte, rotation float32) {
	value := uint16(rotation / 6.28318 * 65535.0)
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, value)
	*list = append(*list, buf...)
}