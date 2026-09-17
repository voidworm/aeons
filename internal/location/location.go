package location

import (
	"errors"
	"slices"
)

type LocationEntity struct {
	ID                  int
	Name                string
	OutgoingConnections []*LocationEntity
}

func (le *LocationEntity) GetShortestPathTo(target *LocationEntity) ([]string, error) {

	//stuff we need to check
	queue := []*LocationEntity{le}

	//stuff we have checked
	visited := []*LocationEntity{le}

	//the parents we took
	parents := make(map[string]string)

	found := false

	for len(queue) != 0 {
		current := queue[0]
		queue = queue[1:]
		if current == target {
			found = true
			break
		}
		for _, v := range current.OutgoingConnections {
			alreadyVisited := slices.ContainsFunc(visited, func(location *LocationEntity) bool {
				return (location == v)
			})
			if !alreadyVisited {
				visited = append(visited, v)
				queue = append(queue, v)
				parents[v.Name] = current.Name
			}
		}
	}

	if found {
		path := []string{target.Name}
		currentChild := target.Name
		for {
			parent := parents[currentChild]
			if parent == "" {
				break
			}
			path = append(path, parent)
			currentChild = parent
		}
		slices.Reverse(path)
		return path, nil
	}

	//since maps should be a graph this should never happen but you never know
	return []string{}, errors.New("Locations are not connected")
}
