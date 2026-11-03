package game

import (
	"aeons/internal/creature"
	"aeons/internal/location"
	"log"
)

const plantName = "Suspicious Plant"

func (gs *GameState) spawn(creatureTemplate creature.Unit, at *location.Unit) *creature.Unit {
	c := creatureTemplate
	c.ID = gs.NextID()
	c.CurrentLocation = at
	c.Home = at
	c.BondThreshold = (c.Size + 1) / 2
	log.Printf("[INIT] Spawned creature %s at %s", c.NameWithID(), at.Name)
	return &c
}

func (gs *GameState) NewWolfpack(at *location.Unit) *creature.Unit {
	return gs.spawn(creature.Unit{
		Name:        "Wolfpack",
		IdentNames:  []string{"Wolf", "Collective", "Animal"},
		Flavor:      "Run with the wolves.",
		Size:        6,
		Role:        creature.Predator,
		MaxDistress: 5,
		Radius:      3,
	}, at)
}

func (gs *GameState) NewMoose(at *location.Unit) *creature.Unit {
	return gs.spawn(creature.Unit{
		Name:        "Moose",
		IdentNames:  []string{"Moose", "Animal", "Grazer"},
		Flavor:      "In solitude it wanders.",
		Size:        7,
		Role:        creature.Prey,
		MaxDistress: 8,
		Radius:      2,
	}, at)
}

func (gs *GameState) NewStag(at *location.Unit) *creature.Unit {
	return gs.spawn(creature.Unit{
		Name:        "Stag",
		IdentNames:  []string{"Stag", "Animal", "Grazer"},
		Flavor:      "Gentle and with purpose.",
		Size:        6,
		Role:        creature.Prey,
		MaxDistress: 5,
		Radius:      3,
	}, at)
}

func (gs *GameState) NewRaccoon(at *location.Unit) *creature.Unit {
	return gs.spawn(creature.Unit{
		Name:        "Raccoon",
		IdentNames:  []string{"Raccoon", "Animal"},
		Flavor:      "A nifty little raccoon, known to steal berries. Especially yours.",
		Size:        3,
		Role:        creature.Prey,
		MaxDistress: 3,
		Radius:      2,
		CanHide:     true,
	}, at)
}

func (gs *GameState) NewHare(at *location.Unit) *creature.Unit {
	return gs.spawn(creature.Unit{
		Name:        "Hare",
		IdentNames:  []string{"Hare", "Critter", "Animal"},
		Flavor:      "The warren is empty.",
		Size:        2,
		Role:        creature.Prey,
		MaxDistress: 2,
		Radius:      3,
		CanHide:     true,
	}, at)
}

func (gs *GameState) NewFox(at *location.Unit) *creature.Unit {
	return gs.spawn(creature.Unit{
		Name:        "Fox",
		IdentNames:  []string{"Fox", "Animal"},
		Flavor:      "A fox does what a fox must do.",
		Size:        4,
		Role:        creature.PredatorPrey,
		MaxDistress: 3,
		Radius:      2,
	}, at)
}

func (gs *GameState) NewSuspiciousPlant(at *location.Unit) *creature.Unit {
	return gs.spawn(creature.Unit{
		Name:        plantName,
		IdentNames:  []string{"Plant", "Suspicious"},
		Flavor:      "Something peculiar...",
		Size:        3,
		Role:        creature.Plant,
		MaxDistress: 6,
	}, at)
}

func (gs *GameState) NewWatchfulMycelium(at *location.Unit) *creature.Unit {
	return gs.spawn(creature.Unit{
		Name:        "Watchful Mycelium",
		IdentNames:  []string{"Plant", "Fungus", "Collective"},
		Flavor:      "Eyes in the trees...",
		Size:        1,
		Role:        creature.Plant,
		MaxDistress: 2,
	}, at)
}
