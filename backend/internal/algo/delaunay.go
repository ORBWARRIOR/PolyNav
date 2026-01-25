package algo

import (
	"errors"
	"math"
)

// NewDelaunay initializes the mesh with a Super Triangle ensuring convex hull coverage.
func NewDelaunay(points []Point) (*Delaunay, error) {
	if len(points) < 3 {
		return nil, errors.New("insufficient points")
	}

	// Preallocate with factor 2*N (See docs/MATHEMATICS.md#4-memory-allocation-eulers-formula)
	d := &Delaunay{
		Points:    make([]Point, 0, len(points)+3),
		Triangles: make([]Triangle, 0, len(points)*2),
	}

	d.Points = append(d.Points, points...)

	// 1.1.2 Super Triangle (See docs/ALGORITHMS.md#11-the-super-triangle)
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, p := range points {
		if p.X < minX {
			minX = p.X
		}
		if p.Y < minY {
			minY = p.Y
		}
		if p.X > maxX {
			maxX = p.X
		}
		if p.Y > maxY {
			maxY = p.Y
		}
	}

	dx, dy := maxX-minX, maxY-minY
	deltaMax := math.Max(dx, dy) * 20.0 // See docs/ALGORITHMS.md#11-the-super-triangle
	midX, midY := (minX+maxX)/2.0, (minY+maxY)/2.0

	p1 := Point{midX - deltaMax, midY - deltaMax}
	p2 := Point{midX + deltaMax, midY - deltaMax}
	p3 := Point{midX, midY + deltaMax}

	idx1, idx2, idx3 := len(d.Points), len(d.Points)+1, len(d.Points)+2
	d.Points = append(d.Points, p1, p2, p3)
	d.superIndices = [3]int{idx1, idx2, idx3}

	// Create Root Triangle
	d.Triangles = append(d.Triangles, Triangle{
		A: idx1, B: idx2, C: idx3,
		T1: -1, T2: -1, T3: -1,
		Active: true,
	})
	d.lastCreated = 0

	return d, nil
}

// Triangulate executes Incremental Insertion (Lawson's Algorithm).
func (d *Delaunay) Triangulate() {
	originalCount := len(d.Points) - 3
	for i := 0; i < originalCount; i++ {
		d.insertPoint(i)
	}
	d.cleanup()
}

func (d *Delaunay) cleanup() {
	var clean []Triangle
	// Simple compaction
	for _, t := range d.Triangles {
		if !t.Active {
			continue
		}
		// Check super vertices
		isSuper := false
		for _, idx := range []int{t.A, t.B, t.C} {
			if idx == d.superIndices[0] || idx == d.superIndices[1] || idx == d.superIndices[2] {
				isSuper = true
				break
			}
		}
		if !isSuper {
			clean = append(clean, t)
		}
	}
	d.Triangles = clean
}
