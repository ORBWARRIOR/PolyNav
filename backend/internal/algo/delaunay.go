package algo

import (
	"errors"
	"math"
	"sort"
)

// NewDelaunay initialises the mesh with a Super Triangle ensuring convex hull coverage.
func NewDelaunay(points []Point) (*Delaunay, error) {
	// Input validation for NaN/Inf values
	if err := validatePoints(points); err != nil {
		return nil, err
	}

	uniquePoints := deduplicatePoints(points)

	if len(uniquePoints) < 3 {
		return nil, errors.New("insufficient points (needs 3+ unique points)")

	}

	// OPTIMIZATION: "Unidimensional Sorting" approximates spatial binning.
	// Sorting by X-coordinate improves walking search locality, keeping runtime
	// closer to O(N^5/4) without implementing full binning.
	// See docs/ALGORITHMS.md#13-deviations-from-sloans-1987-algorithm
	sort.Slice(uniquePoints, func(i, j int) bool { return uniquePoints[i].X < uniquePoints[j].X })

	// Preallocate with factor 2.5*N (Tuned based on experimental churn)
	// See docs/MATHEMATICS.md#4-memory-allocation-eulers-formula
	d := &Delaunay{
		Points:    make([]Point, 0, len(uniquePoints)+3),
		Triangles: make([]Triangle, 0, int(float64(len(uniquePoints))*2.5)+100),
	}

	d.Points = append(d.Points, uniquePoints...)

	// 1.1.2 Super Triangle (See docs/ALGORITHMS.md#11-the-super-triangle)
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, p := range uniquePoints {
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
	deltaMax := math.Max(dx, dy) * 10.0
	midX, midY := (minX+maxX)/2.0, (minY+maxY)/2.0

	p1 := Point{midX - deltaMax, midY - deltaMax}
	p2 := Point{midX + deltaMax, midY - deltaMax}
	p3 := Point{midX, midY + deltaMax}

	idx1, idx2, idx3 := len(d.Points), len(d.Points)+1, len(d.Points)+2
	d.Points = append(d.Points, p1, p2, p3)
	d.superIndices = [3]int{idx1, idx2, idx3}

	// Create Root Triangle
	d.Triangles = append(d.Triangles, Triangle{
		A: int32(idx1), B: int32(idx2), C: int32(idx3),
		T1: -1, T2: -1, T3: -1,
		Active: true,
	})
	d.lastCreated = 0

	return d, nil
}

// Triangulate executes Incremental Insertion with Lawson's Flip.
// See docs/ALGORITHMS.md#1-delaunay-triangulation-strategy
func (d *Delaunay) Triangulate() {
	originalCount := len(d.Points) - 3
	for i := 0; i < originalCount; i++ {
		d.insertPoint(i)
	}
	d.cleanup()
}

// Input validation prevents geometry predicate failures
func validatePoints(points []Point) error {
	for _, p := range points {
		if math.IsNaN(p.X) || math.IsNaN(p.Y) {
			return errors.New("point contains NaN value")
		}
		if math.IsInf(p.X, 0) || math.IsInf(p.Y, 0) {
			return errors.New("point contains infinite value")
		}
	}
	return nil
}

// Remove duplicate coordinates using epsilon comparison
func deduplicatePoints(points []Point) []Point {
	if len(points) == 0 {
		return nil
	}
	// Sort points to group potential duplicates
	sorted := make([]Point, len(points))
	copy(sorted, points)
	sort.Slice(sorted, func(i, j int) bool {
		if math.Abs(sorted[i].X-sorted[j].X) > EPSILON {
			return sorted[i].X < sorted[j].X
		}
		return sorted[i].Y < sorted[j].Y
	})

	result := make([]Point, 0, len(points))
	result = append(result, sorted[0])

	for i := 1; i < len(sorted); i++ {
		curr := sorted[i]
		prev := result[len(result)-1]

		if math.Abs(curr.X-prev.X) > EPSILON || math.Abs(curr.Y-prev.Y) > EPSILON {
			result = append(result, curr)
		}
	}
	return result
}

func (d *Delaunay) cleanup() {
	// In-place compaction to remove triangles connected to Super Triangle
	n := 0
	for _, t := range d.Triangles {
		if !t.Active {
			continue
		}

		isSuper := false
		for _, idx := range []int{int(t.A), int(t.B), int(t.C)} {
			if idx == d.superIndices[0] || idx == d.superIndices[1] || idx == d.superIndices[2] {
				isSuper = true
				break
			}
		}
		if !isSuper {
			d.Triangles[n] = t
			n++
		}
	}
	d.Triangles = d.Triangles[:n]
}
