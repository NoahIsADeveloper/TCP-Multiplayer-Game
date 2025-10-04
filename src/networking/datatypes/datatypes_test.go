package datatypes

import (
	"math"
	"testing"
)

func TestStringEncodeDecode(t *testing.T) {
	roundTrip(t, func(data *[]byte, s string) { AppendString(data, s) }, ReadString, "test")
}

func TestVarIntEncodeDecode(t *testing.T) {
	roundTrip(t, AppendVarInt, ReadVarInt, 123456)
}

func TestRotationEncodeDecode(t *testing.T) {
	roundTripFloat32(t, AppendRotation, ReadRotation, 2.3912, 0.04)
}

func TestPositionEncodeDecode(t *testing.T) {
	roundTripPosition(t, AppendPosition, ReadPosition, 3931, 9213)
}

func TestVarIntMultiple(t *testing.T) {
	cases := []int{
		0,
		1,
		123456,
		98765,
		1,
		9812312893,
		3213,
	}

	for _, c := range cases {
		roundTrip(t, AppendVarInt, ReadVarInt, c)
	}
}

func TestStringMultiple(t *testing.T) {
	cases := []string{
		"",
		"a",
		"hello",
		"こんにちは",
		"long string with spaces and symbols !@#",
		"😭",
	}

	for _, c := range cases {
		t.Run(c, func(t *testing.T) {
			roundTrip(t, func(data *[]byte, s string) { AppendString(data, s) }, ReadString, c)
		})
	}
}

func TestRotationMultiple(t *testing.T) {
	cases := []float32{
		0,
		3.14159,
		2.3912,
	}

	epsilon := float32(0.0001)
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			roundTripFloat32(t, AppendRotation, ReadRotation, c, epsilon)
		})
	}
}

func TestPositionMultiple(t *testing.T) {
	cases := [][2]uint16{
		{0, 0},
		{1, 1},
		{3931, 9213},
		{65535, 65535},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			roundTripPosition(t, AppendPosition, ReadPosition, c[0], c[1])
		})
	}
}

func roundTripFloat32(t *testing.T, appendFunc func(*[]byte, float32), readFunc func([]byte, *int) (float32, error), original float32, epsilon float32) {
	t.Helper()
	data := []byte{}
	appendFunc(&data, original)

	offset := 0
	result, err := readFunc(data, &offset)
	if err != nil {
		t.Fatal(err)
	}

	if math.Abs(float64(result-original)) > float64(epsilon) {
		t.Errorf("round-trip failed: got %v, want %v", result, original)
	}
}

func roundTrip[T comparable](t *testing.T, appendFunc func(*[]byte, T), readFunc func([]byte, *int) (T, error), original T) {
	t.Helper()
	data := []byte{}
	appendFunc(&data, original)

	offset := 0
	result, err := readFunc(data, &offset)
	if err != nil {
		t.Fatal(err)
	}

	if result != original {
		t.Errorf("round-trip failed: got %+v, want %+v", result, original)
	}
}

func roundTripPosition(t *testing.T, appendFunc func(*[]byte, uint16, uint16), readFunc func([]byte, *int) (uint16, uint16, error), x, y uint16) {
	t.Helper()
	data := []byte{}
	appendFunc(&data, x, y)

	offset := 0
	readX, readY, err := readFunc(data, &offset)
	if err != nil {
		t.Fatal(err)
	}

	if readX != x || readY != y {
		t.Errorf("round-trip failed: got %d %d, want %d %d", readX, readY, x, y)
	}
}
