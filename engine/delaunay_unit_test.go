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
		{"PointOnEdge", newFan3(), Point{X: 0, Y: 1}, 2, true},
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
