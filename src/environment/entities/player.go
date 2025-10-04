package entities

type Player struct{
	Entity
	name string

	doUpdate bool
	updateNumber int
}

func (player *Player) Move(x uint16, y uint16, rotation float32, updateNumber int) {
	player.mutex.Lock(); defer player.mutex.Unlock()

	if updateNumber <= player.updateNumber { return }

	player.x = x
	player.y = y
	player.rotation = rotation
	player.updateNumber = updateNumber

	rotationDiff := player.rotation - rotation
	if player.x == x && player.y == y && (rotationDiff > 0.04 || rotationDiff < -0.04) {
		return
	}

	player.doUpdate = true
}

func (player *Player) DoUpdate() bool {
	player.mutex.Lock(); defer player.mutex.Unlock()
	if !player.doUpdate {
		return false
	}

	player.doUpdate = false
	return true
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
	player.Move(1 << 15, 1 << 15, 0, 0)

	return player
}