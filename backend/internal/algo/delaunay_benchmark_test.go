package algo

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkDelaunayInsertion(b *testing.B) {
	sizes := []int{1000, 5000, 10000, 50000, 100000, 200000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("N_%d", size), func(b *testing.B) {
			points := generateTestPoints(size, 42)

			b.ResetTimer()
			b.ReportAllocs()

			var totalDuration time.Duration
			var finalTriangles int
			var memoryPerPoint float64

			for i := 0; i < b.N; i++ {
				start := time.Now()
				d := runTriangulationBench(b, points)
				duration := time.Since(start)
				totalDuration += duration
				finalTriangles = len(d.Triangles)

				// Estimate memory per point (rough calculation)
				memoryPerPoint = float64(len(d.Points)*16+len(d.Triangles)*64) / float64(len(d.Points))
			}

			avgDuration := totalDuration / time.Duration(b.N)
			reportBenchmarkResult(b, fmt.Sprintf("N_%d", size), size, avgDuration, finalTriangles, memoryPerPoint)
		})
	}
}

func BenchmarkWalkLocate(b *testing.B) {
	points := generateTestPoints(1000, 42)
	d := runTriangulationBench(b, points)

	// Test points for location
	testPoints := generateTestPoints(1000, 42)
	b.ResetTimer()
	b.ReportAllocs()

	var totalDuration time.Duration

	for i := 0; i < b.N; i++ {
		start := time.Now()
		for _, p := range testPoints {
			d.walkLocate(p, d.lastCreated)
		}
		duration := time.Since(start)
		totalDuration += duration
	}

	avgDuration := totalDuration / time.Duration(b.N)
	reportBenchmarkResult(b, "WalkLocate", len(testPoints), avgDuration, len(d.Triangles), 0)
}

func runTriangulationBench(b *testing.B, points []Point) *Delaunay {
	d, err := NewDelaunay(points)
	if err != nil {
		b.Fatalf("Failed to initialise: %v", err)
	}
	d.Triangulate()
	return d
}

func reportBenchmarkResult(b *testing.B, testName string, points int, duration time.Duration, triangles int, memoryPerPoint float64) {
	timePerPoint := float64(duration.Nanoseconds()) / float64(points) / 1000 // microseconds
	throughput := float64(points) / duration.Seconds()
	expectedTriangles := 2 * points // Euler's formula approximation

	status := "✓ PASSED EXPECTATIONS"
	if triangles < points || triangles > 3*points {
		status = "✗ FAILED TRIANGLE COUNT"
	}

	fmt.Printf("\n=== BENCHMARK REPORT: %s ===\n", testName)
	fmt.Printf("Test Case: %s\n", testName)
	fmt.Printf("- Points: %d\n", points)
	fmt.Printf("- Total Time: %v\n", duration)
	fmt.Printf("- Time per Point: %.2fμs\n", timePerPoint)
	fmt.Printf("- Memory per Point: %.1f bytes\n", memoryPerPoint)
	fmt.Printf("- Triangles: %d (expected: ~%d)\n", triangles, expectedTriangles)
	fmt.Printf("- Throughput: %.0f points/sec\n", throughput)
	fmt.Printf("- Status: %s\n", status)
	fmt.Printf("=====================================\n\n")

	b.ReportMetric(float64(duration.Nanoseconds())/float64(points), "ns/op")
	b.ReportMetric(memoryPerPoint, "B/op")
	b.ReportMetric(float64(triangles), "triangles")
}
