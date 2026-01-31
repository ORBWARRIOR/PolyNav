package algo

const (
	EPSILON = 1e-9
)

type Point struct {
	X, Y float64
}

// Triangle stores vertex indices and neighbour indices.
// T1, T2, T3 are neighbours opposite to vertices A, B, C respectively.
// -1 indicates convex hull boundary.
type Triangle struct {
	A, B, C    int  // Vertex Indices
	T1, T2, T3 int  // Neighbor Indices. -1 indicates convex hull boundary.
	Active     bool // Logical deletion
}

type Delaunay struct {
	Points       []Point
	Triangles    []Triangle
	superIndices [3]int
	lastCreated  int // Cache for Sloan's Walking Search
}

// GraphNode represents a Voronoi vertex for pathfinding.
// See docs/MATHEMATICS.md#2-voronoi-duality for dual graph theory.
// See docs/ALGORITHMS.md#3-pathfinding-heuristics for pathfinding context.
type GraphNode struct {
	ID        int
	X, Y      float64 // Coordinates of the Circumcenter (Voronoi Vertex)
	Neighbors []int
	Costs     []float64 // Edge costs for A* pathfinding
}
