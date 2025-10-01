package datatypes

import (
	"fmt"
)

func ReadVarInt(data []byte, offset *int) (int, error) {
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
			return 0, fmt.Errorf("varint: too big")
		}
	}
	return result, nil
}

func AppendVarInt(list *[]byte, value int) {
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

	*list = append(*list, result...)
}