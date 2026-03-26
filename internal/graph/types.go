// Package graph defines the structures and types for graph representation.
package graph

// Node represents a single node in the graph.
type Node struct {
	ID   string   // Unique identifier for the node
	Edges []string // List of IDs of nodes this node is connected to
}

// Graph represents a collection of nodes.
type Graph struct {
	Nodes map[string]*Node
}

// NewGraph creates and initializes a new Graph.
func NewGraph() *Graph {
	return &Graph{
		Nodes: make(map[string]*Node),
	}
}

// AddNode adds a new node to the graph.
func (g *Graph) AddNode(id string) {
	if _, exists := g.Nodes[id]; !exists {
		g.Nodes[id] = &Node{ID: id, Edges: []string{}}
	}
}

// AddEdge adds a directed edge from one node to another.
func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from) // Ensure from node exists
	g.AddNode(to)   // Ensure to node exists
	g.Nodes[from].Edges = append(g.Nodes[from].Edges, to)
}