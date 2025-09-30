package entities

type Entity struct{
	kind uint8

	x uint16
	y uint16
	rotation float32
}

func (entity *Entity) Move(x uint16, y uint16) {
	entity.x = x
	entity.y = y
}

func (entity *Entity) GetPosition() (uint16, uint16) {
	return entity.x, entity.y
}

func (entity *Entity) InRange(x, y, distance uint16) bool {
	distX := int32(entity.x) - int32(x)
	distY := int32(entity.y) - int32(y)
	return (distX * distX + distY * distY) <= (int32(distance)*int32(distance))
}

func CreateEntity() *Entity {
	entity := &Entity{kind: 0}
	entity.Move(1 << 15, 1 << 15)

	return entity
}