package inventory

import "sort"

type Graph struct {
	packages map[string]Package
	reverse  map[string][]string
}

func NewGraph(s Snapshot) *Graph {
	g := &Graph{packages: make(map[string]Package), reverse: make(map[string][]string)}
	for _, p := range s.Packages {
		g.packages[p.ID] = p
	}
	for id, p := range g.packages {
		seen := map[string]bool{}
		for _, dep := range p.Dependencies {
			if !seen[dep] {
				g.reverse[dep] = append(g.reverse[dep], id)
				seen[dep] = true
			}
		}
	}
	for id := range g.reverse {
		sort.Strings(g.reverse[id])
	}
	return g
}
func (g *Graph) Dependencies(id string) []string {
	return append([]string(nil), g.packages[id].Dependencies...)
}
func (g *Graph) Dependents(id string) []string { return append([]string(nil), g.reverse[id]...) }
func (g *Graph) RootPaths(id string) [][]string {
	var result [][]string
	ancestors := map[string]bool{}
	var mark func(string)
	mark = func(node string) {
		if ancestors[node] {
			return
		}
		ancestors[node] = true
		for _, parent := range g.reverse[node] {
			mark(parent)
		}
	}
	mark(id)
	budget := 10000
	var walk func(string, []string, map[string]bool)
	walk = func(current string, path []string, seen map[string]bool) {
		if len(result) >= 100 || seen[current] || !ancestors[current] || budget <= 0 {
			return
		}
		budget--
		seen[current] = true
		defer delete(seen, current)
		path = append(path, current)
		if current == id {
			result = append(result, append([]string(nil), path...))
			return
		}
		for _, dep := range g.Dependencies(current) {
			walk(dep, path, seen)
		}
	}
	var roots []string
	for id, p := range g.packages {
		if p.Explicit {
			roots = append(roots, id)
		}
	}
	sort.Strings(roots)
	for _, root := range roots {
		walk(root, nil, make(map[string]bool))
	}
	return result
}
