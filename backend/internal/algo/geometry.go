package algo

// orient2d returns 2*SignedArea. Positive for counter-clockwise orientation.
// See docs/MATHEMATICS.md#11-orientation-point-in-triangle
func (d *Delaunay) orient2d(a, b, c Point) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

func (d *Delaunay) inCircumcircle(tIdx int, p Point) bool {
	t := d.Triangles[tIdx]
	a, b, c := d.Points[int(t.A)], d.Points[int(t.B)], d.Points[int(t.C)]

	// inCircumcircle tests if point is inside triangle's circumcircle.
	// Uses robust determinant-based predicate to avoid explicit circumcentre calculation.
	// See docs/MATHEMATICS.md#12-in-circle-test
	ax, ay := a.X-p.X, a.Y-p.Y
	bx, by := b.X-p.X, b.Y-p.Y
	cx, cy := c.X-p.X, c.Y-p.Y

	return (ax*ax+ay*ay)*(bx*cy-cx*by)-
		(bx*bx+by*by)*(ax*cy-cx*ay)+
		(cx*cx+cy*cy)*(ax*by-bx*ay) > EPSILON
}

func (d *Delaunay) contains(tIdx int, p Point) bool {
	t := d.Triangles[tIdx]
	return d.orient2d(d.Points[int(t.A)], d.Points[int(t.B)], p) >= -EPSILON &&
		d.orient2d(d.Points[int(t.B)], d.Points[int(t.C)], p) >= -EPSILON &&
		d.orient2d(d.Points[int(t.C)], d.Points[int(t.A)], p) >= -EPSILON
}
