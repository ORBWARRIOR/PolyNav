package engine

import (
	"fmt"
	"math"
)

func InsertPoint(m *Mesh, p Point, pID VertexID) (*Mesh, error) {
	edge, ok := WalkToPoint(m, p)
	if !ok {
		return m, fmt.Errorf("Failed to walk to point %v", p)
	}

	onEdge := false
	for i := 0; i < 3; i++ {
		nextEdge := m.HalfEdges[edge].Next
		a := m.HalfEdges[edge].Origin
		b := m.HalfEdges[nextEdge].Origin
		if Orient(m.Points[a], m.Points[b], p) < Epsilon { // Negligible area, treat as collinear
			onEdge = true
			break
		}
		edge = nextEdge
	}

	var newEdges []EdgeID
	if onEdge {
		result := splitEdge(m, edge, pID)
		newEdges = result[:]
	} else {
		result := splitTriangle(m, edge, pID)
		newEdges = result[:]
	}
	fmt.Println(newEdges)

	return m, nil
}

// Returns an edge of the triangle bounding P
func WalkToPoint(m *Mesh, p Point) (EdgeID, bool) {
	currentEdge := m.LastInsertedEdge
	maxIter := len(m.Triangles)
	for i := 0; i < maxIter; i++ {
		crossed := false
		for i := 0; i < 3; i++ {
			nextEdge := m.HalfEdges[currentEdge].Next
			a := m.HalfEdges[currentEdge].Origin
			b := m.HalfEdges[nextEdge].Origin

			if Orient(m.Points[a], m.Points[b], p) < -Epsilon { // Negative area significant of Epsilon, RHS
				twin := m.HalfEdges[currentEdge].Twin // Jump to neighbour
				if twin == NoneEdge {
					return 0, false
				}
				// P cannot be right of twin, preemptively walk to next edge
				currentEdge = m.HalfEdges[twin].Next
				crossed = true
				break
			}
			currentEdge = nextEdge
		}
		// If no jump occurs, we have encountered the triangle bounding point P
		if !crossed {
			return currentEdge, true
		}
	}
	// failed to walk
	return 0, false
}

// returns the signed area of a triangle * 2
func Orient(a, b, c Point) float64 {
	// the cross product tells us if a point is on LHS or RHS of an edge
	// Positive = LHS, Negative = RHS, -Epsilon < x < Epsilon = collinear
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}

func splitTriangle(m *Mesh, AB EdgeID, p VertexID) [3]EdgeID {
	// get the original triangle VertexIDs and EdgeIDs
	BC := m.HalfEdges[AB].Next
	CA := m.HalfEdges[BC].Next
	a := m.HalfEdges[AB].Origin
	b := m.HalfEdges[BC].Origin
	c := m.HalfEdges[CA].Origin
	m.Triangles[m.HalfEdges[AB].Triangle].Tombstoned = true

	// create 3 new triangles
	ABP := m.addTriangle(AB)
	BCP := m.addTriangle(BC)
	CAP := m.addTriangle(CA)

	// create 3 new half edges and their twins
	BP, PB := m.addEdgePair(b, p)
	CP, PC := m.addEdgePair(c, p)
	AP, PA := m.addEdgePair(a, p)

	// assign edge properties to each edge
	m.HalfEdges[AB].Next = BP
	m.HalfEdges[BP].Next = PA
	m.HalfEdges[PA].Next = AB
	m.HalfEdges[AB].Triangle = ABP
	m.HalfEdges[BP].Triangle = ABP
	m.HalfEdges[PA].Triangle = ABP

	m.HalfEdges[BC].Next = CP
	m.HalfEdges[CP].Next = PB
	m.HalfEdges[PB].Next = BC
	m.HalfEdges[BC].Triangle = BCP
	m.HalfEdges[CP].Triangle = BCP
	m.HalfEdges[PB].Triangle = BCP

	m.HalfEdges[CA].Next = AP
	m.HalfEdges[AP].Next = PC
	m.HalfEdges[PC].Next = CA
	m.HalfEdges[CA].Triangle = CAP
	m.HalfEdges[AP].Triangle = CAP
	m.HalfEdges[PC].Triangle = CAP

	return [3]EdgeID{BP, CP, AP}
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
