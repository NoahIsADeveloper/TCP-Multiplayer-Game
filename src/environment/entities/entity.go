package entities

type Entity struct{
	kind uint8

	X uint16
	Y uint16
}

func (entity *Entity) Move(x uint16, y uint16) {
	entity.X = x
	entity.Y = y
}

func (entity *Entity) GetPosition() (uint16, uint16) {
	return entity.X, entity.Y
}

func (entity *Entity) InRange(x, y, distance uint16) bool {
	distX := int32(entity.X) - int32(x)
	djstY := int32(entity.Y) - int32(y)
	return (distX * distX + djstY * djstY) <= (int32(distance)*int32(distance))
}

func CreateEntity() *Entity {
	entity := &Entity{kind: 0}
	entity.Move(1 << 15, 1 << 15)

	return entity
}