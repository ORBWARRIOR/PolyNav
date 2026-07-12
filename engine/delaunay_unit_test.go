// Internal tests for the engine package.
// Uses engine package to access unexported functions.
// These exercise the triangulation pipeline, not the Wails runtime.

package engine

import (
	"reflect"
	"testing"
)

func assert(t *testing.T, variableTested, testName string, result, expected any) bool {
	t.Helper()
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("%s mismatch - Test: %s - expected: %v, got: %v", variableTested, testName, expected, result)
		return false
	}
	return true
}

func TestNormalisePoints(t *testing.T) {
	tests := []struct {
		name           string
		points         []Point
		expectedPoints []Point
		expectedScale  float64
		expectedMinX   float64
		expectedMinY   float64
	}{
		{"SamplePoints", []Point{
			{X: 0, Y: 0},
			{X: 100, Y: 0},
			{X: 0, Y: 100},
		}, []Point{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 0, Y: 1},
		}, 100.0, 0.0, 0.0},
		{"IdenticalPoints", []Point{
			{X: 5, Y: 5},
			{X: 5, Y: 5},
			{X: 5, Y: 5},
		}, []Point{
			{X: 0, Y: 0},
			{X: 0, Y: 0},
			{X: 0, Y: 0},
		}, 1.0, 5.0, 5.0},
		{"NegativePoints", []Point{
			{X: 0, Y: 100},
			{X: 100, Y: -100},
			{X: -100, Y: -100},
		}, []Point{
			{X: 0.5, Y: 1},
			{X: 1, Y: 0},
			{X: 0, Y: 0},
		}, 200.0, -100.0, -100.0},
		{"LessThanEpsilon", []Point{
			{X: 0, Y: 1e-10},
			{X: 1e-10, Y: -1e-10},
			{X: -1e-10, Y: -1e-10},
		}, []Point{
			{X: 1e-10, Y: 2e-10},
			{X: 2e-10, Y: 0},
			{X: 0, Y: 0},
		}, 1.0, -1e-10, -1e-10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultPoints, resultScale, resultMinX, resultMinY := normalisePoints(tt.points)
			assert(t, "points", tt.name, resultPoints, tt.expectedPoints)
			assert(t, "scale", tt.name, resultScale, tt.expectedScale)
			assert(t, "minX", tt.name, resultMinX, tt.expectedMinX)
			assert(t, "minY", tt.name, resultMinY, tt.expectedMinY)
		})
	}
}

func TestDeduplicatePoints(t *testing.T) {
	tests := []struct {
		name           string
		points         []Point
		expectedPoints []Point
	}{
		{"NoDuplicates", []Point{
			{X: 1, Y: 1},
			{X: 2, Y: 2},
			{X: 3, Y: 3},
		}, []Point{
			{X: 1, Y: 1},
			{X: 2, Y: 2},
			{X: 3, Y: 3},
		}},
		{"AllDuplicates", []Point{
			{X: 1, Y: 1},
			{X: 1, Y: 1},
			{X: 1, Y: 1},
		}, []Point{
			{X: 1, Y: 1},
		}},
		{"LessThanEpsilon", []Point{
			{X: 0, Y: 0},
			{X: 1e-10, Y: 1e-10},
			{X: -1e-10, Y: -1e-10},
		}, []Point{
			{X: 0, Y: 0},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultPoints := deduplicatePoints(tt.points)
			assert(t, "points", tt.name, resultPoints, tt.expectedPoints)
		})
	}
}

func TestDenormalisePoints(t *testing.T) {
	tests := []struct {
		name   string
		points []Point
	}{
		{"SamplePoints", []Point{
			{X: 0, Y: 0},
			{X: 100, Y: 0},
			{X: 0, Y: 100},
		}},
		{"IdenticalPoints", []Point{
			{X: 5, Y: 5},
			{X: 5, Y: 5},
			{X: 5, Y: 5},
		}},
		{"NegativePoints", []Point{
			{X: 0, Y: 100},
			{X: 100, Y: -100},
			{X: -100, Y: -100},
		}},
		{"LessThanEpsilon", []Point{
			{X: 0, Y: 1e-10},
			{X: 1e-10, Y: -1e-10},
			{X: -1e-10, Y: -1e-10},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultPoints := denormalisePoints(normalisePoints(tt.points))
			assert(t, "points", tt.name, resultPoints, tt.points)
		})
	}
}

func TestOrient(t *testing.T) {
	tests := []struct {
		name         string
		A, B, C      Point
		expectedArea float64
	}{
		{"LHS",
			Point{X: 0, Y: 0},
			Point{X: 1, Y: 0},
			Point{X: 0, Y: 1},
			1.0},
		{"RHS",
			Point{X: 1, Y: 0},
			Point{X: 0, Y: 0},
			Point{X: 0, Y: 1},
			-1.0},
		{"Collinear",
			Point{X: 0, Y: 0},
			Point{X: 1, Y: 1},
			Point{X: 2, Y: 2},
			0.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			area := Orient(tt.A, tt.B, tt.C)
			assert(t, "area", tt.name, area, tt.expectedArea)
		})
	}
}

func TestCircumcircle(t *testing.T) {
	tests := []struct {
		name                                  string
		A, B, C                               Point
		expectedUx, expectedUy, expectedRSqrd float64
		expectedSuccess                       bool
	}{
		{"RightAngleTri",
			Point{X: 0, Y: 0},
			Point{X: 0, Y: 4},
			Point{X: 3, Y: 0},
			1.5, 2, 6.25, true},
		{"Collinear",
			Point{X: 0, Y: 0},
			Point{X: 1, Y: 1},
			Point{X: 2, Y: 2},
			0, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ux, uy, rSqrd, ok := GetCircumcircle(tt.A, tt.B, tt.C)
			assert(t, "success", tt.name, ok, tt.expectedSuccess)
			if tt.expectedSuccess {
				assert(t, "ux", tt.name, ux, tt.expectedUx)
				assert(t, "uy", tt.name, uy, tt.expectedUy)
				assert(t, "rSqrd", tt.name, rSqrd, tt.expectedRSqrd)
			}
		})
	}

	testName := "Cocircular"
	t.Run(testName, func(t *testing.T) {
		A := Point{X: 0, Y: 0}
		B := Point{X: 0, Y: 2}
		C := Point{X: 2, Y: 0}
		D := Point{X: 2, Y: 2}
		expectedUx := 1.0
		expectedUy := 1.0
		expectedRSqrd := 2.0
		ux1, uy1, rSqrd1, ok1 := GetCircumcircle(A, B, C)
		ux2, uy2, rSqrd2, ok2 := GetCircumcircle(B, C, D)

		if !ok1 {
			t.Fatalf("failed to get circumcircle for cocircular points A, B, C")
		}
		if !ok2 {
			t.Fatalf("failed to get circumcircle for cocircular points B, C, D")
		}

		assert(t, "ux1", testName, ux1, expectedUx)
		assert(t, "uy1", testName, uy1, expectedUy)
		assert(t, "rSqrd1", testName, rSqrd1, expectedRSqrd)
		assert(t, "ux2", testName, ux2, expectedUx)
		assert(t, "uy2", testName, uy2, expectedUy)
		assert(t, "rSqrd2", testName, rSqrd2, expectedRSqrd)
	})
}

func TestPointInCircumcircle(t *testing.T) {
	tests := []struct {
		name           string
		pt, A, B, C    Point
		expectedResult bool
	}{
		{"Inside",
			Point{X: 1.5, Y: 1.5},
			Point{X: 0, Y: 0},
			Point{X: 0, Y: 2},
			Point{X: 2, Y: 0},
			true},
		{"Outside",
			Point{X: -1, Y: -1},
			Point{X: 0, Y: 0},
			Point{X: 0, Y: 2},
			Point{X: 2, Y: 0},
			false},
		{"OnCircle",
			Point{X: 0, Y: 0},
			Point{X: 0, Y: 2},
			Point{X: 2, Y: 0},
			Point{X: 2, Y: 2},
			true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pointInCircumcircle(tt.pt, tt.A, tt.B, tt.C)
			assert(t, "result", tt.name, result, tt.expectedResult)
		})
	}
}

func TestMeshAddPoint(t *testing.T) {
	tests := []struct {
		name         string
		point        Point
		expectedPID  VertexID
		mesh         *Mesh
		expectedMesh *Mesh
	}{
		{"PositivePoint", Point{X: 1, Y: 1}, 0, &Mesh{}, &Mesh{Points: []Point{{X: 1, Y: 1}}}},
		{"NegativePoint", Point{X: -1, Y: -1}, 0, &Mesh{}, &Mesh{Points: []Point{{X: -1, Y: -1}}}},
	}

	for _, tt := range tests {
		resultPID := tt.mesh.addPoint(tt.point)
		assert(t, "mesh", tt.name, tt.mesh, tt.expectedMesh)
		assert(t, "resultPID", tt.name, resultPID, tt.expectedPID)
	}
}
func TestMeshAddEdge(t *testing.T) {
	tests := []struct {
		name                               string
		a, b                               VertexID
		expectedEdgeID, expectedTwinEdgeID EdgeID
		mesh                               *Mesh
		expectedMesh                       *Mesh
	}{
		{"SmallEdge", 0, 1, 0, 1, &Mesh{
			Points: []Point{{X: 1e-9, Y: 1e-9}, {X: -1e-9, Y: -1e-9}},
		}, &Mesh{
			Points: []Point{{X: 1e-9, Y: 1e-9}, {X: -1e-9, Y: -1e-9}},
			HalfEdges: []HalfEdge{
				{Origin: 0, Twin: 1, Next: NoneEdge, Triangle: NoneTriangle},
				{Origin: 1, Twin: 0, Next: NoneEdge, Triangle: NoneTriangle}},
		}},
		{"BigEdge", 0, 1, 0, 1, &Mesh{
			Points: []Point{{X: 1, Y: 1}},
		}, &Mesh{
			Points: []Point{{X: 1, Y: 1}},
			HalfEdges: []HalfEdge{
				{Origin: 0, Twin: 1, Next: NoneEdge, Triangle: NoneTriangle},
				{Origin: 1, Twin: 0, Next: NoneEdge, Triangle: NoneTriangle}},
		}},
	}

	for _, tt := range tests {
		AB, BA := tt.mesh.addEdgePair(tt.a, tt.b)
		assert(t, "mesh", tt.name, tt.mesh, tt.expectedMesh)
		assert(t, "AB", tt.name, AB, tt.expectedEdgeID)
		assert(t, "BA", tt.name, BA, tt.expectedTwinEdgeID)
	}
}

func TestNewMesh(t *testing.T) {
	tests := []struct {
		name         string
		numOfPoints  int
		expectedMesh *Mesh
		expectedErr  bool
	}{
		{"OnePoint", 1, nil, true},
		{"TwoPoint", 2, nil, true},
		{"ThreePoint", 3, &Mesh{
			Points: []Point{
				{X: -100, Y: -100},
				{X: 100, Y: -100},
				{X: 0, Y: 100}},
			HalfEdges: []HalfEdge{
				{Origin: 0, Twin: 3, Next: 1, Triangle: 0},
				{Origin: 1, Twin: 4, Next: 2, Triangle: 0},
				{Origin: 2, Twin: 5, Next: 0, Triangle: 0},
				{Origin: 1, Twin: 0, Next: 5, Triangle: NoneTriangle},
				{Origin: 2, Twin: 1, Next: 3, Triangle: NoneTriangle},
				{Origin: 0, Twin: 2, Next: 4, Triangle: NoneTriangle}},
			Triangles:        []Triangle{{Edge: EdgeID(0), Tombstoned: false}},
			LastInsertedEdge: EdgeID(2),
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultMesh, err := NewMeshWithSuperTriangle(tt.numOfPoints)
			assert(t, "mesh", tt.name, resultMesh, tt.expectedMesh)
			assert(t, "err", tt.name, err != nil, tt.expectedErr)
		})
	}
}

func TestCompact(t *testing.T) {
	superTri := func(tombstoned ...TriangleID) *Mesh {
		tombstone := make([]bool, 3)
		for _, id := range tombstoned {
			tombstone[id] = true
		}
		return &Mesh{
			Points: []Point{{-100, -100}, {100, -100}, {0, 100}},
			HalfEdges: []HalfEdge{
				{Origin: 0, Twin: 3, Next: 1, Triangle: 0},
				{Origin: 1, Twin: 4, Next: 2, Triangle: 0},
				{Origin: 2, Twin: 5, Next: 0, Triangle: 0},
				{Origin: 1, Twin: 0, Next: 5, Triangle: NoneTriangle},
				{Origin: 2, Twin: 1, Next: 3, Triangle: NoneTriangle},
				{Origin: 0, Twin: 2, Next: 4, Triangle: NoneTriangle},
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: tombstone[0]},
			},
		}
	}

	fan3 := func(tombstoned ...TriangleID) *Mesh {
		tombstone := make([]bool, 3)
		for _, id := range tombstoned {
			tombstone[id] = true
		}
		return &Mesh{
			Points: []Point{{-100, -100}, {100, -100}, {0, 100}, {0, 0}},
			HalfEdges: []HalfEdge{
				{Origin: 0, Twin: 9, Next: 1, Triangle: 0},
				{Origin: 1, Twin: 5, Next: 2, Triangle: 0},
				{Origin: 3, Twin: 7, Next: 0, Triangle: 0},
				{Origin: 1, Twin: 10, Next: 4, Triangle: 1},
				{Origin: 2, Twin: 8, Next: 5, Triangle: 1},
				{Origin: 3, Twin: 1, Next: 3, Triangle: 1},
				{Origin: 2, Twin: 11, Next: 7, Triangle: 2},
				{Origin: 0, Twin: 2, Next: 8, Triangle: 2},
				{Origin: 3, Twin: 4, Next: 6, Triangle: 2},
				{Origin: 1, Twin: 0, Next: 11, Triangle: NoneTriangle},
				{Origin: 2, Twin: 3, Next: 9, Triangle: NoneTriangle},
				{Origin: 0, Twin: 6, Next: 10, Triangle: NoneTriangle},
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: tombstone[0]},
				{Edge: 3, Tombstoned: tombstone[1]},
				{Edge: 6, Tombstoned: tombstone[2]},
			},
		}
	}

	tests := []struct {
		name         string
		mesh         *Mesh
		expectedMesh *Mesh
	}{
		{
			name:         "EmptyMesh",
			mesh:         &Mesh{},
			expectedMesh: &Mesh{},
		},
		{
			name:         "NothingTombstoned",
			mesh:         superTri(),
			expectedMesh: superTri(),
		},
		{
			name: "AllTombstoned",
			mesh: superTri(0),
			expectedMesh: &Mesh{
				Points: []Point{{-100, -100}, {100, -100}, {0, 100}},
				HalfEdges: []HalfEdge{
					{Origin: 1, Twin: NoneEdge, Next: 2, Triangle: NoneTriangle},
					{Origin: 2, Twin: NoneEdge, Next: 0, Triangle: NoneTriangle},
					{Origin: 0, Twin: NoneEdge, Next: 1, Triangle: NoneTriangle},
				},
			},
		},
		{
			name: "OneTombstonedInFan",
			mesh: fan3(2),
			expectedMesh: &Mesh{
				Points: []Point{{-100, -100}, {100, -100}, {0, 100}, {0, 0}},
				HalfEdges: []HalfEdge{
					{Origin: 0, Twin: 6, Next: 1, Triangle: 0},
					{Origin: 1, Twin: 5, Next: 2, Triangle: 0},
					{Origin: 3, Twin: NoneEdge, Next: 0, Triangle: 0},
					{Origin: 1, Twin: 7, Next: 4, Triangle: 1},
					{Origin: 2, Twin: NoneEdge, Next: 5, Triangle: 1},
					{Origin: 3, Twin: 1, Next: 3, Triangle: 1},
					{Origin: 1, Twin: 0, Next: 8, Triangle: NoneTriangle},
					{Origin: 2, Twin: 3, Next: 6, Triangle: NoneTriangle},
					{Origin: 0, Twin: NoneEdge, Next: 7, Triangle: NoneTriangle},
				},
				Triangles: []Triangle{
					{Edge: 0, Tombstoned: false},
					{Edge: 3, Tombstoned: false},
				},
			},
		},
		{
			name: "AllTombstonedInFan",
			mesh: fan3(0, 1, 2),
			expectedMesh: &Mesh{
				Points: []Point{{-100, -100}, {100, -100}, {0, 100}, {0, 0}},
				HalfEdges: []HalfEdge{
					{Origin: 1, Twin: NoneEdge, Next: 2, Triangle: NoneTriangle},
					{Origin: 2, Twin: NoneEdge, Next: 0, Triangle: NoneTriangle},
					{Origin: 0, Twin: NoneEdge, Next: 1, Triangle: NoneTriangle},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compact(tt.mesh)
			assert(t, "halfEdges", tt.name, tt.mesh.HalfEdges, tt.expectedMesh.HalfEdges)
			assert(t, "triangles", tt.name, tt.mesh.Triangles, tt.expectedMesh.Triangles)
			assert(t, "points", tt.name, tt.mesh.Points, tt.expectedMesh.Points)
		})
	}
}
