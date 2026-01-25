package algo

import "math"

// insertPoint implements Sloan's optimised insertion.
func (d *Delaunay) insertPoint(pIdx int) {
	p := d.Points[pIdx]

	// 1. Locate Triangle (See docs/ALGORITHMS.md#12-point-location-sloans-walk)
	// Start search at the most recently created triangle to exploit locality.
	tIdx := d.walkLocate(p, d.lastCreated)

	if tIdx == -1 {
		// Fallback: linear scan
		for i, t := range d.Triangles {
			if t.Active && d.contains(i, p) {
				tIdx = i
				break
			}
		}
	}

	if tIdx == -1 {
		return
	} // Point outside super-triangle

	// 2. Robustness Check: Degenerate Geometry
	// Check if point P lies on any edge of the found triangle.
	// Creating a triangle with area ~0 will cause NaN circumcenters.

	t := d.Triangles[tIdx]
	pA, pB, pC := d.Points[t.A], d.Points[t.B], d.Points[t.C]

	// orient2d returns ~0 if collinear.
	if math.Abs(d.orient2d(pA, pB, p)) < EPSILON ||
		math.Abs(d.orient2d(pB, pC, p)) < EPSILON ||
		math.Abs(d.orient2d(pC, pA, p)) < EPSILON {
		// DEGENERATE CASE: Point is on an existing edge.
		// Handling this correctly requires a 1-to-4 split (Edge Split).
		// For now, to prevent crashing, we REJECT this point.
		return
	}

	// 3. Split Triangle into 3
	// Mark old triangle as inactive

	d.Triangles[tIdx].Active = false

	a, b, c := t.A, t.B, t.C
	n1, n2, n3 := t.T1, t.T2, t.T3 // Neighbors opposite A, B, C

	// New indices
	newT1Idx := len(d.Triangles)
	newT2Idx := newT1Idx + 1
	newT3Idx := newT1Idx + 2

	// T1: Edge BC preserved (Neighbor n1)
	d.Triangles = append(d.Triangles, Triangle{
		A: b, B: c, C: pIdx,
		T1: newT2Idx, T2: newT3Idx, T3: n1,
		Active: true,
	})

	// T2: Edge CA preserved (Neighbor n2)
	d.Triangles = append(d.Triangles, Triangle{
		A: c, B: a, C: pIdx,
		T1: newT3Idx, T2: newT1Idx, T3: n2,
		Active: true,
	})

	// T3: Edge AB preserved (Neighbor n3)
	d.Triangles = append(d.Triangles, Triangle{
		A: a, B: b, C: pIdx,
		T1: newT1Idx, T2: newT2Idx, T3: n3,
		Active: true,
	})

	d.lastCreated = newT1Idx

	// Update outer neighbors to point to new triangles
	d.updateNeighbor(n1, tIdx, newT1Idx)
	d.updateNeighbor(n2, tIdx, newT2Idx)
	d.updateNeighbor(n3, tIdx, newT3Idx)

	// 3. Flip Edges (See docs/ALGORITHMS.md#2-edge-flipping-lawsons-flip)
	// Check the edges opposite the new point P in the newly created triangles
	d.legaliseEdge(newT1Idx, n1)
	d.legaliseEdge(newT2Idx, n2)
	d.legaliseEdge(newT3Idx, n3)
}

// walkLocate implements the Directed Walk (Sloan's TRILOC).
// Complexity: O(N^0.5) average, vs O(N) linear.
func (d *Delaunay) walkLocate(p Point, startIdx int) int {
	curr := startIdx
	limit := len(d.Triangles) // Safety brake

	for k := 0; k < limit; k++ {
		if !d.Triangles[curr].Active {
			// If we land on a dead triangle (rare in standard walk, but possible in edge cases),
			// revert to linear scan.
			return -1
		}

		t := d.Triangles[curr]
		pA, pB, pC := d.Points[t.A], d.Points[t.B], d.Points[t.C]

		// Check which edge separates P from the triangle.
		// Orientation < 0 means P is to the Right (outside).
		// We walk toward the neighbor opposite that edge.

		if d.orient2d(pB, pC, p) < -EPSILON {
			// P is right of BC. Move to neighbor T1.
			if t.T1 == -1 {
				return curr
			} // P is outside hull, but this is the closest boundary.
			curr = t.T1
		} else if d.orient2d(pC, pA, p) < -EPSILON {
			// P is right of CA. Move to neighbor T2.
			if t.T2 == -1 {
				return curr
			}
			curr = t.T2
		} else if d.orient2d(pA, pB, p) < -EPSILON {
			// P is right of AB. Move to neighbor T3.
			if t.T3 == -1 {
				return curr
			}
			curr = t.T3
		} else {
			// P is left of or ON all edges -> Inside.
			return curr
		}
	}
	return -1
}

// legaliseEdge ensures Delaunay property.
func (d *Delaunay) legaliseEdge(tIdx, nIdx int) {
	if nIdx == -1 {
		return
	}

	t := d.Triangles[tIdx]
	n := d.Triangles[nIdx]

	// Find shared edge indices
	// We need the vertex in N that is opposite the shared edge.
	var nSlot int // 0, 1, 2 for A, B, C
	if n.T1 == tIdx {
		nSlot = 0
	} else if n.T2 == tIdx {
		nSlot = 1
	} else {
		nSlot = 2
	}

	// Vertex opposite shared edge in N
	qIdx := [3]int{n.A, n.B, n.C}[nSlot]
	q := d.Points[qIdx]

	// Robust In-Circle Check
	if d.inCircumcircle(tIdx, q) {
		// FLIP
		// Find shared edge vertices in T
		var tSlot int
		if t.T1 == nIdx {
			tSlot = 0
		} else if t.T2 == nIdx {
			tSlot = 1
		} else {
			tSlot = 2
		}

		// Vertices
		// T: (p, u, v) where p is opposite shared edge
		// N: (q, v, u) where q is opposite shared edge
		// Result T: (p, u, q), N: (p, q, v)

		// Due to the complexity of index swapping, we simply rewrite the indices
		// consistent with CCW orientation.

		// Vertices of T
		pIdx := [3]int{t.A, t.B, t.C}[tSlot]
		uIdx := [3]int{t.A, t.B, t.C}[(tSlot+1)%3]
		vIdx := [3]int{t.A, t.B, t.C}[(tSlot+2)%3]

		// Neighbors
		nT1 := [3]int{n.T1, n.T2, n.T3}[(nSlot+1)%3] // Neighbor opp v in N
		nT2 := [3]int{n.T1, n.T2, n.T3}[(nSlot+2)%3] // Neighbor opp u in N
		tT2 := [3]int{t.T1, t.T2, t.T3}[(tSlot+1)%3] // Neighbor opp v in T
		tT3 := [3]int{t.T1, t.T2, t.T3}[(tSlot+2)%3] // Neighbor opp u in T

		// Update T: A=p, B=u, C=q
		d.Triangles[tIdx].A = pIdx
		d.Triangles[tIdx].B = uIdx
		d.Triangles[tIdx].C = qIdx
		d.Triangles[tIdx].T1 = nT1  // Edge u-q is now boundary with what was n's neighbor
		d.Triangles[tIdx].T2 = nIdx // Edge q-p is shared with N
		d.Triangles[tIdx].T3 = tT3  // Edge p-u preserved

		// Update N: A=p, B=q, C=v
		d.Triangles[nIdx].A = pIdx
		d.Triangles[nIdx].B = qIdx
		d.Triangles[nIdx].C = vIdx
		d.Triangles[nIdx].T1 = nT2
		d.Triangles[nIdx].T2 = tT2
		d.Triangles[nIdx].T3 = tIdx

		// Update outer pointers
		d.updateNeighbor(nT1, nIdx, tIdx)
		d.updateNeighbor(tT2, tIdx, nIdx)

		// Recursive Legalise
		d.legaliseEdge(tIdx, nT1)
		d.legaliseEdge(tIdx, tT3)
		d.legaliseEdge(nIdx, nT2)
		d.legaliseEdge(nIdx, tT2)
	}
}

func (d *Delaunay) updateNeighbor(tIdx, oldN, newN int) {
	if tIdx == -1 {
		return
	}
	t := &d.Triangles[tIdx]
	if t.T1 == oldN {
		t.T1 = newN
		return
	}
	if t.T2 == oldN {
		t.T2 = newN
		return
	}
	if t.T3 == oldN {
		t.T3 = newN
		return
	}
}
