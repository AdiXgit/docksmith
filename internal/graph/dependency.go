package graph

type DependencyGraph struct {
	vertices map[string]struct{}
	edges    map[string]map[string]struct{}
}

// NewDependencyGraph creates a new DependencyGraph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		vertices: make(map[string]struct{}),
		edges:    make(map[string]map[string]struct{}),
	}
}

// AddVertex adds a vertex to the graph
func (g *DependencyGraph) AddVertex(v string) {
	g.vertices[v] = struct{}{}
}

// AddEdge adds a directed edge from vertex v1 to vertex v2
func (g *DependencyGraph) AddEdge(v1, v2 string) {
	if g.edges[v1] == nil {
		g.edges[v1] = make(map[string]struct{})
	}
	g.edges[v1][v2] = struct{}{}
}

// DetectCycle detects if there's a cycle in the graph
func (g *DependencyGraph) DetectCycle() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	for v := range g.vertices {
		if g.detectCycleUtil(v, visited, recStack) {
			return true
		}
	}
	return false
}

// detectCycleUtil is a helper for DetectCycle
func (g *DependencyGraph) detectCycleUtil(v string, visited map[string]bool, recStack map[string]bool) bool {
	if recStack[v] {
		return true
	}
	if visited[v] {
		return false
	}

	visited[v] = true
	recStack[v] = true

	for neighbor := range g.edges[v] {
		if g.detectCycleUtil(neighbor, visited, recStack) {
			return true
		}
	}

	recStack[v] = false
	return false
}

// TopologicalSort returns the topological sort of the graph
func (g *DependencyGraph) TopologicalSort() ([]string, error) {
	visited := make(map[string]bool)
	var stack []string

	for v := range g.vertices {
		if !visited[v] {
			if err := g.topologicalSortUtil(v, visited, &stack); err != nil {
				return nil, err
			}
		}
	}

	return reverse(stack), nil
}

// topologicalSortUtil is a helper for TopologicalSort
func (g *DependencyGraph) topologicalSortUtil(v string, visited map[string]bool, stack *[]string) error {
	visited[v] = true

	for neighbor := range g.edges[v] {
		if !visited[neighbor] {
			if err := g.topologicalSortUtil(neighbor, visited, stack); err != nil {
				return err
			}
		} else if _, ok := g.edges[neighbor][v]; ok {
			return fmt.Errorf("cycle detected")
		}
	}

	*stack = append(*stack, v)
	return nil
}

// reverse reverses a slice of strings
func reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		 s[i], s[j] = s[j], s[i]
	}
	return s
}

// Visualization (optional)
// You can add a function to visualize the graph using a library like Gonum.
// Here you'd typically convert the structure into a format suitable for visual representation.