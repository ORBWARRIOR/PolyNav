package engine

// Epsilon is used for floating-point equality checks.
const Epsilon = 1e-9

// Point represents a 2D coordinate.
type Point struct {
	X, Y float64
}

type VertexID int32
type EdgeID int32
type TriangleID int32

const NoneVertex VertexID = -1
const NoneEdge EdgeID = -1
const NoneTriangle TriangleID = -1

// HalfEdge represents a directed edge of a triangle.
type HalfEdge struct {
	Origin        VertexID   // Point index where this edge starts
	Twin          EdgeID     // The opposite half-edge (None if it's on the boundary/hull)
	Next          EdgeID     // The next edge in the triangle CCW
	Triangle      TriangleID // The triangle this half-edge belongs to
	IsConstrained bool       // Locked in CDT
}

// Triangle references three points by their index in a point slice
type Triangle struct {
	Edge       EdgeID // One of the half-edges bounding this triangle
	Tombstoned bool   // Marks this triangle as deleted
}

// Mesh is the complete output of a Delaunay triangulation.
type Mesh struct {
	Points           []Point    // Normalised input points
	HalfEdges        []HalfEdge // Flat array storing all directed edges
	Triangles        []Triangle // Flat array storing all triangles
	LastInsertedEdge EdgeID     // The most recently inserted edge, or None
}

// Stats provides summary metrics from a triangulation or path plan.
type Stats struct {
	PointCount    int
	TriangleCount int // len(Mesh.Triangles)
	EdgeCount     int // len(Mesh.HalfEdges) / 2
	HullEdges     int // Count of HalfEdges where Twin == None
}

// GetTriangleVertices returns the vertices of a given triangle and whether all indices were valid.
func (m *Mesh) GetTriangleVertices(tID TriangleID) (VertexID, VertexID, VertexID, bool) {
	if tID < 0 || int(tID) >= len(m.Triangles) {
		return 0, 0, 0, false
	}
	edgeA := m.Triangles[tID].Edge
	if edgeA < 0 || int(edgeA) >= len(m.HalfEdges) {
		return 0, 0, 0, false
	}
	edgeB := m.HalfEdges[edgeA].Next
	if edgeB < 0 || int(edgeB) >= len(m.HalfEdges) {
		return 0, 0, 0, false
	}
	edgeC := m.HalfEdges[edgeB].Next
	if edgeC < 0 || int(edgeC) >= len(m.HalfEdges) {
		return 0, 0, 0, false
	}

	return m.HalfEdges[edgeA].Origin, m.HalfEdges[edgeB].Origin, m.HalfEdges[edgeC].Origin, true
}

func (m *Mesh) addTriangle(edge EdgeID) TriangleID {
	m.Triangles = append(m.Triangles, Triangle{Edge: edge, Tombstoned: false})
	return TriangleID(len(m.Triangles) - 1)
}

func (m *Mesh) addEdgePair(x, p VertexID) (EdgeID, EdgeID) {
	id := EdgeID(len(m.HalfEdges))
	m.HalfEdges = append(m.HalfEdges, HalfEdge{Origin: x, Twin: id + 1, Next: NoneEdge, Triangle: NoneTriangle})
	m.HalfEdges = append(m.HalfEdges, HalfEdge{Origin: p, Twin: id, Next: NoneEdge, Triangle: NoneTriangle})
	return id, id + 1
}

func (m *Mesh) addPoint(p Point) VertexID {
	m.Points = append(m.Points, p)
	return VertexID(len(m.Points) - 1)
}
