package game

import (
	"aeons/internal/card"
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/harvesting"
	"aeons/internal/location"
	"aeons/internal/player"
	"fmt"
)

func (gs *GameState) Logf(format string, args ...any) {
	gs.eventLog = append(gs.eventLog, fmt.Sprintf(format, args...))
}

// DrainEventLog returns everything logged since the last drain.
func (gs *GameState) DrainEventLog() []string {
	entries := gs.eventLog
	gs.eventLog = nil
	return entries
}

func (gs *GameState) logPlayedCard(c *card.PlayableCard, owner *player.Unit) {
	switch target := c.Target.(type) {
	case *location.Unit:
		gs.Logf("%s moves to %s", owner.Name, target.Name)
	case *creature.Unit:
		gs.logPlayedOnCreature(c, owner, target)
	case *harvesting.Unit:
		if harvest, ok := c.Effect.(*harvesting.Effect); ok && harvest.Yield != nil {
			gs.Logf("%s harvests %s:", owner.Name, target.Name)
			gs.Logf("      Yield: %d nutrition, %d resources", harvest.Yield.Nutrition, harvest.Yield.Resources)
		}
	}
}

func (gs *GameState) logPlayedOnCreature(c *card.PlayableCard, owner *player.Unit, target *creature.Unit) {
	if c.Type == card.PlayerDistract {
		gs.Logf("%s distracts %s", owner.Name, target.Name)
		return
	}
	if target.CurrentHealth == 0 {
		gs.Logf("%s attacks %s for %d, %s is defeated", owner.Name, target.Name, combat.BaseDamage, target.Name)
		return
	}
	gs.Logf("%s attacks %s for %d (%d/%d)", owner.Name, target.Name, combat.BaseDamage, target.CurrentHealth, target.MaxHealth)
}
