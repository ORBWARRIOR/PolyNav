package engine

import (
	"fmt"
	"math"
	"sort"
)

// Triangulate performs Delaunay triangulation via the Bowyer-Watson algorithm.
//
// Algorithm outline (implement as dummy stubs — fill in the real math):
//
//  1. Input normalisation
//     normalisePoints() — scale all points to 1x1 square
//     deduplicatePoints() — remove coincident points
//
//  2. Super-triangle
//     Create a triangle large enough to contain all points.
//     { (-100,-100), (+100,-100), (0,+100) } in normalised space.
//     The super-triangle vertices are appended at the end of the working slice.
//
//  3. Incremental insertion (Bowyer-Watson)
//     For each input point P:
//     a) Find all "bad" triangles whose circumcircle contains P
//     (use GetCircumcircle)
//     b) The union of bad triangles forms a polygon cavity.
//     Collect all edges of bad triangles.
//     c) Remove interior edges (edges shared by two bad triangles).
//     The remaining edges form the boundary of the cavity.
//     d) Connect P to every vertex of the cavity boundary.
//     This forms the new triangulation.
//
//  4. Super-triangle cleanup
//     Remove any triangle that shares a vertex with the super-triangle.
//
//  5. Output assembly
//     Build the Mesh: copy surviving triangles, extract unique edges,
//     identify boundary (hull) edges.
func Triangulate(points []Point) (*Mesh, error) {
	if len(points) < 3 {
		return nil, fmt.Errorf("delaunay: need at least 3 points, got %d", len(points))
	}

	uniq := deduplicatePoints(normalisePoints(points))
	length := len(uniq)
	if length < 3 {
		return nil, fmt.Errorf("delaunay: only %d unique points after dedup, need 3", length)
	}

	mesh := NewMeshWithSuperTriangle(length)
	for i := range length {
		InsertPoint(mesh, uniq[i], VertexID(i+3))
	}

	return &Mesh{
		Points: uniq,
	}, nil
}

func NewMeshWithSuperTriangle(lengthPoints int) *Mesh {
	m := &Mesh{
		Points:    make([]Point, 0, lengthPoints+3),    // Points + Super Triangle
		HalfEdges: make([]HalfEdge, 0, lengthPoints*6), // TODO:
		Triangles: make([]Triangle, 0, lengthPoints*2), // Look into Eulers Formula
	}

	// Add Super-Triangle Vertices
	m.Points = append(m.Points, Point{-100, -100}) // VertexID 0
	m.Points = append(m.Points, Point{100, -100})  // VertexID 1
	m.Points = append(m.Points, Point{0, 100})     // VertexID 2

	// Add Triangle 0 (The Super-Triangle)
	m.addTriangle(0) // TriangleID 0

	// Add the 3 Internal Half-Edges (Counter-Clockwise)
	m.HalfEdges = append(m.HalfEdges, HalfEdge{Origin: 0, Twin: 3, Next: 1, Triangle: 0}) // Edge 0
	m.HalfEdges = append(m.HalfEdges, HalfEdge{Origin: 1, Twin: 4, Next: 2, Triangle: 0}) // Edge 1
	m.HalfEdges = append(m.HalfEdges, HalfEdge{Origin: 2, Twin: 5, Next: 0, Triangle: 0}) // Edge 2

	// Add the 3 External Boundary Half-Edges (Clockwise Twins)
	// Their Triangle is set to -1 (None) because they face outward to infinity
	m.HalfEdges = append(m.HalfEdges, HalfEdge{Origin: 1, Twin: 0, Next: 5, Triangle: NoneTriangle}) // Edge 3
	m.HalfEdges = append(m.HalfEdges, HalfEdge{Origin: 2, Twin: 1, Next: 3, Triangle: NoneTriangle}) // Edge 4
	m.HalfEdges = append(m.HalfEdges, HalfEdge{Origin: 0, Twin: 2, Next: 4, Triangle: NoneTriangle}) // Edge 5

	m.LastInsertedEdge = EdgeID(2) // EdgeID of the last internal edge
	return m
}

// MeshStats computes summary statistics from a Mesh.
func MeshStats(m *Mesh) Stats {
	// Dummy — count from the data structure.
	edgeCount := len(m.HalfEdges) / 2
	var hullEdges int
	for i := 0; i < edgeCount; i++ {
		if m.HalfEdges[i].Twin == NoneEdge {
			hullEdges++
		}
	}
	return Stats{
		PointCount:    len(m.Points),
		TriangleCount: len(m.Triangles),
		EdgeCount:     edgeCount,
		HullEdges:     hullEdges,
	}
}

func normalisePoints(points []Point) []Point {
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, pt := range points {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}

	scale := math.Max(maxX-minX, maxY-minY)
	if scale <= Epsilon {
		scale = 1.0
	}

	normalised := make([]Point, len(points))
	for i, pt := range points {
		normalised[i] = Point{
			X: (pt.X - minX) / scale,
			Y: (pt.Y - minY) / scale,
		}
	}
	return normalised
}

func deduplicatePoints(points []Point) []Point {
	sort.Slice(points, func(i, j int) bool {
		if math.Abs(points[i].X-points[j].X) > Epsilon {
			return points[i].X < points[j].X
		}
		return points[i].Y < points[j].Y-Epsilon
	})

	uidx := 0
	for i := 1; i < len(points); i++ {
		if !isCoincident(points[uidx], points[i]) {
			uidx++
			points[uidx] = points[i]
		}
	}
	return points[:uidx+1]
}

func isCoincident(p1, p2 Point) bool {
	return math.Abs(p1.X-p2.X) < Epsilon && math.Abs(p1.Y-p2.Y) < Epsilon
}
