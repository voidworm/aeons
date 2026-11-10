package game

import (
	"aeons/internal/combat"
	"aeons/internal/creature"
	"aeons/internal/player"
	"cmp"
	"math/rand/v2"
)

// HuntableAt lists everything the hunter can attack at its location, players included.
func (gs *GameState) HuntableAt(hunter *creature.Unit) []creature.Quarry {
	found := []creature.Quarry{}
	for _, c := range gs.EnemiesAtLocation(hunter.CurrentLocation) {
		if c != hunter && gs.canHunt(hunter, c) {
			found = append(found, c)
		}
	}
	for _, p := range gs.PlayersAtLocation(hunter) {
		if gs.canHunt(hunter, p) {
			found = append(found, p)
		}
	}
	return found
}

// the lowest remaining distress first, then the smaller size, then at random
func pickTarget(targets []creature.Quarry) creature.Quarry {
	best := []creature.Quarry{targets[0]}
	for _, target := range targets[1:] {
		switch compareTargets(target, best[0]) {
		case -1:
			best = []creature.Quarry{target}
		case 0:
			best = append(best, target)
		}
	}
	return best[rand.IntN(len(best))]
}

func compareTargets(a, b creature.Quarry) int {
	return cmp.Or(
		cmp.Compare(a.RemainingDistress(), b.RemainingDistress()),
		cmp.Compare(a.GetSize(), b.GetSize()),
	)
}

// returns the attack if it wounded a player
func (gs *GameState) Attack(hunter *creature.Unit, target creature.Quarry) (CreatureAttack, bool) {
	effect := combat.DamageEffect{Target: target}
	effect.Apply()
	switch t := target.(type) {
	case *player.Unit:
		gs.logCreatureAttack(hunter.NameWithID(), t)
		return CreatureAttack{TargetPlayerID: t.ID, Fainted: t.Fainted()}, true
	case *creature.Unit:
		gs.logCreatureHit("attacks", hunter, t)
	}
	return CreatureAttack{}, false
}
