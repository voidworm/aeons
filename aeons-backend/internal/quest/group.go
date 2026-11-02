package quest

// a Quest or a Group, the building blocks of a scenario
type Item interface {
	begin()
	started() bool
	done() bool
	abandon()
	activeQuests() []*Quest
}

// A scenario is the outermost group.
type Group struct {
	Name       string
	Items      []Item
	Concurrent bool // all items are active at once, otherwise one after another
	Required   int  // concurrent groups are done once this many items are done, 0 means all of them
	hasBegun   bool
}

func (g *Group) Begin() {
	g.begin()
}

func (g *Group) begin() {
	g.hasBegun = true
	if g.Concurrent {
		for _, item := range g.Items {
			item.begin()
		}
		return
	}
	g.Advance()
}

// Advance runs after a quest was finished: a done concurrent group drops what is left of it,
// a sequential group starts its next item.
func (g *Group) Advance() {
	if !g.hasBegun {
		return
	}
	for _, item := range g.Items {
		if sub, ok := item.(*Group); ok {
			sub.Advance()
		}
	}
	if g.Concurrent {
		if g.done() {
			g.abandon()
		}
		return
	}
	for _, item := range g.Items {
		if item.done() {
			continue
		}
		if !item.started() {
			item.begin()
		}
		return
	}
}

func (g *Group) ActiveQuests() []*Quest {
	return g.activeQuests()
}

func (g *Group) activeQuests() []*Quest {
	active := []*Quest{}
	for _, item := range g.Items {
		active = append(active, item.activeQuests()...)
	}
	return active
}

func (g *Group) started() bool {
	return g.hasBegun
}

func (g *Group) done() bool {
	if g.Concurrent {
		return g.doneItems() >= g.required()
	}
	return g.doneItems() == len(g.Items)
}

func (g *Group) doneItems() int {
	n := 0
	for _, item := range g.Items {
		if item.done() {
			n++
		}
	}
	return n
}

func (g *Group) required() int {
	if g.Required == 0 {
		return len(g.Items)
	}
	return g.Required
}

func (g *Group) abandon() {
	for _, item := range g.Items {
		item.abandon()
	}
}
