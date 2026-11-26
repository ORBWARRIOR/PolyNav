package engine

import (
	"math"
	"testing"
)

func TestNewVoronoiMap(t *testing.T) {

	tests := []struct {
		name               string
		mesh               *Mesh
		expectedVoronoiMap *VoronoiMap
	}{
		{"NoTriangles", &Mesh{HalfEdges: []HalfEdge{
			{Origin: 0, Twin: NoneEdge, Next: 4, Triangle: 0},
			{Origin: 1, Twin: NoneEdge, Next: 0, Triangle: 0},
			{Origin: 3, Twin: NoneEdge, Next: 3, Triangle: 1},
			{Origin: 1, Twin: 4, Next: 5, Triangle: 1},
		},
			Triangles: []Triangle{}}, nil},
		{"NoEdges", &Mesh{HalfEdges: []HalfEdge{},
			Triangles: []Triangle{
				{Edge: 4, Tombstoned: false},
				{Edge: 5, Tombstoned: false},
			}}, nil},
		{"SampleMesh", &Mesh{
			Points: []Point{
				{X: 0, Y: 0},
				{X: 0.5, Y: 1.5},
				{X: 2, Y: 0},
				{X: 2, Y: 2},
				{X: 3.5, Y: 1.5},
				{X: 4, Y: 0},
			},
			HalfEdges: []HalfEdge{
				{Origin: 0, Twin: NoneEdge, Next: 4, Triangle: 0},
				{Origin: 1, Twin: NoneEdge, Next: 0, Triangle: 0},
				{Origin: 3, Twin: NoneEdge, Next: 3, Triangle: 1},
				{Origin: 1, Twin: 4, Next: 5, Triangle: 1},
				{Origin: 2, Twin: 3, Next: 1, Triangle: 0},
				{Origin: 2, Twin: 6, Next: 2, Triangle: 1},
				{Origin: 3, Twin: 5, Next: 7, Triangle: 2},
				{Origin: 2, Twin: 8, Next: 9, Triangle: 2},
				{Origin: 4, Twin: 7, Next: 10, Triangle: 3},
				{Origin: 4, Twin: NoneEdge, Next: 6, Triangle: 2},
				{Origin: 2, Twin: NoneEdge, Next: 11, Triangle: 3},
				{Origin: 5, Twin: NoneEdge, Next: 8, Triangle: 3},
			},
			Triangles: []Triangle{
				{Edge: 4, Tombstoned: false},
				{Edge: 5, Tombstoned: false},
				{Edge: 9, Tombstoned: false},
				{Edge: 8, Tombstoned: false},
			},
			LastInsertedEdge: 10,
		}, &VoronoiMap{
			Points: []Point{{X: 1, Y: 0.5}, {X: 1.5, Y: 1}, {X: 2.5, Y: 1}, {X: 3, Y: 0.5}},
			Edges: []VoronoiEdge{
				{Tail: Point{X: 1.5, Y: 1}, Head: Point{X: 1, Y: 0.5}, Weight: math.Sqrt(0.5)},
				{Tail: Point{X: 1.5, Y: 1}, Head: Point{X: 2.5, Y: 1}, Weight: 1},
				{Tail: Point{X: 2.5, Y: 1}, Head: Point{X: 3, Y: 0.5}, Weight: math.Sqrt(0.5)},
			},
			VisitedTri: NewTriangleSet(),
			edgeIndex:  make(map[[2]Point]struct{}),
		}},
	}

	for _, tt := range tests {
		resultMap, _ := NewVoronoiMap(tt.mesh)
		assert(t, "voronoi map", tt.name, resultMap, tt.expectedVoronoiMap)
	}

}
