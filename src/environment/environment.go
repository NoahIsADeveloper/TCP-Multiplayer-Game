package environment

import (
	"potato-bones/src/environment/entities"
	"potato-bones/src/globals"
)

var deltaTime float64 = *globals.GameSpeed / float64(*globals.Tickrate)

type Environment struct{
	entities []*entities.Entity
	staticEntities []*entities.Entity
}

func (environemnt *Environment) UpdateGame() {
	for _, entity := range(environemnt.entities) {
		entity.Update(deltaTime)
	}
}

func CreateEnvironment() *Environment {
	environment := &Environment{
		entities: make([]*entities.Entity, *globals.MaxEntities),
	}

	return environment
}