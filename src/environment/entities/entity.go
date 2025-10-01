package entities

import "sync"

type Entity struct{
	x uint16
	y uint16
	rotation float32
	mutex sync.RWMutex
}

func (entity *Entity) Update(deltaTime float64) {}

func (entity *Entity) Move(x uint16, y uint16, rotation float32) {
	entity.mutex.Lock(); defer entity.mutex.Unlock()
	entity.x = x
	entity.y = y
	entity.rotation = rotation
}

func (entity *Entity) GetPosition() (uint16, uint16, float32) {
	entity.mutex.RLock(); defer entity.mutex.RUnlock()
	return entity.x, entity.y, entity.rotation
}

func (entity *Entity) InRange(x, y, distance uint16) bool {
	entity.mutex.RLock(); defer entity.mutex.RUnlock()
	distX := int32(entity.x) - int32(x)
	distY := int32(entity.y) - int32(y)
	return (distX * distX + distY * distY) <= (int32(distance)*int32(distance))
}

func NewEntity() *Entity {
	entity := &Entity{}
	entity.Move(1 << 15, 1 << 15, 0)

	return entity
}