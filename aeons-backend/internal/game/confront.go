package game

import (
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/player"
)

// confrontEffect distresses the creature. Prey that survives flees, at its burrow it hides because it is distressed now.
type confrontEffect struct {
	gs     *GameState
	Player *player.Unit
	Target *creature.Unit
}

func (e *confrontEffect) Apply() {
	if !e.Target.Bonded {
		e.Target.BondCounter = 0
	}
	damage := combat.DamageEffect{Target: e.Target}
	damage.Apply()
	e.gs.Logf("%s confronts %s (%d/%d distress)", e.Player.Name, e.Target.NameWithID(), e.Target.Distress, e.Target.MaxDistress)

	if !e.Target.IsDefeated() && e.Target.Role.IsPrey() {
		e.gs.Flee(e.Target)
	}
}
