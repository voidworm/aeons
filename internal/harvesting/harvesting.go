package harvesting

import (
	"aeons/internal/location"
	"aeons/internal/player"
	"fmt"
	"log"
)

// harvestable interface
type Harvestable interface {
	HarvestedBy(*player.Unit)
	GenerateYield() *Yield
}

// basic harvesting unit
type Unit struct {
	Name                string
	StaticLocation      *location.LocationEntity
	HealYieldConfig     *YieldConfig
	ResourceYieldConfig *YieldConfig
}

func (u *Unit) HarvestedBy(player *player.Unit) {
	yield := u.GenerateYield()
	player.ResourcesAvailable += yield.Resources
	player.CurrentHealth += 5
	log.Printf("%s dropped loot with %d healing and %d resources\n", u.Name, yield.Healing, yield.Resources)
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

func (u *Unit) HandleTick() {
	u.HealYieldConfig.HandleTick()
	u.ResourceYieldConfig.HandleTick()
}

// config for yields
type YieldConfig struct {
	MaxCapacity     int
	CurrentCapacity int
	Amount          int
	RespawnTicks    int
	CurrentTicks    int
}

func (yc *YieldConfig) HandleTick() {
	yc.CurrentTicks++
	if yc.CurrentTicks == yc.RespawnTicks {
		yc.CurrentCapacity = yc.MaxCapacity
	}
}

func (dc *YieldConfig) Drop() int {
	dropAmount := dc.AvailableAmount()
	dc.CurrentCapacity -= dropAmount
	return dropAmount
}

func (dc *YieldConfig) AvailableAmount() int {
	return max(dc.CurrentCapacity, dc.Amount)
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
	e.TargetHarvestable.HarvestedBy(e.Player)
}
