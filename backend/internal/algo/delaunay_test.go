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
				{0, 0}, {5, 0}, {10, 0}, {5, 5},
			},
			// Should result in 2 triangles connecting to the top point
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
			// 3x3 grid of points makes 2x2 grid of squares = 4 squares
			// Each square is 2 triangles -> 8 triangles
			expectMinTris: 8,
			expectMaxTris: 8,
		},
	}

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
	r := rand.New(rand.NewSource(42))
	count := 1000
	points := make([]Point, count)
	for i := 0; i < count; i++ {
		points[i] = Point{
			X: r.Float64() * 1000,
			Y: r.Float64() * 1000,
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

// BenchmarkDelaunay measures performance of the triangulation
func BenchmarkDelaunay(b *testing.B) {
	// 1000 points expects ~1.4ms
	r := rand.New(rand.NewSource(1337))
	count := 1000
	points := make([]Point, count)
	for i := 0; i < count; i++ {
		points[i] = Point{
			X: r.Float64() * 1000,
			Y: r.Float64() * 1000,
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d, err := NewDelaunay(points)
		if err != nil {
			b.Fatalf("Failed to initialise: %v", err)
		}
		d.Triangulate()
	}
}
