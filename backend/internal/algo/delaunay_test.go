package algo

import (
	"math/rand"
	"testing"
	"time"
)

func TestDelaunayTriangulateScenarios(t *testing.T) {
	t.Parallel() // Allow running in parallel with other tests

	tests := []struct {
		name          string
		points        []Point
		expectMinTris int
		expectMaxTris int
	}{
		{
			name: "Duplicate Points",
			points: []Point{
				{0, 0}, {10, 0}, {5, 10}, // Triangle
				{0, 0}, {10, 0}, // Duplicates
			},
			expectMinTris: 1,
			expectMaxTris: 1,
		},
		{
			name: "On-Edge Point",
			points: []Point{
				{0, 0}, {10, 0}, {5, 10}, // Base Triangle
				{5, 0}, // Exact midpoint of bottom edge
			},
			// With sorting, (5,0) is inserted before (10,0), so it's not "on edge" yet.
			// It becomes a valid vertex. Total 4 valid points -> 2 triangles.
			expectMinTris: 2,
			expectMaxTris: 2,
		},
		{
			name: "Simple Square with Center",
			points: []Point{
				{0, 0}, {10, 0}, {10, 10}, {0, 10}, {5, 5},
			},
			expectMinTris: 4,
			expectMaxTris: 4,
		},
		{
			name: "Small Triangle",
			points: []Point{
				{0, 0}, {10, 0}, {5, 10},
			},
			expectMinTris: 1,
			expectMaxTris: 1,
		},
		{
			name: "Collinear Points (Horizontal)",
			points: []Point{
				{0, 0}, {10, 0}, {5, 5}, {5, 0},
			},
			// With sorting, (5,0) processed before (10,0). Valid mesh.
			expectMinTris: 2,
			expectMaxTris: 2,
		},
		{
			name: "Grid 3x3",
			points: []Point{
				{0, 0}, {5, 0}, {10, 0},
				{0, 5}, {5, 5}, {10, 5},
				{0, 10}, {5, 10}, {10, 10},
			},
			expectMinTris: 8,
			expectMaxTris: 8,
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d, err := NewDelaunay(tt.points)
			if err != nil {
				t.Fatalf("Failed to initialise: %v", err)
			}

			d.Triangulate()

			if len(d.Triangles) < tt.expectMinTris || len(d.Triangles) > tt.expectMaxTris {
				t.Errorf("Triangle count mismatch. Got %d, want between %d and %d",
					len(d.Triangles), tt.expectMinTris, tt.expectMaxTris)
			}
		})
	}
}

func TestDelaunayRandomStress(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(314159265))
	count := 100000
	points := make([]Point, count)
	for i := 0; i < count; i++ {
		points[i] = Point{
			X: r.Float64() * float64(count),
			Y: r.Float64() * float64(count),
		}
	}

	start := time.Now()
	d, err := NewDelaunay(points)
	if err != nil {
		t.Fatalf("Failed to initialise: %v", err)
	}
	d.Triangulate()
	duration := time.Since(start)

	t.Logf("Triangulated %d points in %v", count, duration)
	// Expects 2-7ms for 1000 points
	// Expects 0.5-2.5s for 100,000 points

	// Euler's formula approximation for Delaunay: ~2N triangles
	// We allow some variance due to hull size
	if len(d.Triangles) < count || len(d.Triangles) > 3*count {
		t.Errorf("Triangle count suspicious for %d points: %d", count, len(d.Triangles))
	}
}

func TestDelaunayDebugOutput(t *testing.T) {
	t.Parallel()

	points := []Point{{0, 0}, {10, 0}, {5, 10}}
	d, err := NewDelaunay(points)
	if err != nil {
		t.Fatalf("Failed to initialise: %v", err)
	}
	d.Triangulate()

	json := d.DebugJSON()
	if len(json) < 10 {
		t.Error("Debug JSON output is too short")
	}
}
