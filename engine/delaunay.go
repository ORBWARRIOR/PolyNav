package engine

import (
	"fmt"
	"math"
	"sort"
)

// Triangulate performs Delaunay triangulation via the Bowyer-Watson algorithm.
//
// Algorithm outline (implement as dummy stubs — fill in the real math):
//
//  1. Input normalisation
//     normalisePoints() — scale all points to 1x1 square
//     deduplicatePoints() — remove coincident points
//
//  2. Super-triangle
//     Create a triangle large enough to contain all points.
//     { (-100,-100), (+100,-100), (0,+100) } in normalised space.
//     The super-triangle vertices are appended at the end of the working slice.
//
//  3. Incremental insertion (Bowyer-Watson)
//     For each input point P:
//     a) Find all "bad" triangles whose circumcircle contains P
//     (use GetCircumcircle)
//     b) The union of bad triangles forms a polygon cavity.
//     Collect all edges of bad triangles.
//     c) Remove interior edges (edges shared by two bad triangles).
//     The remaining edges form the boundary of the cavity.
//     d) Connect P to every vertex of the cavity boundary.
//     This forms the new triangulation.
//
//  4. Super-triangle cleanup
//     Remove any triangle that shares a vertex with the super-triangle.
//
//  5. Output assembly
//     Build the Mesh: copy surviving triangles, extract unique edges,
//     identify boundary (hull) edges.
func Triangulate(points []Point) (*Mesh, error) {
	if len(points) < 3 {
		return nil, fmt.Errorf("delaunay: need at least 3 points, got %d", len(points))
	}

	uniq := deduplicatePoints(normalisePoints(points))
	if len(uniq) < 3 {
		return nil, fmt.Errorf("delaunay: only %d unique points after dedup, need 3", len(uniq))
	}

	// stub: returns a single triangle from the first 3 unique points
	return &Mesh{
		Points:   uniq,
		Indices:  []int{0, 1, 2},
		Edges:    []Edge{{0, 1}, {1, 2}, {2, 0}},
		Boundary: []Edge{{0, 1}, {1, 2}, {2, 0}},
	}, nil
}

// GetCircumcircle returns the circumcentre (ux, uy) and squared radius rSqrd
// for the triangle formed by p1, p2, p3.  ok=false when the points are
// collinear (denominator ≈ 0).
func GetCircumcircle(p1, p2, p3 Point) (ux, uy, rSqrd float64, ok bool) {
	p1Sq := p1.X*p1.X + p1.Y*p1.Y
	p2Sq := p2.X*p2.X + p2.Y*p2.Y
	p3Sq := p3.X*p3.X + p3.Y*p3.Y

	d := 2 * (p1.X*(p2.Y-p3.Y) + p2.X*(p3.Y-p1.Y) + p3.X*(p1.Y-p2.Y))
	if math.Abs(d) < Epsilon {
		return 0, 0, 0, false
	}

	ux = (p1Sq*(p2.Y-p3.Y) + p2Sq*(p3.Y-p1.Y) + p3Sq*(p1.Y-p2.Y)) / d
	uy = (p1Sq*(p3.X-p2.X) + p2Sq*(p1.X-p3.X) + p3Sq*(p2.X-p1.X)) / d

	dx := ux - p1.X
	dy := uy - p1.Y
	rSqrd = dx*dx + dy*dy

	return ux, uy, rSqrd, true
}

// PointInCircumcircle checks whether pt lies inside the circumcircle of (a, b, c).
func PointInCircumcircle(pt, a, b, c Point) bool {
	cx, cy, rsq, ok := GetCircumcircle(a, b, c)
	if !ok {
		return false
	}
	dx := pt.X - cx
	dy := pt.Y - cy
	return dx*dx+dy*dy <= rsq+Epsilon
}

// MeshStats computes summary statistics from a Mesh.
func MeshStats(m *Mesh) Stats {
	// Dummy — count from the data structure.
	triCount := len(m.Indices) / 3
	return Stats{
		PointCount:    len(m.Points),
		TriangleCount: triCount,
		EdgeCount:     len(m.Edges),
		HullEdges:     len(m.Boundary),
	}
}

func normalisePoints(points []Point) []Point {
	minX, minY := math.MaxFloat64, math.MaxFloat64
	maxX, maxY := -math.MaxFloat64, -math.MaxFloat64

	for _, pt := range points {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}

	scale := math.Max(maxX-minX, maxY-minY)
	if scale <= Epsilon {
		scale = 1.0
	}

	normalised := make([]Point, len(points))
	for i, pt := range points {
		normalised[i] = Point{
			X: (pt.X - minX) / scale,
			Y: (pt.Y - minY) / scale,
		}
	}
	return normalised
}

func deduplicatePoints(points []Point) []Point {
	sort.Slice(points, func(i, j int) bool {
		if math.Abs(points[i].X-points[j].X) > Epsilon {
			return points[i].X < points[j].X
		}
		return points[i].Y < points[j].Y-Epsilon
	})

	uidx := 0
	for i := 1; i < len(points); i++ {
		if !isCoincident(points[uidx], points[i]) {
			uidx++
			points[uidx] = points[i]
		}
	}
	return points[:uidx+1]
}

func isCoincident(p1, p2 Point) bool {
	return math.Abs(p1.X-p2.X) < Epsilon && math.Abs(p1.Y-p2.Y) < Epsilon
}
