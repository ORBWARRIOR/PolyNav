package engine

// Epsilon is used for floating-point equality checks.
const Epsilon = 1e-9

// Point represents a 2D coordinate.
type Point struct {
	X, Y float64
}

func (p Point) Sub(q Point) Point {
	return Point{X: p.X - q.X, Y: p.Y - q.Y}
}

type VertexID int32
type EdgeID int32
type TriangleID int32

const None EdgeID = -1

// HalfEdge represents a directed edge of a triangle.
type HalfEdge struct {
	Origin        VertexID   // Point index where this edge starts
	Twin          EdgeID     // The opposite half-edge (None if it's on the boundary/hull)
	Next          EdgeID     // The next edge in the triangle CCW
	Triangle      TriangleID // The triangle this half-edge belongs to
	IsConstrained bool       // Locked in CDT
}

// Triangle references three points by their index in a point slice
// Circumcircle values are cached
type Triangle struct {
	Edge        EdgeID // One of the half-edges bounding this triangle
	CircumX     float64
	CircumY     float64
	CircumRsqrd float64
}

// Mesh is the complete output of a Delaunay triangulation.
type Mesh struct {
	Points    []Point    // Normalised input points
	HalfEdges []HalfEdge // Flat array storing all directed edges
	Triangles []Triangle // Flat array storing all triangles
}

// Stats provides summary metrics from a triangulation or path plan.
type Stats struct {
	PointCount    int
	TriangleCount int // len(Mesh.Triangles)
	EdgeCount     int // len(Mesh.HalfEdges) / 2
	HullEdges     int // Count of HalfEdges where Twin == None
}

// GetTriangleVertices returns the vertices of a given triangle
func (m *Mesh) GetTriangleVertices(tID TriangleID) (VertexID, VertexID, VertexID) {
	edgeA := m.Triangles[tID].Edge
	edgeB := m.HalfEdges[edgeA].Next
	edgeC := m.HalfEdges[edgeB].Next

	return m.HalfEdges[edgeA].Origin, m.HalfEdges[edgeB].Origin, m.HalfEdges[edgeC].Origin
}

// GetNeighboringFace returns the TriangleID sharing the given edge, or -1 if none
func (m *Mesh) GetNeighboringFace(eID EdgeID) TriangleID {
	twinID := m.HalfEdges[eID].Twin
	if twinID == None {
		return -1
	}
	return m.HalfEdges[twinID].Triangle
}
