package engine

// Point represents a 2D coordinate.
type Point struct {
	X, Y float64
}

// Triangle references three points by their index in a point slice
// Circumcircle values are cached
type Triangle struct {
	A, B, C     int
	CircumX     float64
	CircumY     float64
	CircumRsqrd float64
}

// Edge connects two point indices. Undirected: (A,B) == (B,A).
type Edge struct {
	A, B int
}

// Mesh is the complete output of a Delaunay triangulation.
type Mesh struct {
	Points   []Point // normalised input points
	Indices  []int   // triangle index
	Edges    []Edge  // unique edges (no duplicates, no interior diagonals doubled)
	Boundary []Edge  // hull edges (edges belonging to only one triangle)
}

// Stats provides summary metrics from a triangulation or path plan.
type Stats struct {
	PointCount    int
	TriangleCount int
	EdgeCount     int
	HullEdges     int
}

// Epsilon is used for floating-point equality checks.
const Epsilon = 1e-9
