package utils

import "sync"

type IDType interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64
}

type IDManager[ID IDType] struct {
	mutex sync.Mutex
	nextID ID
	freeIDs []ID
	maxID ID
}

func (manager *IDManager[ID]) Get() (ID, bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()

	var id ID
	if len(manager.freeIDs) > 0 {
		id = manager.freeIDs[len(manager.freeIDs)-1]
		manager.freeIDs = manager.freeIDs[:len(manager.freeIDs)-1]
		return id, true
	}

	if manager.nextID > manager.maxID {
		var zero ID
		return zero, false
	}

	id = manager.nextID
	manager.nextID++
	return id, true
}

func (manager *IDManager[ID]) Release(id ID) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	manager.freeIDs = append(manager.freeIDs, id)
}

func NewIDManager[ID IDType](max ID) *IDManager[ID] {
	return &IDManager[ID]{
		nextID: 0,
		freeIDs: []ID{},
		maxID: max,
	}
}