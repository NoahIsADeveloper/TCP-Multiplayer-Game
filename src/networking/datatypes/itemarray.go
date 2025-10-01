package datatypes

// I have no idea how you would write this outside of manually?
func ReadItemArray(data []byte, offset *int, read func(data []byte, offset *int) any) ([]any, error) {
	length, err := ReadVarInt(data, offset)
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
