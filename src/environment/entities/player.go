package entities

type Player struct{
	Entity
	Name string
}

func CreatePlayer(name string) *Player {
	player := &Player{
        Entity: Entity{kind: 1},
        Name: name,
    }
	player.Move(1 << 15, 1 << 15)

	return player
}