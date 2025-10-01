package environment

import (
	"potato-bones/src/environment/entities"
	"potato-bones/src/globals"
	"sync"
)

var deltaTime float64 = *globals.GameSpeed / float64(*globals.Tickrate)

type Environment struct{
	entities []*entities.Entity
	staticEntities []*entities.Entity
	mu sync.RWMutex
}

func (environment *Environment) UpdateGame() {
	environment.mu.RLock()
    defer environment.mu.RUnlock()

	for _, entity := range(environment.entities) {
		entity.Update(deltaTime)
	}
}

func (environment *Environment) AddEntity(entity *entities.Entity) {
    environment.mu.Lock()
    defer environment.mu.Unlock()

    for index, entity := range environment.entities {
        if entity == nil {
            environment.entities[index] = entity
            return
        }
    }

    environment.entities = append(environment.entities, entity)
}


func CreateEnvironment() *Environment {
	environment := &Environment{
		entities: make([]*entities.Entity, *globals.MaxEntities),
	}

	return environment
}