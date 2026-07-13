// Internal tests for the engine package.
// Uses engine package to access unexported functions.
// These exercise the triangulation pipeline, not the Wails runtime.

package engine

import (
	"fmt"
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

// Creates a mesh consisting of the super triangle
func newSuperTriangle() *Mesh {
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
		Triangles:        []Triangle{{Edge: 0, Tombstoned: false}},
		LastInsertedEdge: EdgeID(2),
	}
}

// Creates a mesh consisting of the super triangle and a point at (0, 0)
// Looks like a fan with 3 blades
func newFan3() *Mesh {
	return &Mesh{
		Points: []Point{{-100, -100}, {100, -100}, {0, 100}, {0, 0}},
		HalfEdges: []HalfEdge{
			{Origin: 0, Twin: 3, Next: 6, Triangle: 1},
			{Origin: 1, Twin: 4, Next: 8, Triangle: 2},
			{Origin: 2, Twin: 5, Next: 10, Triangle: 3},
			{Origin: 1, Twin: 0, Next: 5, Triangle: NoneTriangle},
			{Origin: 2, Twin: 1, Next: 3, Triangle: NoneTriangle},
			{Origin: 0, Twin: 2, Next: 4, Triangle: NoneTriangle},
			{Origin: 1, Twin: 7, Next: 11, Triangle: 1},
			{Origin: 3, Twin: 6, Next: 1, Triangle: 2},
			{Origin: 2, Twin: 9, Next: 7, Triangle: 2},
			{Origin: 3, Twin: 8, Next: 2, Triangle: 3},
			{Origin: 0, Twin: 11, Next: 9, Triangle: 3},
			{Origin: 3, Twin: 10, Next: 0, Triangle: 1},
		},
		Triangles: []Triangle{
			{Edge: 0, Tombstoned: true},
			{Edge: 0, Tombstoned: false},
			{Edge: 1, Tombstoned: false},
			{Edge: 2, Tombstoned: false},
		},
		LastInsertedEdge: EdgeID(2),
	}
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
		t.Run(tt.name, func(t *testing.T) {
			resultPID := tt.mesh.addPoint(tt.point)
			assert(t, "mesh", tt.name, tt.mesh, tt.expectedMesh)
			assert(t, "resultPID", tt.name, resultPID, tt.expectedPID)
		})
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
		t.Run(tt.name, func(t *testing.T) {
			AB, BA := tt.mesh.addEdgePair(tt.a, tt.b)
			assert(t, "mesh", tt.name, tt.mesh, tt.expectedMesh)
			assert(t, "AB", tt.name, AB, tt.expectedEdgeID)
			assert(t, "BA", tt.name, BA, tt.expectedTwinEdgeID)
		})
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
		{"ThreePoint", 3, newSuperTriangle(), false},
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
		m := newSuperTriangle()
		for _, id := range tombstoned {
			m.Triangles[id].Tombstoned = true
		}
		return m
	}

	fan3 := func(tombstoned ...TriangleID) *Mesh {
		m := newFan3()
		for _, id := range tombstoned {
			m.Triangles[id].Tombstoned = true
		}
		return m
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
			mesh: fan3(1),
			expectedMesh: &Mesh{
				Points: []Point{{-100, -100}, {100, -100}, {0, 100}, {0, 0}},
				HalfEdges: []HalfEdge{
					{Origin: 1, Twin: 3, Next: 6, Triangle: 0},
					{Origin: 2, Twin: 4, Next: 8, Triangle: 1},
					{Origin: 1, Twin: NoneEdge, Next: 4, Triangle: NoneTriangle},
					{Origin: 2, Twin: 0, Next: 2, Triangle: NoneTriangle},
					{Origin: 0, Twin: 1, Next: 3, Triangle: NoneTriangle},
					{Origin: 3, Twin: NoneEdge, Next: 0, Triangle: 0},
					{Origin: 2, Twin: 7, Next: 5, Triangle: 0},
					{Origin: 3, Twin: 6, Next: 1, Triangle: 1},
					{Origin: 0, Twin: NoneEdge, Next: 7, Triangle: 1},
				},
				Triangles: []Triangle{
					{Edge: 0, Tombstoned: false},
					{Edge: 1, Tombstoned: false},
				},
			},
		},
		{
			name: "AllTombstonedInFan",
			mesh: fan3(1, 2, 3),
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

func TestMeshInherentProperties(t *testing.T) {
	m := newSuperTriangle()

	a := VertexID(0)
	b := VertexID(1)
	c := VertexID(2)
	AB := EdgeID(0)
	BC := EdgeID(1)
	CA := EdgeID(2)
	triangleID := TriangleID(0)

	resultEdge := m.HalfEdges[m.HalfEdges[m.HalfEdges[AB].Next].Next].Next
	resultTwinTwin := m.HalfEdges[m.HalfEdges[AB].Twin].Twin
	AB_Origin := m.HalfEdges[AB].Origin
	BC_Origin := m.HalfEdges[BC].Origin
	CA_Origin := m.HalfEdges[CA].Origin
	AB_Triangle := m.HalfEdges[AB].Triangle
	BC_Triangle := m.HalfEdges[BC].Triangle
	CA_Triangle := m.HalfEdges[CA].Triangle
	resultTriangleEdge := m.Triangles[triangleID].Edge
	assert(t, "edge cycle", "InherentEdgeProperty", resultEdge, AB)
	assert(t, "twin's twin", "InherentEdgeProperty", resultTwinTwin, AB)
	assert(t, "edge origin", "InherentEdgeProperty", AB_Origin, a)
	assert(t, "edge origin", "InherentEdgeProperty", BC_Origin, b)
	assert(t, "edge origin", "InherentEdgeProperty", CA_Origin, c)
	assert(t, "edge's triangle", "InherentEdgeProperty", AB_Triangle, triangleID)
	assert(t, "edge's triangle", "InherentEdgeProperty", BC_Triangle, triangleID)
	assert(t, "edge's triangle", "InherentEdgeProperty", CA_Triangle, triangleID)
	assert(t, "triangle's edge", "InherentTriangleProperty", resultTriangleEdge, AB)
}

func TestGetTriangleVertices(t *testing.T) {
	testName := "GetTriangleVertices"
	m := newSuperTriangle()
	expectedA := VertexID(0)
	expectedB := VertexID(1)
	expectedC := VertexID(2)
	expectedSuccess := true
	a, b, c, ok := m.GetTriangleVertices(0)

	assert(t, "success", testName, ok, expectedSuccess)
	assert(t, "success", testName, a, expectedA)
	assert(t, "success", testName, b, expectedB)
	assert(t, "success", testName, c, expectedC)
}

func TestGetEdgeOppositeP(t *testing.T) {
	tests := []struct {
		name         string
		edge         EdgeID
		P            VertexID
		expectedEdge EdgeID
		m            *Mesh
	}{
		{"OneJump", 0, 0, 1, newSuperTriangle()},
		{"ZeroJump", 1, 0, 1, newSuperTriangle()},
		{"TwoJump", 2, 0, 1, newSuperTriangle()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultEdge := getEdgeOppositeP(tt.m, tt.edge, tt.P)
			assert(t, "edge", tt.name, resultEdge, tt.expectedEdge)
		})
	}
}

func TestSplitTriangle(t *testing.T) {
	testName := "SplitTriangle"
	m := newSuperTriangle()
	expectedMesh := newFan3()
	expectedEdgeIDs := [3]EdgeID{6, 8, 10}

	m.addPoint(Point{X: 0, Y: 0})
	resultEdgeIDs := splitTriangle(m, 0, 3)
	assert(t, "mesh", testName, m, expectedMesh)
	assert(t, "edgeIDs", testName, resultEdgeIDs, expectedEdgeIDs)
}

func TestSplitEdge(t *testing.T) {
	m := &Mesh{
		Points: []Point{{X: 0, Y: 1}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 0}, {X: 0.5, Y: 0.5}},
		HalfEdges: []HalfEdge{
			{Origin: 0, Next: 1, Twin: 3, Triangle: 0},        // AB
			{Origin: 1, Next: 2, Twin: NoneEdge, Triangle: 0}, // BC
			{Origin: 2, Next: 0, Twin: NoneEdge, Triangle: 0}, // CA
			{Origin: 1, Next: 4, Twin: 0, Triangle: 1},        // BA
			{Origin: 0, Next: 5, Twin: NoneEdge, Triangle: 1}, // AD
			{Origin: 3, Next: 3, Twin: NoneEdge, Triangle: 1}, // DB
		},
		Triangles: []Triangle{
			{Edge: 0, Tombstoned: false}, {Edge: 3, Tombstoned: false},
		},
		LastInsertedEdge: 5,
	}
	expectedMesh := &Mesh{
		Points: []Point{{X: 0, Y: 1}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 0}, {X: 0.5, Y: 0.5}},
		HalfEdges: []HalfEdge{
			{Origin: 0, Next: 8, Twin: 3, Triangle: 2},         // 0  AB -> AP
			{Origin: 1, Next: 9, Twin: NoneEdge, Triangle: 3},  // 1  BC
			{Origin: 2, Next: 0, Twin: NoneEdge, Triangle: 2},  // 2  CA
			{Origin: 4, Next: 4, Twin: 0, Triangle: 4},         // 3  BA -> PA
			{Origin: 0, Next: 10, Twin: NoneEdge, Triangle: 4}, // 4  AD
			{Origin: 3, Next: 7, Twin: NoneEdge, Triangle: 5},  // 5  DB
			{Origin: 4, Next: 1, Twin: 7, Triangle: 3},         // 6  PB
			{Origin: 1, Next: 11, Twin: 6, Triangle: 5},        // 7  BP
			{Origin: 4, Next: 2, Twin: 9, Triangle: 2},         // 8  PC
			{Origin: 2, Next: 6, Twin: 8, Triangle: 3},         // 9  CP
			{Origin: 3, Next: 3, Twin: 11, Triangle: 4},        // 10 DP
			{Origin: 4, Next: 5, Twin: 10, Triangle: 5},        // 11 PD
		},
		Triangles: []Triangle{
			{Edge: 0, Tombstoned: true}, {Edge: 3, Tombstoned: true},
			{Edge: 0, Tombstoned: false}, {Edge: 1, Tombstoned: false},
			{Edge: 3, Tombstoned: false}, {Edge: 5, Tombstoned: false},
		},
		LastInsertedEdge: 5,
	}

	splitEdge(m, 0, 4)
	assert(t, "mesh", "SplitEdge", m, expectedMesh)
}

func TestWalkToPoint(t *testing.T) {
	tests := []struct {
		name            string
		m               *Mesh
		pt              Point
		expectedEdgeID  EdgeID
		expectedSuccess bool
	}{
		{"PointInTriangle", newFan3(), Point{X: 1, Y: 1}, 7, true},
		{"PointOnEdge", newFan3(), Point{X: 50, Y: 0}, 1, true},
		{"PointOutOfBounds", newFan3(), Point{X: 200, Y: 200}, 0, false},
	}

	for _, tt := range tests {
		resultEdgeID, ok := WalkToPoint(tt.m, tt.pt)

		assert(t, "success", tt.name, ok, tt.expectedSuccess)
		if tt.expectedSuccess {
			assert(t, "edge", tt.name, resultEdgeID, tt.expectedEdgeID)
		}
	}
}

func TestMeshStats(t *testing.T) {
	m := newSuperTriangle()
	resultStats := MeshStats(m)
	expectedStats := Stats{
		PointCount:    3,
		EdgeCount:     3,
		TriangleCount: 1,
		HullEdges:     3,
	}

	assert(t, "mesh stats", "MeshStats", resultStats, expectedStats)
}

func TestFlipEdge(t *testing.T) {
	makeEdgeFlippingMesh := func(D Point, E Point) *Mesh {
		return &Mesh{
			Points: []Point{
				{X: 0, Y: 0}, {X: 0, Y: 4}, {X: -0.1, Y: 2}, {X: 0.1, Y: 2}, D, E, // ABPCDE
			},
			HalfEdges: []HalfEdge{
				{Origin: 0, Next: 1, Twin: 3, Triangle: 0},         // 0  AB
				{Origin: 1, Next: 2, Twin: NoneEdge, Triangle: 0},  // 1  BP
				{Origin: 2, Next: 0, Twin: NoneEdge, Triangle: 0},  // 2  PA
				{Origin: 1, Next: 4, Twin: 0, Triangle: 1},         // 3  BA
				{Origin: 0, Next: 5, Twin: 9, Triangle: 1},         // 4  AC
				{Origin: 3, Next: 3, Twin: 6, Triangle: 1},         // 5  CB
				{Origin: 1, Next: 7, Twin: 5, Triangle: 2},         // 6  BC
				{Origin: 3, Next: 8, Twin: NoneEdge, Triangle: 2},  // 7  CD
				{Origin: 4, Next: 6, Twin: NoneEdge, Triangle: 2},  // 8  DB
				{Origin: 3, Next: 10, Twin: 4, Triangle: 3},        // 9  CA
				{Origin: 0, Next: 11, Twin: NoneEdge, Triangle: 3}, // 10 AE
				{Origin: 5, Next: 9, Twin: NoneEdge, Triangle: 3},  // 11 EC
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: false},
				{Edge: 3, Tombstoned: false},
				{Edge: 6, Tombstoned: false},
				{Edge: 9, Tombstoned: false},
			},
			LastInsertedEdge: 1, // BP
		}
	}

	makeFlippedMesh := func(D Point, E Point) *Mesh {
		return &Mesh{
			Points: []Point{
				{X: 0, Y: 0}, {X: 0, Y: 4}, {X: -0.1, Y: 2}, {X: 0.1, Y: 2}, D, E,
			},
			HalfEdges: []HalfEdge{
				{Origin: 2, Next: 5, Twin: 3, Triangle: 4},         // 0  PC (repurposed AB)
				{Origin: 1, Next: 0, Twin: NoneEdge, Triangle: 4},  // 1  BP
				{Origin: 2, Next: 4, Twin: NoneEdge, Triangle: 5},  // 2  PA
				{Origin: 3, Next: 2, Twin: 0, Triangle: 5},         // 3  CP (repurposed BA)
				{Origin: 0, Next: 3, Twin: 9, Triangle: 5},         // 4  AC
				{Origin: 3, Next: 1, Twin: 6, Triangle: 4},         // 5  CB
				{Origin: 1, Next: 7, Twin: 5, Triangle: 2},         // 6  BC
				{Origin: 3, Next: 8, Twin: NoneEdge, Triangle: 2},  // 7  CD
				{Origin: 4, Next: 6, Twin: NoneEdge, Triangle: 2},  // 8  DB
				{Origin: 3, Next: 10, Twin: 4, Triangle: 3},        // 9  CA
				{Origin: 0, Next: 11, Twin: NoneEdge, Triangle: 3}, // 10 AE
				{Origin: 5, Next: 9, Twin: NoneEdge, Triangle: 3},  // 11 EC
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: true},
				{Edge: 3, Tombstoned: true},
				{Edge: 6, Tombstoned: false},
				{Edge: 9, Tombstoned: false},
				{Edge: 1, Tombstoned: false},
				{Edge: 2, Tombstoned: false},
			},
			LastInsertedEdge: 1,
		}
	}

	tests := []struct {
		name         string
		m            *Mesh
		expectedMesh *Mesh
		edge         EdgeID
	}{
		{"SingleFlip", makeEdgeFlippingMesh(Point{X: 2, Y: 4}, Point{X: 2, Y: 0}), makeFlippedMesh(Point{X: 2, Y: 4}, Point{X: 2, Y: 0}), 0},
		{"DoubleFlipD", makeEdgeFlippingMesh(Point{X: 0.5, Y: 3}, Point{X: 2, Y: 0}), makeFlippedMesh(Point{X: 0.5, Y: 3}, Point{X: 2, Y: 0}), 0},
		{"DoubleFlipE", makeEdgeFlippingMesh(Point{X: 2, Y: 4}, Point{X: 0.5, Y: 1}), makeFlippedMesh(Point{X: 2, Y: 4}, Point{X: 0.5, Y: 1}), 0},
		{"TripleFlip", makeEdgeFlippingMesh(Point{X: 0.5, Y: 3}, Point{X: 0.5, Y: 1}), makeFlippedMesh(Point{X: 0.5, Y: 3}, Point{X: 0.5, Y: 1}), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flipEdge(tt.m, tt.edge)
			assert(t, "mesh", tt.name, tt.m, tt.expectedMesh)
		})
	}
}

func TestLegaliseEdge(t *testing.T) {
	makeMesh := func(D Point, E Point) *Mesh {
		return &Mesh{
			Points: []Point{
				{X: 0, Y: 0}, {X: 0, Y: 4}, {X: -0.1, Y: 2}, {X: 0.1, Y: 2}, D, E,
			},
			HalfEdges: []HalfEdge{
				{Origin: 0, Next: 1, Twin: 3, Triangle: 0},         // 0  AB
				{Origin: 1, Next: 2, Twin: NoneEdge, Triangle: 0},  // 1  BP
				{Origin: 2, Next: 0, Twin: NoneEdge, Triangle: 0},  // 2  PA
				{Origin: 1, Next: 4, Twin: 0, Triangle: 1},         // 3  BA
				{Origin: 0, Next: 5, Twin: 9, Triangle: 1},         // 4  AC
				{Origin: 3, Next: 3, Twin: 6, Triangle: 1},         // 5  CB
				{Origin: 1, Next: 7, Twin: 5, Triangle: 2},         // 6  BC
				{Origin: 3, Next: 8, Twin: NoneEdge, Triangle: 2},  // 7  CD
				{Origin: 4, Next: 6, Twin: NoneEdge, Triangle: 2},  // 8  DB
				{Origin: 3, Next: 10, Twin: 4, Triangle: 3},        // 9  CA
				{Origin: 0, Next: 11, Twin: NoneEdge, Triangle: 3}, // 10 AE
				{Origin: 5, Next: 9, Twin: NoneEdge, Triangle: 3},  // 11 EC
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: false},
				{Edge: 3, Tombstoned: false},
				{Edge: 6, Tombstoned: false},
				{Edge: 9, Tombstoned: false},
			},
			LastInsertedEdge: 1,
		}
	}

	tests := []struct {
		name         string
		m            *Mesh
		expectedMesh *Mesh
	}{
		{
			"SingleFlip",
			makeMesh(Point{X: 2, Y: 4}, Point{X: 2, Y: 0}),
			&Mesh{
				Points: []Point{
					{X: 0, Y: 0}, {X: 0, Y: 4}, {X: -0.1, Y: 2}, {X: 0.1, Y: 2}, {X: 2, Y: 4}, {X: 2, Y: 0},
				},
				HalfEdges: []HalfEdge{
					{Origin: 2, Next: 5, Twin: 3, Triangle: 4},
					{Origin: 1, Next: 0, Twin: NoneEdge, Triangle: 4},
					{Origin: 2, Next: 4, Twin: NoneEdge, Triangle: 5},
					{Origin: 3, Next: 2, Twin: 0, Triangle: 5},
					{Origin: 0, Next: 3, Twin: 9, Triangle: 5},
					{Origin: 3, Next: 1, Twin: 6, Triangle: 4},
					{Origin: 1, Next: 7, Twin: 5, Triangle: 2},
					{Origin: 3, Next: 8, Twin: NoneEdge, Triangle: 2},
					{Origin: 4, Next: 6, Twin: NoneEdge, Triangle: 2},
					{Origin: 3, Next: 10, Twin: 4, Triangle: 3},
					{Origin: 0, Next: 11, Twin: NoneEdge, Triangle: 3},
					{Origin: 5, Next: 9, Twin: NoneEdge, Triangle: 3},
				},
				Triangles: []Triangle{
					{Edge: 0, Tombstoned: true},
					{Edge: 3, Tombstoned: true},
					{Edge: 6, Tombstoned: false},
					{Edge: 9, Tombstoned: false},
					{Edge: 1, Tombstoned: false},
					{Edge: 2, Tombstoned: false},
				},
				LastInsertedEdge: 1,
			},
		},
		{
			"DoubleFlipD",
			makeMesh(Point{X: 0.5, Y: 3}, Point{X: 2, Y: 0}),
			&Mesh{
				Points: []Point{
					{X: 0, Y: 0}, {X: 0, Y: 4}, {X: -0.1, Y: 2}, {X: 0.1, Y: 2}, {X: 0.5, Y: 3}, {X: 2, Y: 0},
				},
				HalfEdges: []HalfEdge{
					{Origin: 2, Next: 7, Twin: 3, Triangle: 7},
					{Origin: 1, Next: 5, Twin: NoneEdge, Triangle: 6},
					{Origin: 2, Next: 4, Twin: NoneEdge, Triangle: 5},
					{Origin: 3, Next: 2, Twin: 0, Triangle: 5},
					{Origin: 0, Next: 3, Twin: 9, Triangle: 5},
					{Origin: 2, Next: 8, Twin: 6, Triangle: 6},
					{Origin: 4, Next: 0, Twin: 5, Triangle: 7},
					{Origin: 3, Next: 6, Twin: NoneEdge, Triangle: 7},
					{Origin: 4, Next: 1, Twin: NoneEdge, Triangle: 6},
					{Origin: 3, Next: 10, Twin: 4, Triangle: 3},
					{Origin: 0, Next: 11, Twin: NoneEdge, Triangle: 3},
					{Origin: 5, Next: 9, Twin: NoneEdge, Triangle: 3},
				},
				Triangles: []Triangle{
					{Edge: 0, Tombstoned: true},
					{Edge: 3, Tombstoned: true},
					{Edge: 6, Tombstoned: true},
					{Edge: 9, Tombstoned: false},
					{Edge: 1, Tombstoned: true},
					{Edge: 2, Tombstoned: false},
					{Edge: 1, Tombstoned: false},
					{Edge: 0, Tombstoned: false},
				},
				LastInsertedEdge: 1,
			},
		},
		{
			"DoubleFlipE",
			makeMesh(Point{X: 2, Y: 4}, Point{X: 0.5, Y: 1}),
			&Mesh{
				Points: []Point{
					{X: 0, Y: 0}, {X: 0, Y: 4}, {X: -0.1, Y: 2}, {X: 0.1, Y: 2}, {X: 2, Y: 4}, {X: 0.5, Y: 1},
				},
				HalfEdges: []HalfEdge{
					{Origin: 2, Next: 5, Twin: 3, Triangle: 4},
					{Origin: 1, Next: 0, Twin: NoneEdge, Triangle: 4},
					{Origin: 2, Next: 10, Twin: NoneEdge, Triangle: 7},
					{Origin: 3, Next: 4, Twin: 0, Triangle: 6},
					{Origin: 2, Next: 11, Twin: 9, Triangle: 6},
					{Origin: 3, Next: 1, Twin: 6, Triangle: 4},
					{Origin: 1, Next: 7, Twin: 5, Triangle: 2},
					{Origin: 3, Next: 8, Twin: NoneEdge, Triangle: 2},
					{Origin: 4, Next: 6, Twin: NoneEdge, Triangle: 2},
					{Origin: 5, Next: 2, Twin: 4, Triangle: 7},
					{Origin: 0, Next: 9, Twin: NoneEdge, Triangle: 7},
					{Origin: 5, Next: 3, Twin: NoneEdge, Triangle: 6},
				},
				Triangles: []Triangle{
					{Edge: 0, Tombstoned: true},
					{Edge: 3, Tombstoned: true},
					{Edge: 6, Tombstoned: false},
					{Edge: 9, Tombstoned: true},
					{Edge: 1, Tombstoned: false},
					{Edge: 2, Tombstoned: true},
					{Edge: 3, Tombstoned: false},
					{Edge: 2, Tombstoned: false},
				},
				LastInsertedEdge: 1,
			},
		},
		{
			"TripleFlip",
			makeMesh(Point{X: 0.5, Y: 3}, Point{X: 0.5, Y: 1}),
			&Mesh{
				Points: []Point{
					{X: 0, Y: 0}, {X: 0, Y: 4}, {X: -0.1, Y: 2}, {X: 0.1, Y: 2}, {X: 0.5, Y: 3}, {X: 0.5, Y: 1},
				},
				HalfEdges: []HalfEdge{
					{Origin: 2, Next: 7, Twin: 3, Triangle: 9},
					{Origin: 1, Next: 5, Twin: NoneEdge, Triangle: 8},
					{Origin: 2, Next: 10, Twin: NoneEdge, Triangle: 7},
					{Origin: 3, Next: 4, Twin: 0, Triangle: 6},
					{Origin: 2, Next: 11, Twin: 9, Triangle: 6},
					{Origin: 2, Next: 8, Twin: 6, Triangle: 8},
					{Origin: 4, Next: 0, Twin: 5, Triangle: 9},
					{Origin: 3, Next: 6, Twin: NoneEdge, Triangle: 9},
					{Origin: 4, Next: 1, Twin: NoneEdge, Triangle: 8},
					{Origin: 5, Next: 2, Twin: 4, Triangle: 7},
					{Origin: 0, Next: 9, Twin: NoneEdge, Triangle: 7},
					{Origin: 5, Next: 3, Twin: NoneEdge, Triangle: 6},
				},
				Triangles: []Triangle{
					{Edge: 0, Tombstoned: true},
					{Edge: 3, Tombstoned: true},
					{Edge: 6, Tombstoned: true},
					{Edge: 9, Tombstoned: true},
					{Edge: 1, Tombstoned: true},
					{Edge: 2, Tombstoned: true},
					{Edge: 3, Tombstoned: false},
					{Edge: 2, Tombstoned: false},
					{Edge: 1, Tombstoned: false},
					{Edge: 0, Tombstoned: false},
				},
				LastInsertedEdge: 1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			legaliseEdge(tt.m, 1, 2) // edge=1 (BP), vertex=2 (P)
			assert(t, "mesh", tt.name, tt.m, tt.expectedMesh)
		})
	}
}

func TestInsertPoint(t *testing.T) {
	points := []Point{
		{X: 0, Y: 0},
		{X: 0.5, Y: 1.5},
		{X: 2, Y: 2},
		{X: 3.5, Y: 1.5},
		{X: 4, Y: 0},
		{X: 2, Y: 0},
	}

	expectedMeshes := []*Mesh{
		{
			Points: []Point{
				{X: -100, Y: -100}, {X: 100, Y: -100}, {X: 0, Y: 100}, {X: 0, Y: 0},
			},
			HalfEdges: []HalfEdge{
				{Origin: 0, Next: 8, Twin: 3, Triangle: 2},
				{Origin: 1, Next: 10, Twin: 4, Triangle: 3},
				{Origin: 2, Next: 6, Twin: 5, Triangle: 1},
				{Origin: 1, Next: 5, Twin: 0, Triangle: -1},
				{Origin: 2, Next: 3, Twin: 1, Triangle: -1},
				{Origin: 0, Next: 4, Twin: 2, Triangle: -1},
				{Origin: 0, Next: 11, Twin: 7, Triangle: 1},
				{Origin: 3, Next: 0, Twin: 6, Triangle: 2},
				{Origin: 1, Next: 7, Twin: 9, Triangle: 2},
				{Origin: 3, Next: 1, Twin: 8, Triangle: 3},
				{Origin: 2, Next: 9, Twin: 11, Triangle: 3},
				{Origin: 3, Next: 2, Twin: 10, Triangle: 1},
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: true},
				{Edge: 2, Tombstoned: false},
				{Edge: 0, Tombstoned: false},
				{Edge: 1, Tombstoned: false},
			},
			LastInsertedEdge: 6,
		},
		{
			Points: []Point{
				{X: -100, Y: -100}, {X: 100, Y: -100}, {X: 0, Y: 100}, {X: 0, Y: 0}, {X: 0.5, Y: 1.5},
			},
			HalfEdges: []HalfEdge{
				{Origin: 0, Next: 8, Twin: 3, Triangle: 2},
				{Origin: 1, Next: 14, Twin: 4, Triangle: 5},
				{Origin: 2, Next: 6, Twin: 5, Triangle: 1},
				{Origin: 1, Next: 5, Twin: 0, Triangle: -1},
				{Origin: 2, Next: 3, Twin: 1, Triangle: -1},
				{Origin: 0, Next: 4, Twin: 2, Triangle: -1},
				{Origin: 0, Next: 11, Twin: 7, Triangle: 1},
				{Origin: 3, Next: 0, Twin: 6, Triangle: 2},
				{Origin: 1, Next: 7, Twin: 9, Triangle: 2},
				{Origin: 3, Next: 12, Twin: 8, Triangle: 4},
				{Origin: 2, Next: 16, Twin: 11, Triangle: 6},
				{Origin: 3, Next: 2, Twin: 10, Triangle: 1},
				{Origin: 1, Next: 17, Twin: 13, Triangle: 4},
				{Origin: 4, Next: 1, Twin: 12, Triangle: 5},
				{Origin: 2, Next: 13, Twin: 15, Triangle: 5},
				{Origin: 4, Next: 10, Twin: 14, Triangle: 6},
				{Origin: 3, Next: 15, Twin: 17, Triangle: 6},
				{Origin: 4, Next: 9, Twin: 16, Triangle: 4},
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: true},
				{Edge: 2, Tombstoned: false},
				{Edge: 0, Tombstoned: false},
				{Edge: 1, Tombstoned: true},
				{Edge: 9, Tombstoned: false},
				{Edge: 1, Tombstoned: false},
				{Edge: 10, Tombstoned: false},
			},
			LastInsertedEdge: 12,
		},
		{
			Points: []Point{
				{X: -100, Y: -100}, {X: 100, Y: -100}, {X: 0, Y: 100}, {X: 0, Y: 0},
				{X: 0.5, Y: 1.5}, {X: 2, Y: 2},
			},
			HalfEdges: []HalfEdge{
				{Origin: 0, Next: 8, Twin: 3, Triangle: 2},
				{Origin: 1, Next: 18, Twin: 4, Triangle: 7},
				{Origin: 2, Next: 6, Twin: 5, Triangle: 1},
				{Origin: 1, Next: 5, Twin: 0, Triangle: -1},
				{Origin: 2, Next: 3, Twin: 1, Triangle: -1},
				{Origin: 0, Next: 4, Twin: 2, Triangle: -1},
				{Origin: 0, Next: 11, Twin: 7, Triangle: 1},
				{Origin: 3, Next: 0, Twin: 6, Triangle: 2},
				{Origin: 1, Next: 7, Twin: 9, Triangle: 2},
				{Origin: 3, Next: 22, Twin: 8, Triangle: 10},
				{Origin: 2, Next: 16, Twin: 11, Triangle: 6},
				{Origin: 3, Next: 2, Twin: 10, Triangle: 1},
				{Origin: 3, Next: 21, Twin: 13, Triangle: 11},
				{Origin: 5, Next: 9, Twin: 12, Triangle: 10},
				{Origin: 2, Next: 20, Twin: 15, Triangle: 8},
				{Origin: 4, Next: 10, Twin: 14, Triangle: 6},
				{Origin: 3, Next: 15, Twin: 17, Triangle: 6},
				{Origin: 4, Next: 12, Twin: 16, Triangle: 11},
				{Origin: 2, Next: 23, Twin: 19, Triangle: 7},
				{Origin: 5, Next: 14, Twin: 18, Triangle: 8},
				{Origin: 4, Next: 19, Twin: 21, Triangle: 8},
				{Origin: 5, Next: 17, Twin: 20, Triangle: 11},
				{Origin: 1, Next: 13, Twin: 23, Triangle: 10},
				{Origin: 5, Next: 1, Twin: 22, Triangle: 7},
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: true},
				{Edge: 2, Tombstoned: false},
				{Edge: 0, Tombstoned: false},
				{Edge: 1, Tombstoned: true},
				{Edge: 9, Tombstoned: true},
				{Edge: 1, Tombstoned: true},
				{Edge: 10, Tombstoned: false},
				{Edge: 1, Tombstoned: false},
				{Edge: 14, Tombstoned: false},
				{Edge: 13, Tombstoned: true},
				{Edge: 22, Tombstoned: false},
				{Edge: 21, Tombstoned: false},
			},
			LastInsertedEdge: 18,
		},
		{
			Points: []Point{
				{X: -100, Y: -100}, {X: 100, Y: -100}, {X: 0, Y: 100}, {X: 0, Y: 0},
				{X: 0.5, Y: 1.5}, {X: 2, Y: 2}, {X: 3.5, Y: 1.5},
			},
			HalfEdges: []HalfEdge{
				{Origin: 0, Next: 8, Twin: 3, Triangle: 2},
				{Origin: 1, Next: 28, Twin: 4, Triangle: 14},
				{Origin: 2, Next: 6, Twin: 5, Triangle: 1},
				{Origin: 1, Next: 5, Twin: 0, Triangle: -1},
				{Origin: 2, Next: 3, Twin: 1, Triangle: -1},
				{Origin: 0, Next: 4, Twin: 2, Triangle: -1},
				{Origin: 0, Next: 11, Twin: 7, Triangle: 1},
				{Origin: 3, Next: 0, Twin: 6, Triangle: 2},
				{Origin: 1, Next: 7, Twin: 9, Triangle: 2},
				{Origin: 3, Next: 26, Twin: 8, Triangle: 15},
				{Origin: 2, Next: 16, Twin: 11, Triangle: 6},
				{Origin: 3, Next: 2, Twin: 10, Triangle: 1},
				{Origin: 3, Next: 21, Twin: 13, Triangle: 11},
				{Origin: 5, Next: 22, Twin: 12, Triangle: 16},
				{Origin: 2, Next: 20, Twin: 15, Triangle: 8},
				{Origin: 4, Next: 10, Twin: 14, Triangle: 6},
				{Origin: 3, Next: 15, Twin: 17, Triangle: 6},
				{Origin: 4, Next: 12, Twin: 16, Triangle: 11},
				{Origin: 2, Next: 24, Twin: 19, Triangle: 12},
				{Origin: 5, Next: 14, Twin: 18, Triangle: 8},
				{Origin: 4, Next: 19, Twin: 21, Triangle: 8},
				{Origin: 5, Next: 17, Twin: 20, Triangle: 11},
				{Origin: 3, Next: 25, Twin: 23, Triangle: 16},
				{Origin: 6, Next: 9, Twin: 22, Triangle: 15},
				{Origin: 5, Next: 29, Twin: 25, Triangle: 12},
				{Origin: 6, Next: 13, Twin: 24, Triangle: 16},
				{Origin: 1, Next: 23, Twin: 27, Triangle: 15},
				{Origin: 6, Next: 1, Twin: 26, Triangle: 14},
				{Origin: 2, Next: 27, Twin: 29, Triangle: 14},
				{Origin: 6, Next: 18, Twin: 28, Triangle: 12},
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: true},
				{Edge: 2, Tombstoned: false},
				{Edge: 0, Tombstoned: false},
				{Edge: 1, Tombstoned: true},
				{Edge: 9, Tombstoned: true},
				{Edge: 1, Tombstoned: true},
				{Edge: 10, Tombstoned: false},
				{Edge: 1, Tombstoned: true},
				{Edge: 14, Tombstoned: false},
				{Edge: 13, Tombstoned: true},
				{Edge: 22, Tombstoned: true},
				{Edge: 21, Tombstoned: false},
				{Edge: 18, Tombstoned: false},
				{Edge: 23, Tombstoned: true},
				{Edge: 1, Tombstoned: false},
				{Edge: 26, Tombstoned: false},
				{Edge: 25, Tombstoned: false},
			},
			LastInsertedEdge: 24,
		},
		{
			Points: []Point{
				{X: -100, Y: -100}, {X: 100, Y: -100}, {X: 0, Y: 100}, {X: 0, Y: 0},
				{X: 0.5, Y: 1.5}, {X: 2, Y: 2}, {X: 3.5, Y: 1.5}, {X: 4, Y: 0},
			},
			HalfEdges: []HalfEdge{
				{Origin: 0, Next: 8, Twin: 3, Triangle: 2},
				{Origin: 1, Next: 28, Twin: 4, Triangle: 14},
				{Origin: 2, Next: 6, Twin: 5, Triangle: 1},
				{Origin: 1, Next: 5, Twin: 0, Triangle: -1},
				{Origin: 2, Next: 3, Twin: 1, Triangle: -1},
				{Origin: 0, Next: 4, Twin: 2, Triangle: -1},
				{Origin: 0, Next: 11, Twin: 7, Triangle: 1},
				{Origin: 3, Next: 0, Twin: 6, Triangle: 2},
				{Origin: 1, Next: 7, Twin: 9, Triangle: 2},
				{Origin: 3, Next: 30, Twin: 8, Triangle: 17},
				{Origin: 2, Next: 16, Twin: 11, Triangle: 6},
				{Origin: 3, Next: 2, Twin: 10, Triangle: 1},
				{Origin: 3, Next: 21, Twin: 13, Triangle: 11},
				{Origin: 5, Next: 34, Twin: 12, Triangle: 20},
				{Origin: 2, Next: 20, Twin: 15, Triangle: 8},
				{Origin: 4, Next: 10, Twin: 14, Triangle: 6},
				{Origin: 3, Next: 15, Twin: 17, Triangle: 6},
				{Origin: 4, Next: 12, Twin: 16, Triangle: 11},
				{Origin: 2, Next: 24, Twin: 19, Triangle: 12},
				{Origin: 5, Next: 14, Twin: 18, Triangle: 8},
				{Origin: 4, Next: 19, Twin: 21, Triangle: 8},
				{Origin: 5, Next: 17, Twin: 20, Triangle: 11},
				{Origin: 5, Next: 33, Twin: 23, Triangle: 21},
				{Origin: 7, Next: 13, Twin: 22, Triangle: 20},
				{Origin: 5, Next: 29, Twin: 25, Triangle: 12},
				{Origin: 6, Next: 22, Twin: 24, Triangle: 21},
				{Origin: 1, Next: 32, Twin: 27, Triangle: 18},
				{Origin: 6, Next: 1, Twin: 26, Triangle: 14},
				{Origin: 2, Next: 27, Twin: 29, Triangle: 14},
				{Origin: 6, Next: 18, Twin: 28, Triangle: 12},
				{Origin: 1, Next: 35, Twin: 31, Triangle: 17},
				{Origin: 7, Next: 26, Twin: 30, Triangle: 18},
				{Origin: 6, Next: 31, Twin: 33, Triangle: 18},
				{Origin: 7, Next: 25, Twin: 32, Triangle: 21},
				{Origin: 3, Next: 23, Twin: 35, Triangle: 20},
				{Origin: 7, Next: 9, Twin: 34, Triangle: 17},
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: true},
				{Edge: 2, Tombstoned: false},
				{Edge: 0, Tombstoned: false},
				{Edge: 1, Tombstoned: true},
				{Edge: 9, Tombstoned: true},
				{Edge: 1, Tombstoned: true},
				{Edge: 10, Tombstoned: false},
				{Edge: 1, Tombstoned: true},
				{Edge: 14, Tombstoned: false},
				{Edge: 13, Tombstoned: true},
				{Edge: 22, Tombstoned: true},
				{Edge: 21, Tombstoned: false},
				{Edge: 18, Tombstoned: false},
				{Edge: 23, Tombstoned: true},
				{Edge: 1, Tombstoned: false},
				{Edge: 26, Tombstoned: true},
				{Edge: 25, Tombstoned: true},
				{Edge: 9, Tombstoned: false},
				{Edge: 26, Tombstoned: false},
				{Edge: 23, Tombstoned: true},
				{Edge: 34, Tombstoned: false},
				{Edge: 33, Tombstoned: false},
			},
			LastInsertedEdge: 30,
		},
		{
			Points: []Point{
				{X: -100, Y: -100}, {X: 100, Y: -100}, {X: 0, Y: 100}, {X: 0, Y: 0},
				{X: 0.5, Y: 1.5}, {X: 2, Y: 2}, {X: 3.5, Y: 1.5}, {X: 4, Y: 0}, {X: 2, Y: 0},
			},
			HalfEdges: []HalfEdge{
				{Origin: 0, Next: 8, Twin: 3, Triangle: 2},
				{Origin: 1, Next: 28, Twin: 4, Triangle: 14},
				{Origin: 2, Next: 6, Twin: 5, Triangle: 1},
				{Origin: 1, Next: 5, Twin: 0, Triangle: -1},
				{Origin: 2, Next: 3, Twin: 1, Triangle: -1},
				{Origin: 0, Next: 4, Twin: 2, Triangle: -1},
				{Origin: 0, Next: 11, Twin: 7, Triangle: 1},
				{Origin: 3, Next: 0, Twin: 6, Triangle: 2},
				{Origin: 1, Next: 7, Twin: 9, Triangle: 2},
				{Origin: 3, Next: 39, Twin: 8, Triangle: 23},
				{Origin: 2, Next: 16, Twin: 11, Triangle: 6},
				{Origin: 3, Next: 2, Twin: 10, Triangle: 1},
				{Origin: 4, Next: 41, Twin: 13, Triangle: 27},
				{Origin: 8, Next: 17, Twin: 12, Triangle: 26},
				{Origin: 2, Next: 20, Twin: 15, Triangle: 8},
				{Origin: 4, Next: 10, Twin: 14, Triangle: 6},
				{Origin: 3, Next: 15, Twin: 17, Triangle: 6},
				{Origin: 4, Next: 37, Twin: 16, Triangle: 26},
				{Origin: 2, Next: 24, Twin: 19, Triangle: 12},
				{Origin: 5, Next: 14, Twin: 18, Triangle: 8},
				{Origin: 4, Next: 19, Twin: 21, Triangle: 8},
				{Origin: 5, Next: 12, Twin: 20, Triangle: 27},
				{Origin: 6, Next: 34, Twin: 23, Triangle: 29},
				{Origin: 8, Next: 25, Twin: 22, Triangle: 28},
				{Origin: 5, Next: 29, Twin: 25, Triangle: 12},
				{Origin: 6, Next: 40, Twin: 24, Triangle: 28},
				{Origin: 1, Next: 32, Twin: 27, Triangle: 18},
				{Origin: 6, Next: 1, Twin: 26, Triangle: 14},
				{Origin: 2, Next: 27, Twin: 29, Triangle: 14},
				{Origin: 6, Next: 18, Twin: 28, Triangle: 12},
				{Origin: 1, Next: 35, Twin: 31, Triangle: 22},
				{Origin: 7, Next: 26, Twin: 30, Triangle: 18},
				{Origin: 6, Next: 31, Twin: 33, Triangle: 18},
				{Origin: 7, Next: 22, Twin: 32, Triangle: 29},
				{Origin: 8, Next: 33, Twin: 35, Triangle: 29},
				{Origin: 7, Next: 38, Twin: 34, Triangle: 22},
				{Origin: 8, Next: 9, Twin: 37, Triangle: 23},
				{Origin: 3, Next: 13, Twin: 36, Triangle: 26},
				{Origin: 8, Next: 30, Twin: 39, Triangle: 22},
				{Origin: 1, Next: 36, Twin: 38, Triangle: 23},
				{Origin: 5, Next: 23, Twin: 41, Triangle: 28},
				{Origin: 8, Next: 21, Twin: 40, Triangle: 27},
			},
			Triangles: []Triangle{
				{Edge: 0, Tombstoned: true},
				{Edge: 2, Tombstoned: false},
				{Edge: 0, Tombstoned: false},
				{Edge: 1, Tombstoned: true},
				{Edge: 9, Tombstoned: true},
				{Edge: 1, Tombstoned: true},
				{Edge: 10, Tombstoned: false},
				{Edge: 1, Tombstoned: true},
				{Edge: 14, Tombstoned: false},
				{Edge: 13, Tombstoned: true},
				{Edge: 22, Tombstoned: true},
				{Edge: 21, Tombstoned: true},
				{Edge: 18, Tombstoned: false},
				{Edge: 23, Tombstoned: true},
				{Edge: 1, Tombstoned: false},
				{Edge: 26, Tombstoned: true},
				{Edge: 25, Tombstoned: true},
				{Edge: 9, Tombstoned: true},
				{Edge: 26, Tombstoned: false},
				{Edge: 23, Tombstoned: true},
				{Edge: 34, Tombstoned: true},
				{Edge: 33, Tombstoned: true},
				{Edge: 35, Tombstoned: false},
				{Edge: 9, Tombstoned: false},
				{Edge: 34, Tombstoned: true},
				{Edge: 13, Tombstoned: true},
				{Edge: 37, Tombstoned: false},
				{Edge: 41, Tombstoned: false},
				{Edge: 40, Tombstoned: false},
				{Edge: 34, Tombstoned: false},
			},
			LastInsertedEdge: 35,
		},
	}

	m, err := NewMeshWithSuperTriangle(6)
	if err != nil {
		t.Fatal(err)
	}

	for i, p := range points {
		t.Run(fmt.Sprintf("Insert_%d_%v", i, p), func(t *testing.T) {
			InsertPoint(m, p, VertexID(i+3))
			assert(t, "mesh", fmt.Sprintf("after inserting %v", p), m, expectedMeshes[i])
		})
	}
}

func TestTriangulate(t *testing.T) {
	points := []Point{
		{X: 0, Y: 0},
		{X: 0.5, Y: 1.5},
		{X: 2, Y: 2},
		{X: 3.5, Y: 1.5},
		{X: 4, Y: 0},
		{X: 2, Y: 0},
	}

	expectedMesh := &Mesh{
		Points: []Point{
			{X: -400, Y: -400},
			{X: 400, Y: -400},
			{X: 0, Y: 400},
			{X: 0, Y: 0},
			{X: 0.5, Y: 1.5},
			{X: 2, Y: 0},
			{X: 2, Y: 2},
			{X: 3.5, Y: 1.5},
			{X: 4, Y: 0},
		},
		HalfEdges: []HalfEdge{
			{Origin: 0, Next: 8, Twin: 3, Triangle: 1},
			{Origin: 1, Next: 30, Twin: 4, Triangle: 7},
			{Origin: 2, Next: 6, Twin: 5, Triangle: 0},
			{Origin: 1, Next: 5, Twin: 0, Triangle: -1},
			{Origin: 2, Next: 3, Twin: 1, Triangle: -1},
			{Origin: 0, Next: 4, Twin: 2, Triangle: -1},
			{Origin: 0, Next: 11, Twin: 7, Triangle: 0},
			{Origin: 3, Next: 0, Twin: 6, Triangle: 1},
			{Origin: 1, Next: 7, Twin: 9, Triangle: 1},
			{Origin: 3, Next: 22, Twin: 8, Triangle: 3},
			{Origin: 2, Next: 16, Twin: 11, Triangle: 2},
			{Origin: 3, Next: 2, Twin: 10, Triangle: 0},
			{Origin: 3, Next: 21, Twin: 13, Triangle: 4},
			{Origin: 5, Next: 9, Twin: 12, Triangle: 3},
			{Origin: 2, Next: 19, Twin: 15, Triangle: 6},
			{Origin: 4, Next: 10, Twin: 14, Triangle: 2},
			{Origin: 3, Next: 15, Twin: 17, Triangle: 2},
			{Origin: 4, Next: 12, Twin: 16, Triangle: 4},
			{Origin: 6, Next: 20, Twin: 19, Triangle: 5},
			{Origin: 4, Next: 29, Twin: 18, Triangle: 6},
			{Origin: 4, Next: 24, Twin: 21, Triangle: 5},
			{Origin: 5, Next: 17, Twin: 20, Triangle: 4},
			{Origin: 1, Next: 13, Twin: 23, Triangle: 3},
			{Origin: 5, Next: 38, Twin: 22, Triangle: 11},
			{Origin: 5, Next: 18, Twin: 25, Triangle: 5},
			{Origin: 6, Next: 26, Twin: 24, Triangle: 9},
			{Origin: 5, Next: 33, Twin: 27, Triangle: 9},
			{Origin: 7, Next: 36, Twin: 26, Triangle: 10},
			{Origin: 2, Next: 32, Twin: 29, Triangle: 8},
			{Origin: 6, Next: 14, Twin: 28, Triangle: 6},
			{Origin: 2, Next: 35, Twin: 31, Triangle: 7},
			{Origin: 7, Next: 28, Twin: 30, Triangle: 8},
			{Origin: 6, Next: 31, Twin: 33, Triangle: 8},
			{Origin: 7, Next: 25, Twin: 32, Triangle: 9},
			{Origin: 1, Next: 40, Twin: 35, Triangle: 12},
			{Origin: 7, Next: 1, Twin: 34, Triangle: 7},
			{Origin: 5, Next: 41, Twin: 37, Triangle: 10},
			{Origin: 8, Next: 23, Twin: 36, Triangle: 11},
			{Origin: 1, Next: 37, Twin: 39, Triangle: 11},
			{Origin: 8, Next: 34, Twin: 38, Triangle: 12},
			{Origin: 7, Next: 39, Twin: 41, Triangle: 12},
			{Origin: 8, Next: 27, Twin: 40, Triangle: 10},
		},
		Triangles: []Triangle{
			{Edge: 2, Tombstoned: false},
			{Edge: 0, Tombstoned: false},
			{Edge: 10, Tombstoned: false},
			{Edge: 22, Tombstoned: false},
			{Edge: 21, Tombstoned: false},
			{Edge: 24, Tombstoned: false},
			{Edge: 29, Tombstoned: false},
			{Edge: 1, Tombstoned: false},
			{Edge: 28, Tombstoned: false},
			{Edge: 33, Tombstoned: false},
			{Edge: 27, Tombstoned: false},
			{Edge: 23, Tombstoned: false},
			{Edge: 34, Tombstoned: false},
		},
		LastInsertedEdge: 36,
	}

	m, err := Triangulate(points)
	if err != nil {
		t.Fatal(err)
	}
	assert(t, "mesh", "Triangulate", m, expectedMesh)
}

func TestTriangulateErrors(t *testing.T) {
	tests := []struct {
		name         string
		points       []Point
		expectedMesh *Mesh
		expectedErr  bool
	}{
		{"InsufficientPoints", []Point{{X: 0, Y: 0}}, nil, true},
		{"DuplicatePoints", []Point{{X: 0, Y: 0}, {X: 0, Y: 0}, {X: 0, Y: 0}}, nil, true},
	}

	for _, tt := range tests {
		m, err := Triangulate(tt.points)

		assert(t, "mesh", tt.name, m, tt.expectedMesh)
		assert(t, "success", tt.name, err != nil, tt.expectedErr)
	}
}
