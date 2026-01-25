package algo

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"
)

// Helper to reduce boilerplate
func runTriangulation(t *testing.T, points []Point) *Delaunay {
	d, err := NewDelaunay(points)
	if err != nil {
		t.Fatalf("Failed to initialise: %v", err)
	}
	d.Triangulate()
	return d
}

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
			d := runTriangulation(t, tt.points)

			if len(d.Triangles) < tt.expectMinTris || len(d.Triangles) > tt.expectMaxTris {
				t.Errorf("Triangle count mismatch. Got %d, want between %d and %d",
					len(d.Triangles), tt.expectMinTris, tt.expectMaxTris)
			}
			saveDebugGeoJSON(d, fmt.Sprintf("./savedTests/debug_%s.geojson", tt.name))
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
	d := runTriangulation(t, points)
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
	d := runTriangulation(t, points)

	json, err := d.DebugJSON()
	if err != nil {
		t.Fatalf("Failed to generate debug JSON: %v", err)
	}
	if len(json) < 10 {
		t.Error("Debug JSON output is too short")
	}
	saveDebugGeoJSON(d, "./savedTests/debug_simple_triangle.geojson")
}

const TestEpsilon = 1e-9

func TestDegeneracyDuplicatePoints(t *testing.T) {
	rawPoints := []Point{
		{X: 10.0, Y: 10.0},
		{X: 10.0, Y: 10.0},
		{X: 10.0 + 1e-10, Y: 10.0},
		{X: 10.0, Y: 10.0 - 1e-10},
		{X: 20.0, Y: 20.0},
		{X: 0.0, Y: 0.0},
	}

	d := runTriangulation(t, rawPoints)

	expectedPoints := 3 + 3
	if len(d.Points) != expectedPoints {
		t.Errorf("Deduplication failure. Expected %d points, got %d.", expectedPoints, len(d.Points))
	}
	saveDebugGeoJSON(d, "./savedTests/debug_duplicate_points.geojson")
}

func TestDegeneracyCollinearStress(t *testing.T) {
	points := []Point{
		{0, 0}, {10, 0}, {5, 10},
	}
	for i := 1; i < 10; i++ {
		points = append(points, Point{X: float64(i), Y: 0})
	}
	for i := 1; i < 5; i++ {
		points = append(points, Point{X: 5, Y: float64(i)})
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PANIC during collinear insertion: %v", r)
		}
	}()

	d := runTriangulation(t, points)

	activeCount := 0
	for _, tr := range d.Triangles {
		if tr.Active {
			activeCount++
		}
	}
	if activeCount == 0 {
		t.Error("Triangulation resulted in 0 active triangles.")
	}
}

func TestDegeneracyLargeCoordinates(t *testing.T) {
	offset := 1_000_000.0
	points := []Point{
		{X: offset + 0, Y: offset + 0},
		{X: offset + 10, Y: offset + 0},
		{X: offset + 5, Y: offset + 10},
		{X: offset + 5, Y: offset + 5},
	}

	done := make(chan bool)
	go func() {
		// Can't easily use helper here inside goroutine if we want to catch Fatalf on t?
		// Actually t.Fatalf works from goroutine but stops that goroutine.
		// Use helper is fine.
		runTriangulation(t, points)
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout: Triangulation stuck.")
	}
}

func TestDegeneracyGrid(t *testing.T) {
	var points []Point
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			points = append(points, Point{float64(x), float64(y)})
		}
	}

	d := runTriangulation(t, points)

	if len(d.Triangles) < 100 {
		t.Errorf("Grid triangulation suspiciously small: %d triangles", len(d.Triangles))
	}
}

func TestRegressionStackOverflow(t *testing.T) {
	count := 5000
	points := make([]Point, count)
	for i := 0; i < count; i++ {
		x := float64(i)
		y := x * x * 0.001
		points[i] = Point{x, y}
	}

	rand.New(rand.NewSource(42)).Shuffle(len(points), func(i, j int) {
		points[i], points[j] = points[j], points[i]
	})

	runTriangulation(t, points)
}

func saveDebugGeoJSON(d *Delaunay, filename string) {
	// Check the exact signature in debug.go, but typically:
	data, err := d.DebugJSON()
	if err != nil {
		panic(err)
	}

	// Save to a file
	if err := os.WriteFile(filename, []byte(data), 0644); err != nil {
		panic(err)
	}
	fmt.Printf("Exported debug mesh to %s\n", filename)
}
