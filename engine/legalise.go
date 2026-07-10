package engine

import "math"

func legaliseEdge(m *Mesh, edge EdgeID, p VertexID) {
	if edge == NoneEdge || int(edge) > len(m.HalfEdges) {
		return
	}
	tID := m.HalfEdges[edge].Triangle
	if tID == NoneTriangle || m.Triangles[tID].Tombstoned {
		return
	}
	oppEdge := getEdgeOppositeP(m, edge, p)

	twin := m.HalfEdges[oppEdge].Twin
	if twin == NoneEdge {
		return
	}
	twinTID := m.HalfEdges[twin].Triangle
	if twinTID == NoneTriangle || m.Triangles[twinTID].Tombstoned {
		return
	}

	A := m.HalfEdges[oppEdge].Origin
	B := m.HalfEdges[twin].Origin
	C := m.HalfEdges[m.HalfEdges[m.HalfEdges[twin].Next].Next].Origin

	AC := NoneEdge
	CB := NoneEdge
	if pointInCircumcircle(m.Points[C], m.Points[p], m.Points[A], m.Points[B]) {
		AC = m.HalfEdges[twin].Next
		CB = m.HalfEdges[AC].Next
		flipEdge(m, oppEdge)
		legaliseEdge(m, AC, p)
		legaliseEdge(m, CB, p)
	}

}

func flipEdge(m *Mesh, AB EdgeID) {
	BA := m.HalfEdges[AB].Twin
	if BA == NoneEdge {
		return
	}

	BP := m.HalfEdges[AB].Next
	PA := m.HalfEdges[BP].Next
	AC := m.HalfEdges[BA].Next
	CB := m.HalfEdges[AC].Next
	if AB == NoneEdge || BP == NoneEdge || PA == NoneEdge || BA == NoneEdge || AC == NoneEdge || CB == NoneEdge {
		return
	}

	t1 := m.HalfEdges[AB].Triangle
	t2 := m.HalfEdges[BA].Triangle
	if t1 == NoneTriangle || m.Triangles[t1].Tombstoned || t2 == NoneTriangle || m.Triangles[t2].Tombstoned {
		return
	}

	P := m.HalfEdges[PA].Origin
	C := m.HalfEdges[CB].Origin

	m.Triangles[t1].Tombstoned = true
	m.Triangles[t2].Tombstoned = true

	t3 := m.addTriangle(BP)
	t4 := m.addTriangle(PA)

	// Repurpose AB to PC, rewire the triangle
	m.HalfEdges[BP].Next = AB
	m.HalfEdges[AB].Origin = P
	m.HalfEdges[AB].Next = CB
	m.HalfEdges[CB].Next = BP

	m.HalfEdges[BP].Triangle = t3
	m.HalfEdges[AB].Triangle = t3
	m.HalfEdges[CB].Triangle = t3

	// Repurpose BA to CP, reqire the triangle
	m.HalfEdges[AC].Next = BA
	m.HalfEdges[BA].Origin = C
	m.HalfEdges[BA].Next = PA
	m.HalfEdges[PA].Next = AC

	m.HalfEdges[AC].Triangle = t4
	m.HalfEdges[BA].Triangle = t4
	m.HalfEdges[PA].Triangle = t4
}

// Orients itself according to P, returns the edge opposite P
func getEdgeOppositeP(m *Mesh, edge EdgeID, p VertexID) EdgeID {
	// After insertion, edge is expected to be incident to P
	// From recursion, edge is expected to be the edge opposite P
	if he := m.HalfEdges[edge]; he.Origin == p {
		return he.Next
	} else if he2 := m.HalfEdges[he.Next]; he2.Origin == p {
		return he2.Next
	} else {
		return edge
	}
}

// PointInCircumcircle checks whether pt lies inside the circumcircle of (a, b, c).
func pointInCircumcircle(pt, a, b, c Point) bool {
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
