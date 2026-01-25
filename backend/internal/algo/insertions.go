package algo

import "math"

// insertPoint implements Sloan's optimised insertion.
func (d *Delaunay) insertPoint(pIdx int) {
	p := d.Points[pIdx]

	// 1. Locate Triangle
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
	}

	// 2. Robustness: Check for Point on Edge
	t := d.Triangles[tIdx]
	pA, pB, pC := d.Points[t.A], d.Points[t.B], d.Points[t.C]

	// Check edge AB
	if math.Abs(d.orient2d(pA, pB, p)) < EPSILON {
		d.splitEdge(pIdx, tIdx, t.T3, t.A, t.B, t.C)
		return
	}
	// Check edge BC
	if math.Abs(d.orient2d(pB, pC, p)) < EPSILON {
		d.splitEdge(pIdx, tIdx, t.T1, t.B, t.C, t.A)
		return
	}
	// Check edge CA
	if math.Abs(d.orient2d(pC, pA, p)) < EPSILON {
		d.splitEdge(pIdx, tIdx, t.T2, t.C, t.A, t.B)
		return
	}

	// 3. Normal Split (1-to-3)
	d.Triangles[tIdx].Active = false

	a, b, c := t.A, t.B, t.C
	n1, n2, n3 := t.T1, t.T2, t.T3

	newT1Idx := len(d.Triangles)
	newT2Idx := newT1Idx + 1
	newT3Idx := newT1Idx + 2

	// T1: BC-P
	d.Triangles = append(d.Triangles, Triangle{
		A: b, B: c, C: pIdx,
		T1: newT2Idx, T2: newT3Idx, T3: n1,
		Active: true,
	})
	// T2: CA-P
	d.Triangles = append(d.Triangles, Triangle{
		A: c, B: a, C: pIdx,
		T1: newT3Idx, T2: newT1Idx, T3: n2,
		Active: true,
	})
	// T3: AB-P
	d.Triangles = append(d.Triangles, Triangle{
		A: a, B: b, C: pIdx,
		T1: newT1Idx, T2: newT2Idx, T3: n3,
		Active: true,
	})

	d.lastCreated = newT1Idx

	d.updateNeighbor(n1, tIdx, newT1Idx)
	d.updateNeighbor(n2, tIdx, newT2Idx)
	d.updateNeighbor(n3, tIdx, newT3Idx)

	d.legaliseEdge(newT1Idx, n1)
	d.legaliseEdge(newT2Idx, n2)
	d.legaliseEdge(newT3Idx, n3)
}

// splitEdge handles the degenerate case where P is on an edge shared by T and N.
// u, v are the vertices of the shared edge. o is the opposite vertex in T.
func (d *Delaunay) splitEdge(pIdx, tIdx, nIdx, u, v, o int) {
	d.Triangles[tIdx].Active = false

	// Neighbors of T
	// We need to identify which neighbor in T corresponds to edge v-o and o-u.
	// T is (u, v, o). Edge opposite o is uv (Neighbor N, passed as nIdx).
	// Edge opposite u is vo. Edge opposite v is ou.
	
	// Re-fetch T to get neighbor indices correctly
	t := d.Triangles[tIdx]
	var n_vo, n_ou int
	if t.A == u && t.B == v { // C is o. Edge vo is opp u (T.T1? No T.T1 opp A).
		// T: A=u, B=v, C=o.
		// Opp A (u): edge BC (vo) -> T1
		// Opp B (v): edge CA (ou) -> T2
		n_vo = t.T1
		n_ou = t.T2
	} else if t.B == u && t.C == v { // A is o
		// T: A=o, B=u, C=v.
		// Opp B (u): edge CA (vo) -> T2
		// Opp C (v): edge AB (ou) -> T3
		n_vo = t.T2
		n_ou = t.T3
	} else { // t.C == u && t.A == v. B is o
		// T: A=v, B=o, C=u
		// Opp C (u): edge AB (vo) -> T3
		// Opp A (v): edge BC (ou) -> T1
		n_vo = t.T3
		n_ou = t.T1
	}

	// Create 2 triangles for T: (p, v, o) and (p, o, u)
	// T1: (p, v, o). Edges: vo (n_vo), op (newT2), pv (shared with N part 2? No, pv is on split edge)
	// We need to match orientation.
	// T was (u, v, o) CCW.
	// Split into (u, p, o) and (p, v, o).
	
	// New Indices
	t1Idx := len(d.Triangles)     // (p, v, o)
	t2Idx := t1Idx + 1            // (u, p, o)
	
	d.Triangles = append(d.Triangles, Triangle{
		A: pIdx, B: v, C: o,
		T1: n_vo, T2: t2Idx, T3: -1, // T3 will be N's new tri
		Active: true,
	})
	
	d.Triangles = append(d.Triangles, Triangle{
		A: u, B: pIdx, C: o,
		T1: t1Idx, T2: n_ou, T3: -1, // T3 will be N's new tri
		Active: true,
	})
	
	d.lastCreated = t1Idx
	d.updateNeighbor(n_vo, tIdx, t1Idx)
	d.updateNeighbor(n_ou, tIdx, t2Idx)

	// Handle Neighbor N
	if nIdx != -1 {
		d.Triangles[nIdx].Active = false
		n := d.Triangles[nIdx]
		
		// Find o_n (opposite vertex in N)
		// N shares edge v-u (reversed).
		var o_n int
		var n_uo_n, n_o_nv int
		
		// Identify o_n and neighbors
		if (n.A == v && n.B == u) {
			o_n = n.C
			n_uo_n = n.T1 // Opp v: u-o_n
			n_o_nv = n.T2 // Opp u: o_n-v
		} else if (n.B == v && n.C == u) {
			o_n = n.A
			n_uo_n = n.T2
			n_o_nv = n.T3
		} else { // n.C == v && n.A == u
			o_n = n.B
			n_uo_n = n.T3
			n_o_nv = n.T1
		}
		
		// Create 2 triangles for N: (p, u, o_n) and (p, o_n, v)
		// N was (v, u, o_n) CCW.
		// Split into (v, p, o_n) and (p, u, o_n).
		
		n1Idx := len(d.Triangles) // (p, u, o_n)
		n2Idx := n1Idx + 1        // (v, p, o_n)
		
		// N1: (p, u, o_n). Edges: u-o_n (n_uo_n), o_n-p (n2Idx), p-u (shared with T2)
		d.Triangles = append(d.Triangles, Triangle{
			A: pIdx, B: u, C: o_n,
			T1: n_uo_n, T2: n2Idx, T3: t2Idx,
			Active: true,
		})
		
		// N2: (v, p, o_n). Edges: p-o_n (n1Idx), o_n-v (n_o_nv), v-p (shared with T1)
		d.Triangles = append(d.Triangles, Triangle{
			A: v, B: pIdx, C: o_n,
			T1: n1Idx, T2: n_o_nv, T3: t1Idx,
			Active: true,
		})
		
		// Link back T's undefined neighbors
		d.Triangles[t1Idx].T3 = n2Idx
		d.Triangles[t2Idx].T3 = n1Idx
		
		d.updateNeighbor(n_uo_n, nIdx, n1Idx)
		d.updateNeighbor(n_o_nv, nIdx, n2Idx)
		
		d.legaliseEdge(n1Idx, n_uo_n)
		d.legaliseEdge(n2Idx, n_o_nv)
	} else {
		// Boundary case: T3 of new triangles remains -1
	}

	d.legaliseEdge(t1Idx, n_vo)
	d.legaliseEdge(t2Idx, n_ou)
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