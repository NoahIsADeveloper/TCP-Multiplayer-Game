package entities

type Player struct{
	Entity
	name string
}

func (player *Player) Rename(name string) {
	player.name = name
}

func (player *Player) GetName() string {
	return  player.name
}

func CreatePlayer(name string) *Player {
	player := &Player{
        Entity: *CreateEntity(),
    }
	player.Rename(name)
	player.Move(1 << 15, 1 << 15, 0)

	return player
}