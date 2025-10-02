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

	rotationDiff := player.rotation - rotation
	if player.x == x && player.y == y && (rotationDiff > 3 || rotationDiff < -3) {
		return
	}

	player.x = x
	player.y = y
	player.rotation = rotation

	player.doUpdate = true
	player.updateNumber = updateNumber
}

func (player *Player) DoUpdate() bool {
	player.mutex.RLock(); defer player.mutex.RUnlock()
	return player.doUpdate
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