package datatypes

import (
	"encoding/binary"
	"fmt"
)

func AppendPosition(list *[]byte, x uint16, y uint16) {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[0:2], uint16(x))
	binary.BigEndian.PutUint16(buf[2:4], uint16(y))
	*list = append(*list, buf...)
}

func ReadPosition(data []byte, offset *int) (uint16, uint16, error) {
	if *offset + 4 > len(data) {
		return 0, 0, fmt.Errorf("position: not enough data")
	}

	x := binary.BigEndian.Uint16(data[*offset : *offset+2])
	y := binary.BigEndian.Uint16(data[*offset+2 : *offset+4])

	*offset += 4
	return x, y, nil
}