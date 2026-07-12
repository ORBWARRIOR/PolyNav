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
