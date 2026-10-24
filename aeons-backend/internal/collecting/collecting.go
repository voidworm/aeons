package collecting

import (
	"aeons/internal/card"
	"aeons/internal/location"
	"aeons/internal/player"
	"log"
	"math/rand/v2"
)

// Collecting is the general term: harvesting gathers food, foraging gathers material.
type Mode string

const (
	Harvest Mode = "harvest"
	Forage  Mode = "forage"
)

func (m Mode) Verb() string {
	if m == Harvest {
		return "harvests"
	}
	return "forages"
}

// Pool is one kind of stock on a node. A RegrowthTimer of 0 means it never grows back.
type Pool struct {
	Remaining     int
	Max           int
	RegrowthTimer int
	ticks         int
}

func NewPool(start, max, regrowthTimer int) *Pool {
	return &Pool{Remaining: start, Max: max, RegrowthTimer: regrowthTimer}
}

func (p *Pool) Available() bool {
	return p.Remaining > 0
}

func (p *Pool) Take() {
	p.Remaining--
}

// Tick grows one unit back every RegrowthTimer turns, up to Max.
func (p *Pool) Tick() {
	if p.RegrowthTimer == 0 || p.Remaining >= p.Max {
		p.ticks = 0
		return
	}
	p.ticks++
	if p.ticks >= p.RegrowthTimer {
		p.Remaining++
		p.ticks = 0
	}
}

func (p *Pool) TicksUntilRegrowth() int {
	if p.RegrowthTimer == 0 || p.Remaining >= p.Max {
		return 0
	}
	return p.RegrowthTimer - p.ticks
}

type Node struct {
	ID             int
	Name           string
	Flavor         string
	StaticLocation *location.Unit
	Food           *Pool
	Material       *Pool
}

func (n *Node) pool(mode Mode) *Pool {
	if mode == Harvest {
		return n.Food
	}
	return n.Material
}

func (n *Node) CanCollect(mode Mode) bool {
	return (mode == Harvest || mode == Forage) && n.pool(mode).Available()
}

// Modes lists what can currently be collected here.
func (n *Node) Modes() []Mode {
	modes := []Mode{}
	for _, m := range []Mode{Harvest, Forage} {
		if n.CanCollect(m) {
			modes = append(modes, m)
		}
	}
	return modes
}

func (n *Node) Tick() {
	n.Food.Tick()
	n.Material.Tick()
}

func (n *Node) GetTargetType() card.TargetType {
	return card.TargetTypeNode
}

func (n *Node) GetID() int {
	return n.ID
}

// Effect takes one unit from the node's pool and hands it to the player.
type Effect struct {
	Player *player.Unit
	Node   *Node
	Mode   Mode
}

func (e *Effect) Apply() {
	e.Node.pool(e.Mode).Take()
	if e.Mode == Harvest {
		e.Player.Gather(1, 0)
	} else {
		e.Player.Gather(0, 1)
	}
	log.Printf("[COLLECTING] %s %s %s\n", e.Player.Name, e.Mode.Verb(), e.Node.Name)
}

type GenerateFunc func(id int, location *location.Unit) *Node

func GenerateRandomNode(id int, location *location.Unit) *Node {
	generateFuncs := []GenerateFunc{
		GenerateBerryBush,
		GenerateBrush,
		GenerateMushroomField,
		GenerateFruitTree,
	}
	f := generateFuncs[rand.IntN(len(generateFuncs))]
	return f(id, location)
}

func GenerateBerryBush(id int, location *location.Unit) *Node {
	return &Node{
		ID:             id,
		Name:           "Berry Bush",
		Flavor:         "The berries around here are very sweet. Raccoons in particular like that.",
		StaticLocation: location,
		Food:           NewPool(2, 3, 3),
		Material:       NewPool(2, 3, 3),
	}
}

func GenerateBrush(id int, location *location.Unit) *Node {
	return &Node{
		ID:             id,
		Name:           "Brush",
		Flavor:         "Thorns and sticks. You'll find a way to use these.",
		StaticLocation: location,
		Food:           NewPool(0, 0, 0),
		Material:       NewPool(2, 4, 3),
	}
}

func GenerateMushroomField(id int, location *location.Unit) *Node {
	return &Node{
		ID:             id,
		Name:           "Mushroom Field",
		Flavor:         "A multitude of fungi growing in a large patch.",
		StaticLocation: location,
		Food:           NewPool(2, 4, 2),
		Material:       NewPool(0, 0, 0),
	}
}

func GenerateFruitTree(id int, location *location.Unit) *Node {
	return &Node{
		ID:             id,
		Name:           "Fruit Tree",
		Flavor:         "An almost impossible to exhaust source of both fruit and branches.",
		StaticLocation: location,
		Food:           NewPool(8, 10, 3),
		Material:       NewPool(4, 13, 5),
	}
}
