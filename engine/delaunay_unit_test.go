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
