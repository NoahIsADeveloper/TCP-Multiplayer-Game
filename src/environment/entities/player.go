package entities

type Player struct{
	Entity
	name string
}

func (player *Player) GetName() string {
	player.mutex.RLock(); defer player.mutex.RUnlock()
	return player.name
}

func NewPlayer(name string) *Player {
	player := &Player{
        Entity: *NewEntity(),
		name: name,
    }
	player.Move(1 << 15, 1 << 15, 0)

	return player
}