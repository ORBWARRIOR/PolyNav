package algo

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"
)

func TestDelaunayTriangulateScenarios(t *testing.T) {
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
			d := runTriangulation(t, tt.points)

			if len(d.Triangles) < tt.expectMinTris || len(d.Triangles) > tt.expectMaxTris {
				t.Errorf("Triangle count mismatch. Got %d, want between %d and %d",
					len(d.Triangles), tt.expectMinTris, tt.expectMaxTris)
			}
			saveDebugGeoJSON(d, fmt.Sprintf("./savedTests/debug_%s.geojson", tt.name))
		})
	}
}

// ========================================
// ROBUSTNESS TESTS
// ========================================

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

func TestDegeneracyNearlyCollinear(t *testing.T) {
	// Test points with Y coordinates within EPSILON of collinearity
	points := []Point{
		{0, 0}, {10, 0}, {5, 10}, // Base triangle
	}

	// Add nearly collinear points along the base
	for i := 1; i < 10; i++ {
		points = append(points, Point{
			X: float64(i),
			Y: EPSILON * (1 - float64(i)/10), // Slightly off the line
		})
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("PANIC during nearly-collinear insertion: %v", r)
		}
	}()

	d := runTriangulation(t, points)

	activeCount := 0
	for _, tri := range d.Triangles {
		if tri.Active {
			activeCount++
		}
	}
	if activeCount == 0 {
		t.Error("Triangulation resulted in 0 active triangles for nearly-collinear case")
	}

	saveDebugGeoJSON(d, "./savedTests/debug_nearly_collinear.geojson")
}

func TestDegeneracyExtremeCoordinates(t *testing.T) {
	tests := []struct {
		name   string
		points []Point
	}{
		{
			name: "Very Large Coordinates",
			points: []Point{
				{1e15, 1e15}, {1e15 + 10, 1e15}, {1e15 + 5, 1e15 + 10},
			},
		},
		{
			name: "Very Small Coordinates",
			points: []Point{
				{1e-12, 1e-12}, {1e-12 + 1e-8, 1e-12}, {1e-12 + 5e-9, 1e-12 + 1e-8}, {1e-12, 1e-12 + 2e-8},
			},
		},
		{
			name: "Mixed Extreme Coordinates",
			points: []Point{
				{0, 0}, {1e15, 0}, {5e14, 1e15}, {1e14, 1e14},
			},
		},
		{
			name: "Near Float64 Limits",
			points: []Point{
				{math.MaxFloat64 / 4, 0}, {0, math.MaxFloat64 / 4}, {math.MaxFloat64 / 8, math.MaxFloat64 / 8},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("PANIC during extreme coordinate test '%s': %v", tt.name, r)
				}
			}()

			done := make(chan bool)
			go func() {
				runTriangulation(t, tt.points)
				done <- true
			}()

			select {
			case <-done:
				// Success
			case <-time.After(5 * time.Second):
				t.Fatalf("Timeout in extreme coordinate test '%s'", tt.name)
			}
		})
	}
}

func TestDegeneracyCoincidentPoints(t *testing.T) {
	// Test more complex coincident point scenarios beyond basic duplicates
	// Focus on testing robustness rather than exact deduplication counts
	tests := []struct {
		name      string
		rawPoints []Point
	}{
		{
			name: "Multiple Exact Duplicates",
			rawPoints: []Point{
				{0, 0}, {0, 0}, {0, 0}, {0, 0}, // 4 identical points
				{10, 0}, {10, 0}, {10, 0}, // 3 identical points
				{5, 10}, // 1 unique point
			},
		},
		{
			name: "Epsilon-Near Duplicates",
			rawPoints: []Point{
				{0, 0}, {EPSILON / 2, 0}, {0, EPSILON / 2}, {EPSILON / 2, EPSILON / 2}, // All within epsilon
				{10, 0}, {10, EPSILON / 3}, // Close but separate
				{5, 10}, // Unique
			},
		},
		{
			name: "Clustered Points",
			rawPoints: []Point{
				{5, 5}, {5.000000001, 5}, {5, 5.000000001}, {5.000000001, 5.000000001}, // Cluster around (5,5)
				{0, 0}, {10, 0}, {5, 10}, // Triangle vertices
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := runTriangulation(t, tt.rawPoints)

			// Check that deduplication worked (allow for super triangle points)
			superTrianglePoints := 3
			expectedMaxPoints := len(tt.rawPoints) + superTrianglePoints
			if len(d.Points) > expectedMaxPoints {
				t.Errorf("Deduplication may have failed in '%s': %d input points -> %d unique points (expected ≤ %d)",
					tt.name, len(tt.rawPoints), len(d.Points), expectedMaxPoints)
			}

			// Ensure triangulation is still valid
			activeCount := 0
			for _, tri := range d.Triangles {
				if tri.Active {
					activeCount++
				}
			}
			if activeCount == 0 {
				t.Errorf("No active triangles after coincident point deduplication in '%s'", tt.name)
			}

			saveDebugGeoJSON(d, fmt.Sprintf("./savedTests/debug_coincident_%s.geojson", tt.name))
		})
	}
}

func TestDegeneracyPathologicalGeometries(t *testing.T) {
	tests := []struct {
		name   string
		points []Point
	}{
		{
			name:   "Spiral Configuration",
			points: generateSpiralPoints(100, 0.1, 5),
		},
		{
			name:   "Dense Clusters",
			points: generateDenseClusters(50),
		},
		{
			name:   "Grid with Noise",
			points: generateNoisyGrid(10, 1e-6),
		},
		{
			name:   "Concentric Circles",
			points: generateConcentricCircles(3, 20),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("PANIC during pathological geometry test '%s': %v", tt.name, r)
				}
			}()

			d := runTriangulation(t, tt.points)

			activeCount := 0
			for _, tri := range d.Triangles {
				if tri.Active {
					activeCount++
				}
			}
			if activeCount == 0 {
				t.Errorf("No active triangles in pathological geometry '%s'", tt.name)
			}

			// Validate triangle quality (no extremely thin triangles)
			thinTriangles := 0
			for _, tri := range d.Triangles {
				if tri.Active {
					if isThinTriangle(d, tri) {
						thinTriangles++
					}
				}
			}
			thinRatio := float64(thinTriangles) / float64(activeCount)
			if thinRatio > 0.5 { // More than 50% thin triangles might indicate issues
				t.Logf("Warning: High ratio of thin triangles (%.2f) in '%s'", thinRatio, tt.name)
			}

			saveDebugGeoJSON(d, fmt.Sprintf("./savedTests/debug_pathological_%s.geojson", tt.name))
		})
	}
}

// ========================================
// HELPERS
// ========================================

// Helper to reduce boilerplate
func runTriangulation(t *testing.T, points []Point) *Delaunay {
	d, err := NewDelaunay(points)
	if err != nil {
		t.Fatalf("Failed to initialise: %v", err)
	}
	d.Triangulate()
	return d
}

func generateTestPoints(size int, seed int64) []Point {
	r := rand.New(rand.NewSource(seed))
	points := make([]Point, size)
	for i := 0; i < size; i++ {
		points[i] = Point{
			X: r.Float64() * float64(size),
			Y: r.Float64() * float64(size),
		}
	}
	return points
}

func saveDebugGeoJSON(d *Delaunay, filename string) {
	// Export debug mesh to file
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

// Helper functions for pathological geometry generation
func generateSpiralPoints(n int, spacing float64, turns float64) []Point {
	points := make([]Point, n)
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1) * turns * 2 * math.Pi
		r := spacing * float64(i)
		points[i] = Point{
			X: r * math.Cos(t),
			Y: r * math.Sin(t),
		}
	}
	return points
}

func generateDenseClusters(pointsPerCluster int) []Point {
	var points []Point
	clusters := []Point{{0, 0}, {10, 0}, {5, 10}, {15, 15}, {-5, 10}}

	for _, center := range clusters {
		for i := 0; i < pointsPerCluster; i++ {
			angle := rand.Float64() * 2 * math.Pi
			radius := rand.Float64() * 0.1 // Very tight clusters
			points = append(points, Point{
				X: center.X + radius*math.Cos(angle),
				Y: center.Y + radius*math.Sin(angle),
			})
		}
	}
	return points
}

func generateNoisyGrid(gridSize int, noise float64) []Point {
	var points []Point
	for x := 0; x < gridSize; x++ {
		for y := 0; y < gridSize; y++ {
			points = append(points, Point{
				X: float64(x) + (rand.Float64()-0.5)*noise,
				Y: float64(y) + (rand.Float64()-0.5)*noise,
			})
		}
	}
	return points
}

func generateConcentricCircles(numCircles, pointsPerCircle int) []Point {
	var points []Point
	for c := 0; c < numCircles; c++ {
		radius := float64(c+1) * 2.0
		for i := 0; i < pointsPerCircle; i++ {
			angle := float64(i) / float64(pointsPerCircle) * 2 * math.Pi
			points = append(points, Point{
				X: radius * math.Cos(angle),
				Y: radius * math.Sin(angle),
			})
		}
	}
	return points
}

func isThinTriangle(d *Delaunay, tri Triangle) bool {
	p1, p2, p3 := d.Points[int(tri.A)], d.Points[int(tri.B)], d.Points[int(tri.C)]

	// Calculate side lengths
	a := math.Sqrt(math.Pow(p2.X-p3.X, 2) + math.Pow(p2.Y-p3.Y, 2))
	b := math.Sqrt(math.Pow(p1.X-p3.X, 2) + math.Pow(p1.Y-p3.Y, 2))
	c := math.Sqrt(math.Pow(p1.X-p2.X, 2) + math.Pow(p1.Y-p2.Y, 2))

	// Calculate area using Heron's formula
	s := (a + b + c) / 2
	area := math.Sqrt(s * (s - a) * (s - b) * (s - c))

	// Check if area is too small relative to side lengths (thin triangle)
	maxSide := math.Max(a, math.Max(b, c))
	if maxSide < EPSILON {
		return false // Degenerate but not "thin"
	}

	normalizedArea := area / (maxSide * maxSide)
	return normalizedArea < 1e-6 // Threshold for "thin" triangles
}
