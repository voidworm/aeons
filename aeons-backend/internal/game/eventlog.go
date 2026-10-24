package game

import (
	"aeons/internal/card"
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/collecting"
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
		if c.Type == card.PlayerMoveKindred {
			gs.Logf("Kindred %s moves to %s", gs.KindredOf(owner).NameWithID(), target.Name)
			return
		}
		gs.Logf("%s moves to %s", owner.Name, target.Name)
	case *creature.Unit:
		gs.logPlayedOnCreature(c, owner, target)
	case *collecting.Node:
		if gathered, ok := c.Effect.(*collecting.Effect); ok {
			gs.Logf("%s %s %s", owner.Name, gathered.Mode.Verb(), target.Name)
		}
	}
}

func (gs *GameState) logPlayedOnCreature(c *card.PlayableCard, owner *player.Unit, target *creature.Unit) {
	if c.Type == card.PlayerDistract {
		gs.Logf("%s distracts %s", owner.Name, target.NameWithID())
		return
	}
	if c.Type == card.PlayerBond {
		gs.Logf("%s bonds with %s (%d/%d)", owner.Name, target.NameWithID(), target.BondCounter, target.BondThreshold)
		if target.Bonded {
			gs.Logf("%s is now kindred to %s", target.NameWithID(), owner.Name)
		}
		return
	}
	gs.Logf("%s attacks %s for %d (%d/%d)", owner.Name, target.NameWithID(), combat.BaseDamage, target.CurrentHealth, target.MaxHealth)
	if target.CurrentHealth == 0 {
		gs.Logf("%s is defeated", target.NameWithID())
	}
}
