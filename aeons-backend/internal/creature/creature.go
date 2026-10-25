package creature

import (
	"aeons/internal/card"
	"aeons/internal/combat"
	"aeons/internal/moving"
	"aeons/internal/player"
	"fmt"
)

type Unit struct {
	ID   int
	Name string
	Flavor string
	moving.MovingEntity
	combat.HealthPool
	*CreatureBehaviourConfig
}

type CreatureBehaviourConfig struct {
	*CreatureAloofConfig
	*CreatureMovementConfig
	*CreatureTurnStateConfig
	*CreatureBondConfig
	*CreatureHuntingConfig
}

func GenerateHuntingConfig(aggressive bool, prey bool) *CreatureHuntingConfig {
	return &CreatureHuntingConfig{
		Aggressive: aggressive,
		Prey:       prey,
	}
}

func GenerateBondConfig(threshold int) *CreatureBondConfig {
	return &CreatureBondConfig{
		BondThreshold: threshold,
	}
}

func GenerateAloofConfig(on bool) *CreatureAloofConfig {
	return &CreatureAloofConfig{
		Aloof: on,
	}
}

func GenerateShyConfig() *CreatureMovementConfig {
	return &CreatureMovementConfig{
		Shy:     true,
		Curious: false,
	}
}

func GenerateCuriousConfig() *CreatureMovementConfig {
	return &CreatureMovementConfig{
		Shy:     false,
		Curious: true,
	}
}

func GenerateCuriousShyConfig() *CreatureMovementConfig {
	return &CreatureMovementConfig{
		Shy:     true,
		Curious: true,
	}
}

func GenerateDefaultTurnStateConfig() *CreatureTurnStateConfig {
	return &CreatureTurnStateConfig{
		Enraged:   false,
		Distracted: false,
	}
}

type CreatureAloofConfig struct {
	Aloof bool // aloof creatures at locations do not prevent player movement
}

type CreatureHuntingConfig struct {
	Aggressive bool // aggressive creatures attack players and prey at their location in the world phase, aggressive and curious makes a hunter
	Prey       bool // prey is hunted and attacked by aggressive creatures, attacked prey flees, or retaliates if it is aggressive itself
}

type CreatureMovementConfig struct {
	Curious bool //curious creatures will move towards the closest player or creature in the world phase, hunters only towards prey and players
	Shy     bool //shy creatures will move away from a location if a player or an aggressive creature is there in their world phase
}

type CreatureTurnStateConfig struct {
	Enraged   bool //creatures enrage on damage, overwrites aloof for the turn
	Distracted bool //creatures exhaust on evade or attack, putting them to functional aloof for the turn
}

type CreatureBondConfig struct {
	Bonded        bool // bonded creatures do not attack, do not prevent player movement and cannot be attacked or distracted
	BondedTo      *player.Unit // the player that pushed the bond counter to the threshold, the creature is their kindred
	BondCounter   int
	BondThreshold int // the bond counter at which the creature becomes bonded
}

func (u *Unit) Bond(by *player.Unit) {
	u.BondCounter++
	if u.BondCounter >= u.BondThreshold {
		u.Bonded = true
		u.BondedTo = by
	}
}

func (u *Unit) IsBondable() bool {
	return !u.Bonded && (u.Aloof || u.Distracted)
}

func (u *Unit) Tick() {
	u.Enraged = false
	u.Distracted = false
}

// only damage dealt by players wears down the bond progress, hunters do not
type PlayerAttackEffect struct {
	Target *Unit
}

func (e *PlayerAttackEffect) Apply() {
	if !e.Target.Bonded {
		e.Target.BondCounter = 0
	}
	damage := combat.DamageEffect{Target: e.Target}
	damage.Apply()
}

func (u *Unit) TakeDamage(amount int) int {
	u.Enraged = true
	u.CurrentHealth = max(u.CurrentHealth-amount, 0)
	return u.CurrentHealth
}

func (u *Unit) IsEvadable() bool {

	if u.Bonded {
		return false
	}

	if u.Distracted {
		return false
	}

	if u.Enraged {
		return true
	}

	if u.Aloof {
		return false
	}

	return true
}

func (u *Unit) GetTargetType() card.TargetType {
	return card.TargetTypePlayer
}

func (u *Unit) NameWithID() string {
	return fmt.Sprintf("%s (%d)", u.Name, u.ID)
}

func (u *Unit) GetID() int {
	return u.ID
}
