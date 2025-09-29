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

func (player *Player) GetPosition() (uint16, uint16) {
	return player.X, player.Y
}

func CreateEntity() *Entity {
	entity := &Entity{kind: 0}
	entity.Move(1 << 15, 1 << 15)

	return entity
}
