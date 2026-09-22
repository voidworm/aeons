package creature

import (
	"aeons/internal/combat"
	"aeons/internal/moving"
)

type Unit struct {
	ID   int
	Name string
	moving.MovingEntity
	combat.HealthPool
	*CreatureBehaviourConfig
	Damage int
}

type CreatureBehaviourConfig struct {
	*CreatureAloofConfig
	*CreatureMovementConfig
	*CreatureTurnStateConfig
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

func GenerateIdleConfig() *CreatureMovementConfig {
	return &CreatureMovementConfig{
		Shy:     false,
		Curious: false,
	}
}

func GenerateDefaultTurnStateConfig() *CreatureTurnStateConfig {
	return &CreatureTurnStateConfig{
		Enraged:   false,
		Exhausted: false,
	}
}

type CreatureAloofConfig struct {
	Aloof bool // aloof creatures at locations do not prevent player movement and will not attack players in the creature phase
}

type CreatureMovementConfig struct {
	Curious bool //curious creatures will move towards a player in the creature phase
	Shy     bool //shy creatures will move away from a location if a player is there in there creature phase
}

type CreatureTurnStateConfig struct {
	Enraged   bool //creatures enrage on damage, overwrites aloof for the turn
	Exhausted bool //creatures exhaust on evade or attack, putting them to functional aloof for the turn
}

func (u *Unit) Tick() {
	u.Enraged = false
	u.Exhausted = false
}

func (u *Unit) TakeDamage(amount int) int {
	u.Enraged = true
	u.CurrentHealth = max(u.CurrentHealth-amount, 0)
	return u.CurrentHealth
}

func (ee *Unit) OutgoingDamage() int {
	return ee.Damage
}

func (u *Unit) IsEvadable() bool {

	if u.Exhausted {
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
