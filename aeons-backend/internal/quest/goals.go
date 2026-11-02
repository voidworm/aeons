package quest

import (
	"aeons/internal/creature"
	"aeons/internal/location"
	"aeons/internal/player"
	"slices"
)

// what the goals look at, rebuilt from the game state whenever they are checked
type World struct {
	Players   []*player.Unit
	Creatures []*creature.Unit
}

// a location that may only be known once the quest is reached
type Place func() *location.Unit

func Fixed(text string) func() string {
	return func() string { return text }
}

func PlaceAt(l *location.Unit) Place {
	return func() *location.Unit { return l }
}

type Payment struct {
	Where  Place // only the players here can pay
	Amount int
}

func (p Payment) Payers(w World) []*player.Unit {
	return playersAt(w, p.Where())
}

type Goal struct {
	Text    func() string
	Met     func(World) bool
	Payment *Payment // the resources are spent when the quest is finished
}

func AllPlayersHaveKindred(text string) *Goal {
	return &Goal{
		Text: Fixed(text),
		Met: func(w World) bool {
			return slices.IndexFunc(w.Players, func(p *player.Unit) bool { return kindredOf(w, p) == nil }) == -1
		},
	}
}

func AllPlayersAt(where Place, text func() string) *Goal {
	return &Goal{
		Text: text,
		Met: func(w World) bool {
			return len(playersAt(w, where())) == len(w.Players)
		},
	}
}

func AnyPlayerAt(where Place, text string) *Goal {
	return &Goal{
		Text: Fixed(text),
		Met:  func(w World) bool { return len(playersAt(w, where())) > 0 },
	}
}

func OnlyBondedCreaturesWithPlayers(text string) *Goal {
	return &Goal{
		Text: Fixed(text),
		Met: func(w World) bool {
			for _, p := range w.Players {
				for _, c := range w.Creatures {
					if c.CurrentLocation == p.CurrentLocation && !c.Bonded {
						return false
					}
				}
			}
			return true
		},
	}
}

func CreaturesAt(where Place, atLeast int, matching func(*creature.Unit) bool, text string) *Goal {
	return &Goal{
		Text: Fixed(text),
		Met: func(w World) bool {
			n := 0
			for _, c := range w.Creatures {
				if c.CurrentLocation == where() && matching(c) {
					n++
				}
			}
			return n >= atLeast
		},
	}
}

func AllPlayersAndBondedCreaturesAt(where Place, text func() string) *Goal {
	return &Goal{
		Text: text,
		Met: func(w World) bool {
			for _, c := range w.Creatures {
				if c.Bonded && c.CurrentLocation != where() {
					return false
				}
			}
			return len(playersAt(w, where())) == len(w.Players)
		},
	}
}

// the creature only exists once it has been spawned, and is defeated once it is gone from the game
func CreatureDefeated(spawned func() *creature.Unit, text string) *Goal {
	return &Goal{
		Text: Fixed(text),
		Met: func(w World) bool {
			c := spawned()
			return c != nil && !slices.Contains(w.Creatures, c)
		},
	}
}

func CanSpend(where Place, amount int, text string) *Goal {
	payment := &Payment{Where: where, Amount: amount}
	return &Goal{
		Text:    Fixed(text),
		Payment: payment,
		Met: func(w World) bool {
			held := 0
			for _, p := range payment.Payers(w) {
				held += p.ResourcesAvailable
			}
			return held >= amount
		},
	}
}

func playersAt(w World, l *location.Unit) []*player.Unit {
	found := []*player.Unit{}
	for _, p := range w.Players {
		if p.CurrentLocation == l {
			found = append(found, p)
		}
	}
	return found
}

func kindredOf(w World, p *player.Unit) *creature.Unit {
	for _, c := range w.Creatures {
		if c.Bonded && c.BondedTo == p {
			return c
		}
	}
	return nil
}
