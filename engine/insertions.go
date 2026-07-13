package engine

import (
	"fmt"
)

func InsertPoint(m *Mesh, p Point) (*Mesh, error) {
	pID := m.addPoint(p)
	edge, ok := walkToPoint(m, p)
	if !ok {
		return m, fmt.Errorf("failed to walk to point %v", p)
	}

	// Determine if P lies inside or on the edge of the triangle
	onEdge := false
	for range 3 {
		nextEdge := m.HalfEdges[edge].Next
		a := m.HalfEdges[edge].Origin
		b := m.HalfEdges[nextEdge].Origin
		if orient(m.Points[a], m.Points[b], p) < Epsilon { // Negligible area, treat as collinear
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
	for _, edge := range newEdges {
		legaliseEdge(m, edge, pID)
	}

	m.LastInsertedEdge = newEdges[0]
	return m, nil
}

// Returns an edge of the triangle bounding P
func walkToPoint(m *Mesh, p Point) (EdgeID, bool) {
	currentEdge := m.LastInsertedEdge
	maxIter := len(m.Triangles)
	for range maxIter {
		crossed := false

		for range 3 {

			currentHE := m.HalfEdges[currentEdge]
			if tID := currentHE.Triangle; tID == NoneTriangle || m.Triangles[tID].Tombstoned {
				currentEdge = m.HalfEdges[currentHE.Twin].Next // Do not traverse deleted triangles, jump to neighbour
				crossed = true
				break
			}
			nextEdge := currentHE.Next
			a := currentHE.Origin
			b := m.HalfEdges[nextEdge].Origin

			if orient(m.Points[a], m.Points[b], p) < -Epsilon { // Negative area significant of Epsilon, RHS
				twin := currentHE.Twin // Jump to neighbour
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
func orient(a, b, c Point) float64 {
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

	// Rewire the new triangles
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

func splitEdge(m *Mesh, AB EdgeID, p VertexID) [4]EdgeID {
	BC := m.HalfEdges[AB].Next
	CA := m.HalfEdges[BC].Next
	BA := m.HalfEdges[AB].Twin
	AD := m.HalfEdges[BA].Next
	DB := m.HalfEdges[AD].Next

	A := m.HalfEdges[AB].Origin
	B := m.HalfEdges[BC].Origin
	C := m.HalfEdges[CA].Origin
	D := m.HalfEdges[DB].Origin

	if tID := m.HalfEdges[AB].Triangle; tID != NoneTriangle {
		m.Triangles[tID].Tombstoned = true
	}
	if tID := m.HalfEdges[BA].Triangle; tID != NoneTriangle {
		m.Triangles[tID].Tombstoned = true
	}

	APC := m.addTriangle(AB)
	CPB := m.addTriangle(BC)
	DPA := m.addTriangle(BA)
	BPD := m.addTriangle(DB)

	PB, BP := m.addEdgePair(p, B)
	PC, CP := m.addEdgePair(p, C)
	DP, PD := m.addEdgePair(D, p)

	// Repurpose AB to AP					\\       	A--------C
	m.HalfEdges[AB].Origin = A //			\\       	 \      |
	m.HalfEdges[AB].Next = PC  //			\\       	  \    |
	m.HalfEdges[AB].Twin = BA  //			\\       	   \  |
	// Rewire the triangle					\\       	    \|
	m.HalfEdges[PC].Next = CA      //		\\       	     P
	m.HalfEdges[CA].Next = AB      //		\\       	      \
	m.HalfEdges[AB].Triangle = APC //		\\       	       \
	m.HalfEdges[PC].Triangle = APC //		\\       	        \
	m.HalfEdges[CA].Triangle = APC //		\\       	         B
	//										\\
	// Repurpose BA to PA					\\       	A
	m.HalfEdges[BA].Origin = p //			\\       	\\
	m.HalfEdges[BA].Next = AD  //			\\       	\ \
	m.HalfEdges[BA].Twin = AB  //			\\       	\  \
	// Rewire the triangle  				\\       	\   \
	m.HalfEdges[AD].Next = DP      //		\\       	\    P
	m.HalfEdges[DP].Next = BA      //		\\       	\   | \
	m.HalfEdges[BA].Triangle = DPA //		\\       	\  |   \
	m.HalfEdges[AD].Triangle = DPA //		\\       	\ |     \
	m.HalfEdges[DP].Triangle = DPA //		\\       	 D       B
	//										\\
	//										\\       			  C
	m.HalfEdges[PB].Next = BC      //		\\          		 /\
	m.HalfEdges[BC].Next = CP      //		\\       			/ \
	m.HalfEdges[CP].Next = PB      //		\\       		   /  \
	m.HalfEdges[PB].Triangle = CPB //		\\       		  P   \
	m.HalfEdges[BC].Triangle = CPB //		\\       		   \  \
	m.HalfEdges[CP].Triangle = CPB //		\\       			\ \
	//										\\                   \\
	m.HalfEdges[BP].Next = PD      //		\\       			  B
	m.HalfEdges[PD].Next = DB      //		\\
	m.HalfEdges[DB].Next = BP      //		\\       	    P
	m.HalfEdges[BP].Triangle = BPD //		\\       	   | \
	m.HalfEdges[PD].Triangle = BPD //		\\       	  |   \
	m.HalfEdges[DB].Triangle = BPD //		\\       	 |     \
	// 										\\       	D-------B
	return [4]EdgeID{AB, BP, CP, DP}
}
