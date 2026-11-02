package game

import (
	"aeons/internal/creature"
	"aeons/internal/effect"
	"aeons/internal/location"
	"aeons/internal/quest"
	"fmt"
	"math/rand/v2"
	"slices"
)

const plantName = "Suspicious Plant"

// Spawns a creature when the quest it belongs to starts and remembers it for the goals.
type spawnCreature struct {
	gs      *GameState
	create  func(at *location.Unit) *creature.Unit
	where   quest.Place
	spawned *creature.Unit
}

func (s *spawnCreature) Apply() {
	s.spawned = s.create(s.where())
	s.gs.Creatures = append(s.gs.Creatures, s.spawned)
	s.gs.Logf("%s appears", s.spawned.NameWithID())
}

func (s *spawnCreature) Creature() *creature.Unit {
	return s.spawned
}

func (gs *GameState) locationNamed(name string) *location.Unit {
	i := slices.IndexFunc(gs.Locations, func(l *location.Unit) bool { return l.Name == name })
	if i == -1 {
		panic(fmt.Sprintf("no location named %s", name))
	}
	return gs.Locations[i]
}

func randomLocationMovesFrom(from *location.Unit, moves int, among []*location.Unit) quest.Place {
	return func() *location.Unit {
		candidates := slices.DeleteFunc(slices.Clone(among), func(l *location.Unit) bool {
			return from.DistanceTo(l)-1 != moves
		})
		return candidates[rand.IntN(len(candidates))]
	}
}

func (gs *GameState) InitScenario() {
	const scenarioName = "A Journey"
	const elementsName = "Awaken the Elements"

	hermitsRecluse := quest.PlaceAt(gs.locationNamed("Hermit's Recluse"))
	redhornLake := quest.PlaceAt(gs.locationNamed("Redhorn Lake"))
	coniferousGrove := quest.PlaceAt(gs.locationNamed("Coniferous Grove"))
	windsweptPlains := quest.PlaceAt(gs.locationNamed("Windswept Plains"))

	beyondTheBody := &quest.Quest{
		ID:        gs.NextID(),
		Name:      "Beyond the Body",
		Flavor:    "I hope that you remember home.",
		Goals:     []*quest.Goal{quest.AllPlayersHaveKindred("Each player has a kindred")},
	}

	openTheEye := &quest.Quest{
		ID:        gs.NextID(),
		Name:      "Open the Eye",
		Flavor:    "The wind alone...",
		Mandatory: true,
		Goals: []*quest.Goal{
			quest.AllPlayersAt(hermitsRecluse, quest.Fixed("Each player is at Hermit's Recluse")),
			quest.OnlyBondedCreaturesWithPlayers("There's no creature but bonded creatures at your location."),
		},
	}

	aFloodIsComing := &quest.Quest{
		ID:        gs.NextID(),
		Name:      "A Flood is Coming",
		Flavor:    "May you awaken with the fjords.",
		Goals:     []*quest.Goal{quest.CanSpend(redhornLake, 10, "The players at Redhorn Lake can spend 10 resources.")},
	}

	plantSpawn := &spawnCreature{
		gs:     gs,
		create: gs.NewPlant,
		where:  randomLocationMovesFrom(coniferousGrove(), 2, gs.Locations),
	}
	becomeOne := &quest.Quest{
		ID:        gs.NextID(),
		Name:      "Become One",
		Flavor:    "May you bond with the copse.",
		OnStart:   []effect.Effect{plantSpawn},
		Goals: []*quest.Goal{
			quest.AnyPlayerAt(coniferousGrove, "1+ player is at Coniferous Grove"),
			quest.CreaturesAt(coniferousGrove, 1, func(c *creature.Unit) bool { return c.Name == plantName }, "A Suspicious Plant is at Coniferous Grove"),
		},
	}

	acrossTheEternalPlains := &quest.Quest{
		ID:        gs.NextID(),
		Name:      "Across the Eternal Plains",
		Flavor:    "May you wander with the tundra.",
		Goals: []*quest.Goal{
			quest.AnyPlayerAt(windsweptPlains, "1+ player is at Windswept Plains"),
			quest.CreaturesAt(windsweptPlains, 2, func(c *creature.Unit) bool { return c.Distracted }, "2+ distracted creatures at Windswept Plains"),
		},
	}

	awakenTheElements := &quest.Group{
		Name:       elementsName,
		Items:      []quest.Item{aFloodIsComing, becomeOne, acrossTheEternalPlains},
		Concurrent: true,
		Required:   2,
	}

	// the element that was not awakened is where the journey ends
	endOfTheJourney := func() *location.Unit {
		switch {
		case aFloodIsComing.Status != quest.StatusFinished:
			return redhornLake()
		case becomeOne.Status != quest.StatusFinished:
			return coniferousGrove()
		default:
			return windsweptPlains()
		}
	}

	toUnderstandEventually := &quest.Quest{
		ID:        gs.NextID(),
		Name:      "To understand eventually...",
		Flavor:    "This land was in great pain.",
		Goals: []*quest.Goal{
			quest.AllPlayersAndBondedCreaturesAt(endOfTheJourney, func() string {
				return "All players and bonded creatures are at " + endOfTheJourney().Name
			}),
		},
	}

	spiritSpawn := &spawnCreature{
		gs: gs,
		create: func(at *location.Unit) *creature.Unit {
			spirit := gs.NewStag(at)
			spirit.Name = "Enraged Spirit"
			spirit.Flavor = "Lost in the eventide."
			return spirit
		},
		where: endOfTheJourney,
	}
	atLongLastHomebound := &quest.Quest{
		ID:        gs.NextID(),
		Name:      "At long last, homebound.",
		Flavor:    "We leave...",
		Mandatory: true,
		OnStart:   []effect.Effect{spiritSpawn},
		Goals:     []*quest.Goal{quest.CreatureDefeated(spiritSpawn.Creature, "Enraged Spirit is defeated")},
	}

	gs.Scenario = &quest.Group{
		Name:  scenarioName,
		Items: []quest.Item{beyondTheBody, openTheEye, awakenTheElements, toUnderstandEventually, atLongLastHomebound},
	}
	gs.Scenario.Begin()
}
