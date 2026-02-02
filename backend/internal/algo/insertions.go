package algo

import "math"

// insertPoint implements Sloan's optimised insertion with edge splitting.
// See docs/ALGORITHMS.md#14-degeneracy-handling for edge cases.
// See docs/ALGORITHMS.md#2-edge-flipping-lawsons-flip for legalisation.
func (d *Delaunay) insertPoint(pIdx int) {
	p := d.Points[pIdx]

	// 1. Locate containing triangle using walking search
	tIdx := d.walkLocate(p, d.lastCreated)

	if tIdx == -1 {
		// Fallback to linear scan if walking search fails
		tIdx = -1
		for i, t := range d.Triangles {
			if t.Active && d.contains(i, p) {
				tIdx = i
				break
			}
		}
		if tIdx == -1 {

			return
		}
	}

	// 2. Handle degenerate case: point on edge
	t := d.Triangles[tIdx]
	pA, pB, pC := d.Points[int(t.A)], d.Points[int(t.B)], d.Points[int(t.C)]

	if math.Abs(d.orient2d(pA, pB, p)) < EPSILON {
		d.splitEdge(pIdx, tIdx, int(t.T3), int(t.A), int(t.B), int(t.C))
		return
	}
	if math.Abs(d.orient2d(pB, pC, p)) < EPSILON {
		d.splitEdge(pIdx, tIdx, int(t.T1), int(t.B), int(t.C), int(t.A))
		return
	}
	if math.Abs(d.orient2d(pC, pA, p)) < EPSILON {
		d.splitEdge(pIdx, tIdx, int(t.T2), int(t.C), int(t.A), int(t.B))
		return
	}

	// 3. Normal case: point inside triangle (1-to-3 split)
	d.Triangles[tIdx].Active = false

	a, b, c := t.A, t.B, t.C
	n1, n2, n3 := t.T1, t.T2, t.T3

	newT1Idx := len(d.Triangles)
	newT2Idx := newT1Idx + 1
	newT3Idx := newT1Idx + 2

	// T1: BC-P
	d.Triangles = append(d.Triangles, Triangle{
		A: b, B: c, C: int32(pIdx),
		T1: int32(newT2Idx), T2: int32(newT3Idx), T3: n1,
		Active: true,
	})
	// T2: CA-P
	d.Triangles = append(d.Triangles, Triangle{
		A: c, B: a, C: int32(pIdx),
		T1: int32(newT3Idx), T2: int32(newT1Idx), T3: n2,
		Active: true,
	})
	// T3: AB-P
	d.Triangles = append(d.Triangles, Triangle{
		A: a, B: b, C: int32(pIdx),
		T1: int32(newT1Idx), T2: int32(newT2Idx), T3: n3,
		Active: true,
	})

	d.lastCreated = newT1Idx

	d.updateNeighbor(int(n1), tIdx, newT1Idx)
	d.updateNeighbor(int(n2), tIdx, newT2Idx)
	d.updateNeighbor(int(n3), tIdx, newT3Idx)

	d.legaliseEdge(newT1Idx, int(n1))
	d.legaliseEdge(newT2Idx, int(n2))
	d.legaliseEdge(newT3Idx, int(n3))
}

// splitEdge implements 1-to-4 split for points on shared edges.
// Detailed algorithm and edge case handling in docs/ALGORITHMS.md#14-degeneracy-handling.
func (d *Delaunay) splitEdge(pIdx, tIdx, nIdx, u, v, o int) {
	d.Triangles[tIdx].Active = false

	// Identify neighbours for edges opposite each vertex in triangle T

	// Re-fetch T to get neighbour indices correctly
	t := d.Triangles[tIdx]
	var n_vo, n_ou int32
	if int(t.A) == u && int(t.B) == v { // C is o. Edge vo is opp u (T.T1? No T.T1 opp A).
		// T: A=u, B=v, C=o.
		// Opp A (u): edge BC (vo) -> T1
		// Opp B (v): edge CA (ou) -> T2
		n_vo = t.T1
		n_ou = t.T2
	} else if int(t.B) == u && int(t.C) == v { // A is o
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
	// Create 2 new triangles for T: (p, v, o) and (u, p, o)
	t1Idx := len(d.Triangles) // (p, v, o)
	t2Idx := t1Idx + 1        // (u, p, o)

	d.Triangles = append(d.Triangles, Triangle{
		A: int32(pIdx), B: int32(v), C: int32(o),
		T1: n_vo, T2: int32(t2Idx), T3: -1, // T3 will be N's new tri
		Active: true,
	})

	d.Triangles = append(d.Triangles, Triangle{
		A: int32(u), B: int32(pIdx), C: int32(o),
		T1: int32(t1Idx), T2: n_ou, T3: -1, // T3 will be N's new tri
		Active: true,
	})

	d.lastCreated = t1Idx
	d.updateNeighbor(int(n_vo), tIdx, t1Idx)
	d.updateNeighbor(int(n_ou), tIdx, t2Idx)

	// Handle Neighbor N
	if nIdx != -1 {
		d.Triangles[nIdx].Active = false
		n := d.Triangles[nIdx]

		// Find o_n (opposite vertex in N)
		// N shares edge v-u (reversed).
		var o_n int32
		var n_uo_n, n_o_nv int32

		// Identify o_n and neighbours
		if int(n.A) == v && int(n.B) == u {
			o_n = n.C
			n_uo_n = n.T1 // Opp v: u-o_n
			n_o_nv = n.T2 // Opp u: o_n-v
		} else if int(n.B) == v && int(n.C) == u {
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
			A: int32(pIdx), B: int32(u), C: o_n,
			T1: n_uo_n, T2: int32(n2Idx), T3: int32(t2Idx),
			Active: true,
		})

		// N2: (v, p, o_n). Edges: p-o_n (n1Idx), o_n-v (n_o_nv), v-p (shared with T1)
		d.Triangles = append(d.Triangles, Triangle{
			A: int32(v), B: int32(pIdx), C: o_n,
			T1: int32(n1Idx), T2: n_o_nv, T3: int32(t1Idx),
			Active: true,
		})

		// Link back T's undefined neighbours
		d.Triangles[t1Idx].T3 = int32(n2Idx)
		d.Triangles[t2Idx].T3 = int32(n1Idx)

		d.updateNeighbor(int(n_uo_n), nIdx, n1Idx)
		d.updateNeighbor(int(n_o_nv), nIdx, n2Idx)

		d.legaliseEdge(n1Idx, int(n_uo_n))
		d.legaliseEdge(n2Idx, int(n_o_nv))
	} else {
		// Boundary case: T3 of new triangles remains -1
	}

	d.legaliseEdge(t1Idx, int(n_vo))
	d.legaliseEdge(t2Idx, int(n_ou))
}

// walkLocate implements Sloan's Directed Walk for point location.
// Complexity: O(N^0.5) average vs O(N) linear scan.
// See docs/ALGORITHMS.md#12-point-location-sloans-walk
func (d *Delaunay) walkLocate(p Point, startIdx int) int {
	// Validate start index bounds
	if startIdx < 0 || startIdx >= len(d.Triangles) {
		return -1
	}

	curr := startIdx
	limit := len(d.Triangles) // Safety brake

	for k := 0; k < limit; k++ {
		// Validate current triangle index bounds
		if curr < 0 || curr >= len(d.Triangles) {
			return -1
		}

		if !d.Triangles[curr].Active {
			// If we land on a dead triangle (rare in standard walk, but possible in edge cases),
			// revert to linear scan.
			return -1
		}

		t := d.Triangles[curr]
		pA, pB, pC := d.Points[int(t.A)], d.Points[int(t.B)], d.Points[int(t.C)]

		// Check which edge separates P from the triangle.
		// Orientation < 0 means P is to the Right (outside).
		// We walk toward the neighbour opposite that edge.

		if d.orient2d(pB, pC, p) < -EPSILON {
			// P is right of BC. Move to neighbour T1.
			if t.T1 == -1 {
				return curr
			} // P is outside hull, but this is the closest boundary.
			curr = int(t.T1)
		} else if d.orient2d(pC, pA, p) < -EPSILON {
			// P is right of CA. Move to neighbour T2.
			if t.T2 == -1 {
				return curr
			}
			curr = int(t.T2)
		} else if d.orient2d(pA, pB, p) < -EPSILON {
			// P is right of AB. Move to neighbour T3.
			if t.T3 == -1 {
				return curr
			}
			curr = int(t.T3)
		} else {
			// P is left of or ON all edges -> Inside.
			return curr
		}
	}
	return -1
}

// legaliseEdge performs Lawson's Flip to restore Delaunay property.
// See docs/ALGORITHMS.md#2-edge-flipping-lawsons-flip
func (d *Delaunay) legaliseEdge(tIdx, nIdx int) {
	if nIdx == -1 {
		return
	}

	n := d.Triangles[nIdx]

	// Identify shared edge and opposite vertex for edge flip check
	var nSlot int // 0, 1, 2 for A, B, C
	if int(n.T1) == tIdx {
		nSlot = 0
	} else if int(n.T2) == tIdx {
		nSlot = 1
	} else {
		nSlot = 2
	}

	// Vertex opposite shared edge in N
	qIdx := [3]int32{n.A, n.B, n.C}[nSlot]
	q := d.Points[int(qIdx)]

	// Check if edge needs flipping using in-circle test
	if d.inCircumcircle(tIdx, q) {
		d.flipEdge(tIdx, nIdx)

		// Recursive Legalise
		// The neighbours to check are the ones that were "outer" edges of the quad
		// flipEdge updates T1, T2, T3. We need to check them.
		// After flip:
		// T: (p, u, q). Neighbors: nT1 (opp u), nIdx (opp p, ignored), tT3 (opp q).
		// N: (p, q, v). Neighbors: nT2 (opp p), tT2 (opp q), tIdx (opp v, ignored).
		
		// Wait, finding the indices of the neighbors to check is slightly annoying without context.
		// But I can just check all 4 external neighbors?
		// Or I can calculate them inside flipEdge and return them?
		// Simpler: Just check the neighbors of tIdx and nIdx, ignoring nIdx and tIdx respectively.
		
		d.legaliseEdge(tIdx, int(d.Triangles[tIdx].T1))
		d.legaliseEdge(tIdx, int(d.Triangles[tIdx].T3))
		
		d.legaliseEdge(nIdx, int(d.Triangles[nIdx].T1))
		d.legaliseEdge(nIdx, int(d.Triangles[nIdx].T2))
	}
}

// flipEdge performs the topological flip of the edge shared by tIdx and nIdx.
// Assumes they are valid neighbors.
func (d *Delaunay) flipEdge(tIdx, nIdx int) {
	t := d.Triangles[tIdx]
	n := d.Triangles[nIdx]

	// Find slots
	var tSlot int
	if int(t.T1) == nIdx { tSlot = 0 } else if int(t.T2) == nIdx { tSlot = 1 } else { tSlot = 2 }

	var nSlot int
	if int(n.T1) == tIdx { nSlot = 0 } else if int(n.T2) == tIdx { nSlot = 1 } else { nSlot = 2 }

	// Vertices
	// T: (p, u, v) where p is opposite shared edge
	// N: (q, v, u) where q is opposite shared edge (Order in N is reversed? No, CCW).
	// Shared edge is u-v.
	// T has u-v as edge opposite p.
	// N has v-u as edge opposite q.
	
	pIdx := [3]int32{t.A, t.B, t.C}[tSlot]
	uIdx := [3]int32{t.A, t.B, t.C}[(tSlot+1)%3]
	vIdx := [3]int32{t.A, t.B, t.C}[(tSlot+2)%3]
	
	qIdx := [3]int32{n.A, n.B, n.C}[nSlot]

	// Neighbors
	nT1 := [3]int32{n.T1, n.T2, n.T3}[(nSlot+1)%3] // Neighbor opp v in N
	nT2 := [3]int32{n.T1, n.T2, n.T3}[(nSlot+2)%3] // Neighbor opp u in N
	tT2 := [3]int32{t.T1, t.T2, t.T3}[(tSlot+1)%3] // Neighbor opp v in T
	tT3 := [3]int32{t.T1, t.T2, t.T3}[(tSlot+2)%3] // Neighbor opp u in T

	// Update T: A=p, B=u, C=q
	d.Triangles[tIdx].A = pIdx
	d.Triangles[tIdx].B = uIdx
	d.Triangles[tIdx].C = qIdx
	d.Triangles[tIdx].T1 = nT1  // Edge u-q is now boundary with what was n's neighbour
	d.Triangles[tIdx].T2 = int32(nIdx) // Edge q-p is shared with N
	d.Triangles[tIdx].T3 = tT3  // Edge p-u preserved

	// Update N: A=p, B=q, C=v
	d.Triangles[nIdx].A = pIdx
	d.Triangles[nIdx].B = qIdx
	d.Triangles[nIdx].C = vIdx
	d.Triangles[nIdx].T1 = nT2
	d.Triangles[nIdx].T2 = tT2
	d.Triangles[nIdx].T3 = int32(tIdx)

	// Update outer pointers
	d.updateNeighbor(int(nT1), nIdx, tIdx)
	d.updateNeighbor(int(tT2), tIdx, nIdx)
}

func (d *Delaunay) updateNeighbor(tIdx, oldN, newN int) {
	if tIdx == -1 {
		return
	}
	t := &d.Triangles[tIdx]
	if int(t.T1) == oldN {
		t.T1 = int32(newN)
		return
	}
	if int(t.T2) == oldN {
		t.T2 = int32(newN)
		return
	}
	if int(t.T3) == oldN {
		t.T3 = int32(newN)
		return
	}
}