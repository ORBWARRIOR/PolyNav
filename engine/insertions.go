package engine

import "math"

func InsertPoint(m *Mesh, point Point) *Mesh {

	return m
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

	// distance from c to p1
	dx := ux - p1.X
	dy := uy - p1.Y
	// c^2 = a^2 + b^2
	rSqrd = dx*dx + dy*dy

	return ux, uy, rSqrd, true
}

// PointInCircumcircle checks whether pt lies inside the circumcircle of (a, b, c).
func PointInCircumcircle(pt, a, b, c Point) bool {
	cx, cy, rSqrd, ok := GetCircumcircle(a, b, c)
	if !ok {
		return false
	}
	// distance from pt to circumcentre
	dx := pt.X - cx
	dy := pt.Y - cy
	// if rSqrd > pt's distance^2, it is inside the circumcircle
	return dx*dx+dy*dy <= rSqrd+Epsilon
}
