package game

import (
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/harvesting"
	"aeons/internal/location"
	"aeons/internal/moving"
	"aeons/internal/player"
	"log"
	"math/rand/v2"
	"slices"
)

func (gs *GameState) Init() {

	gs.InitLocations()
	gs.InitHarvestingUnits()
	gs.InitPlayers()
	gs.InitCreatures()
}

func (gs *GameState) InitLocations() {

	log.Println("[LOCATIONS] Setting up Locations....")
	SecludedDen := &location.LocationEntity{ID: 1, Name: "Secluded Den"}
	WindsweptPlains := &location.LocationEntity{ID: 1, Name: "Windswept Plains"}
	RedhornLake := &location.LocationEntity{ID: 1, Name: "Redhorn Lake"}
	ConiferousGrove := &location.LocationEntity{ID: 1, Name: "Coniferous Grove"}
	AridPlateau := &location.LocationEntity{ID: 1, Name: "Arid Plateau"}
	SlumberingCrag := &location.LocationEntity{ID: 1, Name: "Slumbering Crag"}
	HermitsRecluse := &location.LocationEntity{ID: 1, Name: "Hermit's Recluse"}

	SecludedDen.OutgoingConnections = append(SecludedDen.OutgoingConnections, WindsweptPlains)
	WindsweptPlains.OutgoingConnections = append(WindsweptPlains.OutgoingConnections, SecludedDen, RedhornLake, ConiferousGrove, AridPlateau)
	RedhornLake.OutgoingConnections = append(RedhornLake.OutgoingConnections, WindsweptPlains, ConiferousGrove)
	ConiferousGrove.OutgoingConnections = append(ConiferousGrove.OutgoingConnections, WindsweptPlains, RedhornLake)
	AridPlateau.OutgoingConnections = append(AridPlateau.OutgoingConnections, WindsweptPlains, SlumberingCrag, HermitsRecluse)
	SlumberingCrag.OutgoingConnections = append(SlumberingCrag.OutgoingConnections, AridPlateau, SecludedDen)
	HermitsRecluse.OutgoingConnections = append(HermitsRecluse.OutgoingConnections, AridPlateau)

	gs.Locations = append(gs.Locations, SecludedDen, WindsweptPlains, RedhornLake, ConiferousGrove, AridPlateau, SlumberingCrag, HermitsRecluse)
}

func (gs *GameState) InitHarvestingUnits() {
	grabBag := append([]*location.LocationEntity(nil), gs.Locations...)
	for range 5 {
		r := rand.IntN(len(grabBag))
		randomLocation := grabBag[r]
		grabBag = slices.Delete(grabBag, r, r+1)

		h := harvesting.GenerateRandomPredefinedHarvestingUnit(randomLocation)
		gs.HarvestUnits = append(gs.HarvestUnits, h)
		log.Printf("[HARVESTING] Spwaning %s at %s\n", h.Name, randomLocation.Name)
	}
}

func (gs *GameState) InitPlayers() {
	log.Println("[PLAYERS] Setting up Druid...")
	Druid := &player.Unit{
		MovingEntity:       moving.MovingEntity{CurrentLocation: gs.Locations[0]},
		HealthPool:         combat.HealthPool{CurrentHealth: 10, MaxHealth: 10},
		ID:                 1,
		Name:               "Druid",
		CardsInHand:        7,
		ResourcesAvailable: 5,
		RemainingActions:   3,
		Damage:             1,
	}

	log.Println("[PLAYERS] Setting up Scoundrel...")
	Scoundrel := &player.Unit{
		MovingEntity:       moving.MovingEntity{CurrentLocation: gs.Locations[0]},
		HealthPool:         combat.HealthPool{CurrentHealth: 8, MaxHealth: 8},
		ID:                 1,
		Name:               "Scoundrel",
		CardsInHand:        7,
		ResourcesAvailable: 5,
		RemainingActions:   3,
		Damage:             1,
	}

	gs.Players = append(gs.Players, Druid, Scoundrel)
}

func (gs *GameState) getRandomLocation() *location.LocationEntity {
	return gs.Locations[rand.IntN(len(gs.Locations))]
}

func (gs *GameState) InitCreatures() {

	Stag := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: gs.getRandomLocation()},
		HealthPool:   combat.HealthPool{CurrentHealth: 5, MaxHealth: 5},
		ID:           1,
		Name:         "Stag",
		Aloof:        true,
		Hunter:       true,
		Damage:       2,
	}

	log.Printf("[CREATURES] Spawned %s at %s", Stag.Name, Stag.CurrentLocation.Name)

	Racoon := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: gs.getRandomLocation()},
		HealthPool:   combat.HealthPool{CurrentHealth: 2, MaxHealth: 2},
		ID:           2,
		Name:         "Raccoon",
		Aloof:        false,
		Hunter:       false,
		Damage:       1,
	}

	log.Printf("[CREATURES] Spawned %s at %s", Racoon.Name, Racoon.CurrentLocation.Name)

	Plant := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: gs.getRandomLocation()},
		HealthPool:   combat.HealthPool{CurrentHealth: 2, MaxHealth: 2},
		ID:           1,
		Name:         "Suspicious Plant",
		Aloof:        true,
		Hunter:       false,
		Damage:       1,
	}

	log.Printf("[CREATURES] Spawned %s at %s", Plant.Name, Plant.CurrentLocation.Name)

	gs.Creatures = append(gs.Creatures, Stag, Racoon, Plant)
}
