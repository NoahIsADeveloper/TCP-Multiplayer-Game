package datatypes

import "fmt"

func AppendString(list *[]byte, str string) {
	data := []byte(str)
	AppendVarInt(list, len(data))
	*list = append(*list, data...)
}

func ReadString(data []byte, offset *int) (string, error) {
	length, err := ReadVarInt(data, offset)
	if err != nil {
		return "", err
	}

	if length < 0 || *offset + length > len(data) {
		return "", fmt.Errorf("string: invalid length %d", length)
	}

	str := string(data[*offset : *offset+length])
	*offset += length
	return str, nil
}
