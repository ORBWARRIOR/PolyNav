package engine

import (
	"fmt"
	"math"
	"sort"
)

// Triangulate performs Delaunay triangulation via the Sloan incremental insertion algorithm
func Triangulate(points []Point) (*Mesh, error) {
	if len(points) < 3 {
		return nil, fmt.Errorf("delaunay: need at least 3 points, got %d", len(points))
	}
	normalised, scale, minX, minY := normalisePoints(points)
	uniq := deduplicatePoints(normalised)
	numOfPoints := len(uniq)
	if numOfPoints < 3 {
		return nil, fmt.Errorf("delaunay: only %d unique points after dedup, need 3", numOfPoints)
	}

	mesh, err := NewMeshWithSuperTriangle(numOfPoints)
	if err != nil {
		return nil, err
	}

	for i := range numOfPoints {
		InsertPoint(mesh, uniq[i], VertexID(i+3))
	}
	compact(mesh)
	mesh.Points = denormalisePoints(mesh.Points, scale, minX, minY)
	return mesh, nil
}

func NewMeshWithSuperTriangle(numOfPoints int) (*Mesh, error) {
	if numOfPoints < 3 {
		return nil, fmt.Errorf("delaunay: received %d points after dedup, need 3", numOfPoints)
	}
	m := &Mesh{
		Points:    make([]Point, 0, numOfPoints+3),    // Points + Super Triangle
		HalfEdges: make([]HalfEdge, 0, numOfPoints*6), // TODO:
		Triangles: make([]Triangle, 0, numOfPoints*2), // Look into Eulers Formula
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
	return m, nil
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

func normalisePoints(points []Point) (normalised []Point, scale, minX, minY float64) {

	minX, minY = math.MaxFloat64, math.MaxFloat64
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

	scale = math.Max(maxX-minX, maxY-minY)
	if scale <= Epsilon {
		scale = 1.0
	}

	normalised = make([]Point, len(points))
	for i, pt := range points {
		normalised[i] = Point{
			X: (pt.X - minX) / scale,
			Y: (pt.Y - minY) / scale,
		}
	}
	return normalised, scale, minX, minY
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

func denormalisePoints(points []Point, scale, minX, minY float64) []Point {
	denormalised := make([]Point, len(points))
	for i, pt := range points {
		denormalised[i] = Point{
			X: pt.X*scale + minX,
			Y: pt.Y*scale + minY,
		}
	}
	return denormalised
}

func compact(m *Mesh) {
	// Build the new half edge slice, ignores tombstoned triangles
	var newEdges []HalfEdge
	edgesMap := make(map[EdgeID]EdgeID)
	for oldIdx, ohe := range m.HalfEdges {
		tID := ohe.Triangle
		if tID != NoneTriangle && m.Triangles[tID].Tombstoned {
			continue
		}
		newIdx := EdgeID(len(newEdges))
		newEdges = append(newEdges, ohe)
		edgesMap[EdgeID(oldIdx)] = newIdx
	}

	// Build the new triangle list, ignores tombstoned triangles
	var newTriangles []Triangle
	trianglesMap := make(map[TriangleID]TriangleID)
	for oldIdx, tri := range m.Triangles {
		if tri.Tombstoned {
			continue
		}
		newIdx := TriangleID(len(newTriangles))
		newTriangles = append(newTriangles, tri)
		trianglesMap[TriangleID(oldIdx)] = newIdx
	}

	// Update half edge pointers
	for i := range newEdges {
		he := &newEdges[i]
		if he.Next != NoneEdge {
			if newNext, ok := edgesMap[he.Next]; ok {
				he.Next = newNext
			} else {
				he.Next = NoneEdge
			}
		}
		if he.Twin != NoneEdge {
			if newTwin, ok := edgesMap[he.Twin]; ok {
				he.Twin = newTwin
			} else {
				he.Twin = NoneEdge
			}
		}
		if he.Triangle != NoneTriangle {
			if newTri, ok := trianglesMap[he.Triangle]; ok {
				he.Triangle = newTri
			} else {
				he.Triangle = NoneTriangle
			}
		}
	}

	// Update triangle pointers
	for i := range newTriangles {
		tri := &newTriangles[i]
		if newEdge, ok := edgesMap[tri.Edge]; ok {
			tri.Edge = newEdge
		}
	}

	m.HalfEdges = newEdges
	m.Triangles = newTriangles
}
