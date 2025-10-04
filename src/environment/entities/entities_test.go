package entities

import (
	"testing"
)

func TestEntityMove(t *testing.T) {
	entity := NewEntity()

	entity.Move(100, 200, 1.5)

	x, y, rotation := entity.GetPosition()
	if x != 100 || y != 200 || rotation != 1.5 {
		t.Errorf("Move/GetPosition failed: got (%d, %d, %f), want (100, 200, 1.5)", x, y, rotation)
	}
}

func TestEntityInRange(t *testing.T) {
	entity := NewEntity()
	entity.Move(50, 50, 0)

	tests := []struct {
		x, y, distance uint16
		want           bool
	}{
		{50, 50, 0, true},
		{51, 50, 1, true},
		{52, 50, 1, false},
		{60, 60, 15, true},
		{60, 60, 5, false},
		{60, 50, 10, true},
		{50, 0, 50, true},
	}

	for _, testValues := range tests {
		got := entity.InRange(testValues.x, testValues.y, testValues.distance)
		if got != testValues.want {
			t.Errorf("InRange(%d,%d,%d) = %v; want %v", testValues.x, testValues.y, testValues.distance, got, testValues.want)
		}
	}
}