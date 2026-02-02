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
	A, B, C    int32 // Vertex Indices (int32 for memory efficiency)
	T1, T2, T3 int32 // Neighbor Indices. -1 indicates convex hull boundary.
	Active     bool  // Logical deletion
	Constrained [3]bool // Bitmask or bools: is edge i constrained?
	Inside      bool    // Part of the constrained interior
}


type Segment struct {
	P1, P2 int
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
