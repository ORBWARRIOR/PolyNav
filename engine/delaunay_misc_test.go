package engine_test

import (
	"math/rand/v2"
	"testing"

	"github.com/ORBWARRIOR/PolyNav/engine"
)

func BenchmarkDelaunay100(b *testing.B) {
	points := fillArray(100)
	for b.Loop() {
		engine.Triangulate(points)
	}
}

func BenchmarkDelaunay1K(b *testing.B) {
	points := fillArray(1_000)
	for b.Loop() {
		engine.Triangulate(points)
	}
}

func BenchmarkDelaunay10K(b *testing.B) {
	points := fillArray(10_000)
	for b.Loop() {
		engine.Triangulate(points)
	}
}

func BenchmarkDelaunay100K(b *testing.B) {
	points := fillArray(100_000)
	for b.Loop() {
		engine.Triangulate(points)
	}
}

// func BenchmarkDelaunay1M(b *testing.B) {
// 	points := fillArray(1_000_000)
// 	for b.Loop() {
// 		engine.Triangulate(points)
// 	}
// }

func fillArray(size int) []engine.Point {
	points := make([]engine.Point, size)
	for range size {
		x := rand.Float64() * float64(rand.IntN(1000))
		y := rand.Float64() * float64(rand.IntN(1000))

		points = append(points, engine.Point{X: x, Y: y})
	}
	return points
}
