package harvesting

import (
	"aeons/internal/card"
	"aeons/internal/location"
	"aeons/internal/player"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
)

// harvestable interface
type Harvestable interface {
	HarvestBy(*player.Unit)
	GenerateYield() *Yield
}

// basic harvesting unit
type Unit struct {
	ID	int
	Name                string
	Flavor string
	StaticLocation      *location.Unit
	HealYieldConfig     *YieldConfig
	ResourceYieldConfig *YieldConfig
}

type GenerateFunc func(id int, location *location.Unit) *Unit

func GenerateRandomPredefinedHarvestingUnit(id int, location *location.Unit) *Unit {
	generateFuncs := []GenerateFunc{
		GenerateBerryBush,
		GenerateBrush,
		GenerateMushroomField,
		GenerateLushCopse,
		GenerateFruitTree,
	}
	f := generateFuncs[rand.IntN(len(generateFuncs))]
	return f(id, location)
}

func GenerateBerryBush(id int, location *location.Unit) *Unit {
	return &Unit{
		ID:				id,
		Name:                "Berry Bush",
		Flavor: "The berries around here are very sweet. Raccoons in particular like that.",
		StaticLocation:      location,
		HealYieldConfig:     GenerateSmallYieldConfig(),
		ResourceYieldConfig: GenerateEmptyYieldConfig(),
	}
}

func GenerateBrush(id int, location *location.Unit) *Unit {
	return &Unit{
		ID:	id,
		Name:                "Brush",
		Flavor: "Thorns and sticks. You'll find a way to use these.",
		StaticLocation:      location,
		HealYieldConfig:     GenerateEmptyYieldConfig(),
		ResourceYieldConfig: GenerateSmallYieldConfig(),
	}
}

func GenerateMushroomField(id int, location *location.Unit) *Unit {
	return &Unit{
		ID: id,
		Name:                "Mushroom Field",
		Flavor: "A multitude of fungi growing in a large patch.",
		StaticLocation:      location,
		HealYieldConfig:     GenerateLargeYieldConfig(),
		ResourceYieldConfig: GenerateSmallYieldConfig(),
	}
}

func GenerateLushCopse(id int, location *location.Unit) *Unit {
	return &Unit{
		ID: id,
		Name:                "Lush Copse",
		Flavor: "This copse is home to life of many sorts, not all of it edible, but vivid.",
		StaticLocation:      location,
		HealYieldConfig:     GenerateSmallYieldConfig(),
		ResourceYieldConfig: GenerateLargeYieldConfig(),
	}
}

func GenerateFruitTree(id int, location *location.Unit) *Unit {
	return &Unit{
		ID: id,
		Name:                "Fruit Tree",
		Flavor:			"An almost impossible to exhaust source of both fruit and branches.",
		StaticLocation:      location,
		HealYieldConfig:     GenerateLargeYieldConfig(),
		ResourceYieldConfig: GenerateLargeYieldConfig(),
	}
}

func (u *Unit) HarvestBy(player *player.Unit) {
	yield := u.GenerateYield()
	player.ResourcesAvailable += yield.Resources
	player.CurrentHealth += yield.Healing
	player.CurrentHealth = min(player.CurrentHealth, player.MaxHealth)
	log.Printf("[HARVESTING] %s was harvested by %s for %d healing and %d resources\n", u.Name, player.Name, yield.Healing, yield.Resources)
}

func (u *Unit) GenerateYield() *Yield {
	name := fmt.Sprintf("%s Drop", u.Name)
	valuesForDrop := make([]int, 0, 2)

	configs := [2]*YieldConfig{u.HealYieldConfig, u.ResourceYieldConfig}

	for _, v := range configs {
		valuesForDrop = append(valuesForDrop, v.Drop())
	}

	yield := &Yield{
		Name:      name,
		Healing:   valuesForDrop[0],
		Resources: valuesForDrop[1],
	}
	return yield
}

func (u *Unit) Tick() {
	u.HealYieldConfig.Tick()
	u.ResourceYieldConfig.Tick()
}

// config for yields
type YieldConfig struct {
	MaxCapacity     int
	CurrentCapacity int
	YieldAmount     int
	RespawnTicks    int
	CurrentTicks    int
}

func GenerateEmptyYieldConfig() *YieldConfig {
	return &YieldConfig{
		MaxCapacity:     0,
		CurrentCapacity: 0,
		YieldAmount:     0,
		RespawnTicks:    math.MaxInt,
		CurrentTicks:    0,
	}
}
func GenerateSmallYieldConfig() *YieldConfig {
	return &YieldConfig{
		MaxCapacity:     3,
		CurrentCapacity: 3,
		YieldAmount:     1,
		RespawnTicks:    3,
		CurrentTicks:    1,
	}
}

func GenerateMediumYieldConfig() *YieldConfig {
	return &YieldConfig{
		MaxCapacity:     5,
		CurrentCapacity: 5,
		YieldAmount:     1,
		RespawnTicks:    3,
		CurrentTicks:    1,
	}
}

func GenerateLargeYieldConfig() *YieldConfig {
	return &YieldConfig{
		MaxCapacity:     10,
		CurrentCapacity: 10,
		YieldAmount:     2,
		RespawnTicks:    5,
		CurrentTicks:    0,
	}
}

func (yc *YieldConfig) Tick() {
	yc.CurrentTicks++
	if yc.CurrentTicks == yc.RespawnTicks {
		yc.CurrentCapacity += yc.YieldAmount
		yc.CurrentTicks = 0
	}
}

func (yc *YieldConfig) Drop() int {
	dropAmount := yc.AvailableAmount()
	yc.CurrentCapacity -= dropAmount
	return dropAmount
}

func (yc *YieldConfig) AvailableAmount() int {
	return min(yc.CurrentCapacity, yc.YieldAmount)
}

// basic yield
type Yield struct {
	Name      string
	Healing   int
	Resources int
}

//effect and effect apply

type Effect struct {
	Player            *player.Unit
	TargetHarvestable Harvestable
}

func (e *Effect) Apply() {
	e.TargetHarvestable.HarvestBy(e.Player)
}


func (u *Unit) GetTargetType() card.TargetType {
	return card.TargetTypeHarvest
}

func (u *Unit) GetID() int {
	return u.ID
}
