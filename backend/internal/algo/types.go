package algo

const (
	EPSILON = 1e-9
)

type Point struct {
	X, Y float64
}

// Triangle stores indices. T1, T2, T3 are neighbors opposite to vertices A, B, C respectively.
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

// GraphNode is used for Pathfinding export.
// Corresponds to the Dual Graph concept in /docs/ALGORITHMS.md.
type GraphNode struct {
	ID        int
	X, Y      float64 // Coordinates of the Circumcenter (Voronoi Vertex)
	Neighbors []int
}
