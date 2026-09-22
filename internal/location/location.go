package location

import (
	"slices"
)

type Unit struct {
	ID                  int
	Name                string
	OutgoingConnections []*Unit
}

// returns the path between two locations, including each of the locations.
// this function assumes that the the locations do have a connection path
// if you intend to load maps that have locations that can't be reached, this will break.
func (le *Unit) GetShortestPathTo(target *Unit) []*Unit {
	return le.breathSearchFirst(target)
}

func (le *Unit) DistanceTo(target *Unit) int {
	return len(le.breathSearchFirst(target))
}

func (le *Unit) breathSearchFirst(target *Unit) []*Unit {
	//stuff we need to check
	queue := []*Unit{le}

	//stuff we have checked
	visited := map[*Unit]bool{le: true}

	//the parents we took
	parents := make(map[*Unit]*Unit)

	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if current == target {
			break
		}
		for _, v := range current.OutgoingConnections {
			alreadyVisited := visited[v]
			if !alreadyVisited {
				visited[v] = true
				queue = append(queue, v)
				parents[v] = current
			}
		}
	}

	path := []*Unit{target}
	currentChild := target
	for {
		parent := parents[currentChild]
		if parent == nil {
			break
		}
		path = append(path, parent)
		currentChild = parent
	}

	slices.Reverse(path)
	return path
}
