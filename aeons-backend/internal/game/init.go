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

	log.Println("[INIT] Setting up locations....")
	SecludedDen := &location.Unit{ID: 1, Name: "Secluded Den"}
	WindsweptPlains := &location.Unit{ID: 1, Name: "Windswept Plains"}
	RedhornLake := &location.Unit{ID: 1, Name: "Redhorn Lake"}
	ConiferousGrove := &location.Unit{ID: 1, Name: "Coniferous Grove"}
	AridPlateau := &location.Unit{ID: 1, Name: "Arid Plateau"}
	SlumberingCrag := &location.Unit{ID: 1, Name: "Slumbering Crag"}
	HermitsRecluse := &location.Unit{ID: 1, Name: "Hermit's Recluse"}

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
	grabBag := append([]*location.Unit(nil), gs.Locations...)
	for range 5 {
		r := rand.IntN(len(grabBag))
		randomLocation := grabBag[r]
		grabBag = slices.Delete(grabBag, r, r+1)

		h := harvesting.GenerateRandomPredefinedHarvestingUnit(randomLocation)
		gs.HarvestUnits = append(gs.HarvestUnits, h)
		log.Printf("[INIT] Spawning harvestable %s at %s\n", h.Name, randomLocation.Name)
	}
}

func (gs *GameState) InitPlayers() {
	log.Println("[INIT] Setting up player Druid...")
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

	log.Println("[INIT] Setting up player Scoundrel...")
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

func (gs *GameState) getRandomLocation() *location.Unit {
	return gs.Locations[rand.IntN(len(gs.Locations))]
}

func (gs *GameState) InitCreatures() {

	Stag := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: gs.getRandomLocation()},
		HealthPool:   combat.HealthPool{CurrentHealth: 5, MaxHealth: 5},
		ID:           1,
		Name:         "Stag",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(true),
			CreatureMovementConfig:  creature.GenerateCuriousConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
		},
		Damage: 2,
	}

	log.Printf("[INIT] Spawned creature %s at %s", Stag.Name, Stag.CurrentLocation.Name)

	Racoon := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: gs.getRandomLocation()},
		HealthPool:   combat.HealthPool{CurrentHealth: 2, MaxHealth: 2},
		ID:           2,
		Name:         "Raccoon",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(false),
			CreatureMovementConfig:  creature.GenerateCuriousConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
		},
		Damage: 1,
	}

	log.Printf("[INIT] Spawned creature %s at %s", Racoon.Name, Racoon.CurrentLocation.Name)

	Plant := &creature.Unit{
		MovingEntity: moving.MovingEntity{CurrentLocation: gs.getRandomLocation()},
		HealthPool:   combat.HealthPool{CurrentHealth: 2, MaxHealth: 2},
		ID:           1,
		Name:         "Suspicious Plant",
		CreatureBehaviourConfig: &creature.CreatureBehaviourConfig{
			CreatureAloofConfig:     creature.GenerateAloofConfig(true),
			CreatureMovementConfig:  creature.GenerateShyConfig(),
			CreatureTurnStateConfig: creature.GenerateDefaultTurnStateConfig(),
		},
		Damage: 1,
	}

	log.Printf("[INIT] Spawned creature %s at %s", Plant.Name, Plant.CurrentLocation.Name)

	gs.Creatures = append(gs.Creatures, Stag, Racoon, Plant)
}
