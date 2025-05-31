package models

// NodePosition represents the position of a node in the graph
type NodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// NodeData represents the data associated with a graph node
type NodeData struct {
	ServiceName  string   `json:"serviceName"`
	Type         string   `json:"type"`
	Status       string   `json:"status"`
	Health       string   `json:"health"`
	Dependencies []string `json:"dependencies"`
	Color        string   `json:"color"`
}

// EdgeData represents the data associated with a graph edge
type EdgeData struct {
	Label    string `json:"label"`
	Type     string `json:"type"`
	Animated bool   `json:"animated"`
	Color    string `json:"color"`
}

// GraphNode represents a node in the dependency graph
type GraphNode struct {
	ID       string       `json:"id"`
	Label    string       `json:"label"`
	Type     string       `json:"type"`
	Status   string       `json:"status"`
	Position NodePosition `json:"position"`
	Data     NodeData     `json:"data"`
}

// GraphEdge represents an edge in the dependency graph
type GraphEdge struct {
	ID     string   `json:"id"`
	Source string   `json:"source"`
	Target string   `json:"target"`
	Type   string   `json:"type"`
	Data   EdgeData `json:"data"`
}

// DependencyGraph represents the complete graph structure
type DependencyGraph struct {
	Nodes []GraphNode `json:"nodes"`
	Edges []GraphEdge `json:"edges"`
}
