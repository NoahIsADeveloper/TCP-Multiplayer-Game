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

func NewPlayer(name string) *Player {
	player := &Player{
        Entity: *NewEntity(),
    }
	player.Rename(name)
	player.Move(1 << 15, 1 << 15, 0)

	return player
}