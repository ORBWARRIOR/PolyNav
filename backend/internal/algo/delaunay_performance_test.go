package algo

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestPerformanceRegression(t *testing.T) {
	count := 50000
	points := generateTestPoints(count, 42)

	start := time.Now()
	runTriangulation(t, points)
	duration := time.Since(start)

	// Performance expectations (in milliseconds)
	maxInitTime := 50 * time.Millisecond
	maxTriangulationTime := 500 * time.Millisecond

	if duration > maxInitTime+maxTriangulationTime {
		t.Errorf("Performance regression: took %v for %d points (expected < %v)",
			duration, count, maxInitTime+maxTriangulationTime)
	}

	t.Logf("Performance: %v for %d points (%.3f μs/point)",
		duration, count, float64(duration.Nanoseconds())/float64(count)/1000)
}

func TestMemoryUsage(t *testing.T) {
	sizes := []int{1000, 5000, 10000, 50000, 100000, 200000}

	for _, size := range sizes {
		t.Run(fmt.Sprintf("N_%d", size), func(t *testing.T) {
			// Force GC before measurement
			runtime.GC()
			var m1 runtime.MemStats
			runtime.ReadMemStats(&m1)

			points := generateTestPoints(size, 42)
			d := runTriangulation(t, points)

			runtime.GC()
			var m2 runtime.MemStats
			runtime.ReadMemStats(&m2)

			allocMB := float64(m2.Alloc-m1.Alloc) / 1024 / 1024
			pointsMB := float64(size*16) / 1024 / 1024                // 8 bytes per float64 * 2 coords
			trianglesMB := float64(len(d.Triangles)*64) / 1024 / 1024 // Rough estimate

			t.Logf("Size: %d, Points: %d, Triangles: %d", size, len(d.Points), len(d.Triangles))
			t.Logf("Memory: %.2f MB total (%.2f MB points, %.2f MB triangles)", allocMB, pointsMB, trianglesMB)
			t.Logf("Avg memory per point: %.2f bytes", float64(m2.Alloc-m1.Alloc)/float64(size))
		})
	}
}
