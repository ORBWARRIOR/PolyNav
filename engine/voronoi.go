package engine

import (
	"fmt"
	"math"
)

type VoronoiEdge struct {
	Tail   Point
	Head   Point
	Weight float64
}

type VoronoiMap struct {
	Points     []Point
	Edges      []VoronoiEdge
	VisitedTri TriangleSet
	edgeIndex  map[[2]Point]struct{} // tracks added edges for O(1) dedup
}

func NewVoronoiMap(m *Mesh) (*VoronoiMap, error) {
	if len(m.Triangles) < 2 {
		return nil, fmt.Errorf("cannot create a voronoi map with less than 2 triangles")
	}
	if len(m.HalfEdges) < 4 {
		return nil, fmt.Errorf("cannot create a voronoi map with less than 4 edges")
	}
	voronoiMap := &VoronoiMap{
		Points:     make([]Point, 0, len(m.Triangles)),
		Edges:      make([]VoronoiEdge, 0, len(m.HalfEdges)),
		VisitedTri: NewTriangleSet(),
		edgeIndex:  make(map[[2]Point]struct{}),
	}

	err := buildVoronoi(m, voronoiMap)

	return voronoiMap, err
}

// BuildVoronoi computes the Voronoi dual graph iteratively over a Delaunay Mesh.
func buildVoronoi(m *Mesh, v *VoronoiMap) error {
	numTriangles := len(m.Triangles)
	circumcenters := make([]Point, numTriangles)
	validTri := make([]bool, numTriangles)

	// Pre-compute circumcenters once per triangle
	for i := range numTriangles {
		he1 := m.HalfEdges[m.Triangles[i].Edge]
		he2 := m.HalfEdges[he1.Next]
		he3 := m.HalfEdges[he2.Next]

		p1 := m.Points[he1.Origin]
		p2 := m.Points[he2.Origin]
		p3 := m.Points[he3.Origin]

		cc, ok := getCircumcenter(p1, p2, p3)
		if ok {
			circumcenters[i] = cc
			validTri[i] = true
		}
	}

	for _, cc := range circumcenters {
		v.Points = append(v.Points, cc)
	}

	visitedEdges := make(map[[2]EdgeID]bool, len(m.HalfEdges))

	// Iterative traversal over primary edges to build dual Voronoi segments.
	for eID, he := range m.HalfEdges {
		if he.Twin == NoneEdge {
			continue // Skip ray projection
		}

		twinID := he.Twin
		t1 := he.Triangle
		t2 := m.HalfEdges[twinID].Triangle

		if !validTri[t1] || !validTri[t2] {
			continue
		}

		// Normalize edge pair to form a key.
		e1, e2 := EdgeID(eID), twinID
		if e1 > e2 {
			e1, e2 = e2, e1
		}
		edgeKey := [2]EdgeID{e1, e2}

		if visitedEdges[edgeKey] {
			continue
		}
		visitedEdges[edgeKey] = true

		p1 := circumcenters[t1]
		p2 := circumcenters[t2]

		dx := p1.X - p2.X
		dy := p1.Y - p2.Y
		distSq := dx*dx + dy*dy

		weight := math.Sqrt(distSq)
		v.Edges = append(v.Edges, VoronoiEdge{
			Tail:   p1,
			Head:   p2,
			Weight: weight,
		})
	}

	return nil
}
