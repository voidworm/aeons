package creature

import (
	"aeons/internal/card"
	"aeons/internal/combat"
	"aeons/internal/location"
	"aeons/internal/moving"
	"aeons/internal/player"
	"fmt"
	"strings"
)

type Role int

const (
	Prey Role = iota
	Predator
	PredatorPrey
	Plant // plants are neither hunters nor hunted
)

func (r Role) IsPredator() bool {
	return r == Predator || r == PredatorPrey
}

func (r Role) IsPrey() bool {
	return r == Prey || r == PredatorPrey
}

func (r Role) IsPlant() bool {
	return r == Plant
}

// tags that decide what a creature eats, they are part of its IdentNames
const (
	Scavenger = "Scavenger"
	Grazer    = "Grazer"
	Gatherer  = "Gatherer"
)

type Unit struct {
	ID         int
	Name       string
	IdentNames []string // what cards address the creature as, so "each Hare" finds a hare named Tim
	Flavor     string
	Size       int // 1 to 10
	Role       Role
	moving.MovingEntity

	Distress    int
	MaxDistress int // the distress at which the creature is defeated

	Home    *location.Unit // where it spawned, creatures that can hide hide here
	Radius  int            // how many moves away from home it likes to wander
	CanHide bool
	Hidden  bool // hidden creatures cannot be hunted or targeted
	HiddenInSight bool // the players saw the creature hide, only then do they see it while it is hidden

	Exhausted    bool
	LastLocation *location.Unit
	Chasing      Quarry // the prey a hunter locked on to

	Bonded        bool         // bonded creatures do not act on their own and cannot be targeted by players
	BondedTo      *player.Unit // the player that pushed the bond counter to the threshold, the creature is their kindred
	BondCounter   int
	BondThreshold int // the bond counter at which the creature becomes bonded
}

// Quarry is what a hunter can hunt: prey and players
type Quarry interface {
	combat.Damageable
	Location() *location.Unit
	GetSize() int
	RemainingDistress() int
	IsHuntable() bool
}

// predators hunt prey that is smaller, predators that are no prey themselves also prey of the same size
func (u *Unit) CanHunt(target Quarry) bool {
	if !u.Role.IsPredator() || !target.IsHuntable() {
		return false
	}
	if target.GetSize() < u.Size {
		return true
	}
	return target.GetSize() == u.Size && !u.Role.IsPrey()
}

func (u *Unit) IsHuntable() bool {
	return u.Role.IsPrey() && !u.Hidden && !u.IsDefeated()
}

func (u *Unit) GetSize() int {
	return u.Size
}

func (u *Unit) RemainingDistress() int {
	return u.MaxDistress - u.Distress
}

func (u *Unit) IsDefeated() bool {
	return u.Distress >= u.MaxDistress
}

func (u *Unit) TakeDamage(amount int) int {
	u.Distress = min(u.Distress+amount, u.MaxDistress)
	return u.Distress
}

func (u *Unit) Hide(seenByPlayers bool) {
	u.Hidden = true
	u.HiddenInSight = seenByPlayers
}

func (u *Unit) Unhide() {
	u.Hidden = false
	u.HiddenInSight = false
}

func (u *Unit) Relieve(amount int) {
	u.Distress = max(u.Distress-amount, 0)
}

func (u *Unit) HasIdentName(name string) bool {
	for _, identName := range u.IdentNames {
		if strings.EqualFold(identName, name) {
			return true
		}
	}
	return false
}

// creatures that are no predators do not care who is around, aloof creatures at locations do not prevent player movement
func (u *Unit) IsAloof() bool {
	return !u.Role.IsPredator()
}

func (u *Unit) Bond(by *player.Unit) {
	u.BondCounter++
	if u.BondCounter >= u.BondThreshold {
		u.Bonded = true
		u.BondedTo = by
		u.Exhausted = false
	}
}

func (u *Unit) IsBondable() bool {
	return !u.Bonded && !u.Hidden && (u.IsAloof() || u.Exhausted)
}

// only prey can be confronted
func (u *Unit) IsConfrontable() bool {
	return u.Role.IsPrey() && !u.Bonded && !u.Hidden
}

// only predators can be distracted
func (u *Unit) IsDistractable() bool {
	return u.Role.IsPredator() && !u.Bonded && !u.Hidden && !u.Exhausted
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
