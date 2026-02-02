package algo

import (
	"fmt"
)

// EdgeRef uniquely identifies an edge in the mesh.
// It points to Triangle[TIdx] and the edge opposite vertex EdgeIdx (0=A, 1=B, 2=C).
// e.g. EdgeIdx=0 means edge BC.
type EdgeRef struct {
	TIdx    int
	EdgeIdx int
}

// AddConstraint enforces a constrained edge between two points.
// If the points are not already in the mesh, they should be inserted first.
// This function assumes u and v are indices of existing points.
func (d *Delaunay) AddConstraint(u, v int) error {
	if u == v {
		return nil
	}
	for {
		edges, splitIdx, err := d.findIntersectingEdges(u, v)
		if err != nil {
			return err
		}

		// If we hit a vertex, split the constraint
		if splitIdx != -1 {
			// Constraint u-v is split into u-splitIdx and splitIdx-v
			if err := d.AddConstraint(u, splitIdx); err != nil {
				return err
			}
			return d.AddConstraint(splitIdx, v)
		}

		if len(edges) == 0 {
			break
		}

		err = d.resolveIntersections(u, v, edges)
		if err != nil {
			return err
		}
	}

	// Finally, mark the resulting edge as constrained
	d.markConstraint(u, v)
	return nil
}

func (d *Delaunay) resolveIntersections(uIdx, vIdx int, edges []EdgeRef) error {
	flipped := false
	for _, e := range edges {
		if d.isConvex(e) {
			d.flipEdge(e.TIdx, d.getNeighborIdx(e))
			flipped = true
			break
		}
	}

	// If we couldn't flip anything but still have edges, something is wrong
	if !flipped && len(edges) > 0 {
		// Sloan (1993) suggests this shouldn't happen for valid input.
		// However, we'll return an error to avoid infinite loop.
		return fmt.Errorf("failed to resolve all intersections: stuck")
	}

	return nil
}

func (d *Delaunay) getNeighborIdx(e EdgeRef) int {
	t := d.Triangles[e.TIdx]
	if e.EdgeIdx == 0 {
		return int(t.T1)
	}
	if e.EdgeIdx == 1 {
		return int(t.T2)
	}
	return int(t.T3)
}

func (d *Delaunay) isConvex(e EdgeRef) bool {
	nIdx := d.getNeighborIdx(e)
	if nIdx == -1 {
		return false
	}

	t := d.Triangles[e.TIdx]
	n := d.Triangles[nIdx]

	// Vertices of shared edge
	var u, v Point
	if e.EdgeIdx == 0 {
		u, v = d.Points[t.B], d.Points[t.C]
	} else if e.EdgeIdx == 1 {
		u, v = d.Points[t.C], d.Points[t.A]
	} else {
		u, v = d.Points[t.A], d.Points[t.B]
	}

	// Opposite vertices
	p := d.Points[[3]int32{t.A, t.B, t.C}[e.EdgeIdx]]

	var nSlot int
	if int(n.T1) == e.TIdx {
		nSlot = 0
	} else if int(n.T2) == e.TIdx {
		nSlot = 1
	} else {
		nSlot = 2
	}
	q := d.Points[[3]int32{n.A, n.B, n.C}[nSlot]]

	// A quad is convex if the opposite vertices p and q lie on opposite sides of uv,
	// AND u and v lie on opposite sides of pq.
	// Relaxed check: Allow collinearity/degeneracy to avoid getting stuck.
	return d.orient2d(u, v, p)*d.orient2d(u, v, q) <= EPSILON &&
		d.orient2d(p, q, u)*d.orient2d(p, q, v) <= EPSILON
}

func (d *Delaunay) markConstraint(u, v int) {
	// Find the triangle(s) sharing edge uv and set Constrained bit
	for i, t := range d.Triangles {
		if !t.Active {
			continue
		}

		// Edge 0: BC, 1: CA, 2: AB
		if (int(t.B) == u && int(t.C) == v) || (int(t.B) == v && int(t.C) == u) {
			d.Triangles[i].Constrained[0] = true
		} else if (int(t.C) == u && int(t.A) == v) || (int(t.C) == v && int(t.A) == u) {
			d.Triangles[i].Constrained[1] = true
		} else if (int(t.A) == u && int(t.B) == v) || (int(t.A) == v && int(t.B) == u) {
			d.Triangles[i].Constrained[2] = true
		}
	}
}

// ClassifyRegions identifies triangles inside and outside the constrained polygons.
// It assumes constraints form closed loops.
func (d *Delaunay) ClassifyRegions() {
	// 1. Identify "Seed" triangles on the convex hull boundary.
	// These are guaranteed to be "Outside" if the polygon is internal.
	// In our case, we'll mark everything as Inside=true by default,
	// then flood fill from the boundary to mark Inside=false.

	for i := range d.Triangles {
		d.Triangles[i].Inside = true
	}

	queue := []int{}
	visited := make(map[int]bool)

	for i, t := range d.Triangles {
		if !t.Active {
			continue
		}
		// If it has a boundary neighbor, it's a candidate for "Outside"
		// ONLY if the boundary edge is NOT constrained.
		if (t.T1 == -1 && !t.Constrained[0]) ||
			(t.T2 == -1 && !t.Constrained[1]) ||
			(t.T3 == -1 && !t.Constrained[2]) {
			queue = append(queue, i)
			visited[i] = true
			d.Triangles[i].Inside = false
		}
	}

	// 2. BFS Flood Fill
	for len(queue) > 0 {
		currIdx := queue[0]
		queue = queue[1:]

		t := d.Triangles[currIdx]
		neighbors := [3]int32{t.T1, t.T2, t.T3}

		for i, nIdx := range neighbors {
			if nIdx == -1 {
				continue
			}
			if visited[int(nIdx)] {
				continue
			}

			// If the edge separating us from the neighbor is constrained,
			// we do NOT cross it.
			if t.Constrained[i] {
				continue
			}

			// Mark neighbor as outside and add to queue
			visited[int(nIdx)] = true
			d.Triangles[int(nIdx)].Inside = false
			queue = append(queue, int(nIdx))
		}
	}

	// 3. Remove outside triangles
	d.filterTriangles()
}

func (d *Delaunay) filterTriangles() {
	// Map old index to new index
	newIndices := make([]int32, len(d.Triangles))
	for i := range newIndices {
		newIndices[i] = -1
	}

	activeCount := 0
	for i, t := range d.Triangles {
		if t.Active && t.Inside {
			newIndices[i] = int32(activeCount)
			activeCount++
		}
	}

	newTriangles := make([]Triangle, 0, activeCount)
	for i, t := range d.Triangles {
		if newIndices[i] != -1 {
			updateN := func(n int32) int32 {
				if n == -1 {
					return -1
				}
				return newIndices[n]
			}
			t.T1 = updateN(t.T1)
			t.T2 = updateN(t.T2)
			t.T3 = updateN(t.T3)
			newTriangles = append(newTriangles, t)
		}
	}
	d.Triangles = newTriangles
}

// findIntersectingEdges finds all edges in the triangulation that intersect the segment uv.
// If the segment strictly passes through a vertex 'k', it returns splitIdx = k.
func (d *Delaunay) findIntersectingEdges(u, v int) ([]EdgeRef, int, error) {
	if u == v {
		return nil, -1, nil
	}
	pU := d.Points[u]
	pV := d.Points[v]

	// 1. Find a triangle incident to u (using scanning)
	var firstIntersectingEdge *EdgeRef

	for i, t := range d.Triangles {
		if !t.Active {
			continue
		}

		var idxA, idxB int32
		var oppEdgeIdx int

		if int(t.A) == u {
			idxA, idxB = t.B, t.C
			oppEdgeIdx = 0
		} else if int(t.B) == u {
			idxA, idxB = t.C, t.A
			oppEdgeIdx = 1
		} else if int(t.C) == u {
			idxA, idxB = t.A, t.B
			oppEdgeIdx = 2
		} else {
			continue
		}

		pa, pb := d.Points[idxA], d.Points[idxB]

		// Check if v lies in the cone
		if d.orient2d(pU, pa, pV) >= -EPSILON && d.orient2d(pU, pb, pV) <= EPSILON {
			firstIntersectingEdge = &EdgeRef{TIdx: i, EdgeIdx: oppEdgeIdx}
			break
		}
	}

	if firstIntersectingEdge == nil {
		return nil, -1, fmt.Errorf("could not find starting triangle for segment %d-%d", u, v)
	}

	// 3. Walk along ray
	intersectingEdges := []EdgeRef{}
	currEdge := *firstIntersectingEdge

	for k := 0; k < len(d.Triangles); k++ {
		t := d.Triangles[currEdge.TIdx]

		// Get vertices of the edge we are about to cross
		var p1Idx, p2Idx int32
		if currEdge.EdgeIdx == 0 {
			p1Idx, p2Idx = t.B, t.C
		} else if currEdge.EdgeIdx == 1 {
			p1Idx, p2Idx = t.C, t.A
		} else {
			p1Idx, p2Idx = t.A, t.B
		}

		// Check if we hit the target v
		if int(p1Idx) == v || int(p2Idx) == v {
			return intersectingEdges, -1, nil
		}

		intersectingEdges = append(intersectingEdges, currEdge)

		var neighborIdx int32
		if currEdge.EdgeIdx == 0 {
			neighborIdx = t.T1
		} else if currEdge.EdgeIdx == 1 {
			neighborIdx = t.T2
		} else {
			neighborIdx = t.T3
		}
		if neighborIdx == -1 {
			return nil, -1, fmt.Errorf("hit boundary before reaching v")
		}

		nT := d.Triangles[neighborIdx]
		var entryEdgeIdx int
		if (nT.B == p1Idx && nT.C == p2Idx) || (nT.B == p2Idx && nT.C == p1Idx) {
			entryEdgeIdx = 0
		} else if (nT.C == p1Idx && nT.A == p2Idx) || (nT.C == p2Idx && nT.A == p1Idx) {
			entryEdgeIdx = 1
		} else {
			entryEdgeIdx = 2
		}

		// Vertex opposite to entry edge in neighbor
		var oppositeVertexIdx int32
		if entryEdgeIdx == 0 {
			oppositeVertexIdx = nT.A
		} else if entryEdgeIdx == 1 {
			oppositeVertexIdx = nT.B
		} else {
			oppositeVertexIdx = nT.C
		}

		// Check if this vertex lies on the segment uv
		if int(oppositeVertexIdx) != v && int(oppositeVertexIdx) != u && d.pointOnSegment(d.Points[oppositeVertexIdx], pU, pV) {
			return intersectingEdges, int(oppositeVertexIdx), nil
		}

		// Determine which exit edge to take
		e1Idx := (entryEdgeIdx + 1) % 3
		e2Idx := (entryEdgeIdx + 2) % 3
		getEdgeVerts := func(idx int, tri Triangle) (Point, Point) {
			if idx == 0 {
				return d.Points[tri.B], d.Points[tri.C]
			}
			if idx == 1 {
				return d.Points[tri.C], d.Points[tri.A]
			}
			return d.Points[tri.A], d.Points[tri.B]
		}

		pA1, pB1 := getEdgeVerts(e1Idx, nT)
		pA2, pB2 := getEdgeVerts(e2Idx, nT)

		int1 := segmentsIntersect(pU, pV, pA1, pB1)
		int2 := segmentsIntersect(pU, pV, pA2, pB2)

		if int1 {
			currEdge = EdgeRef{TIdx: int(neighborIdx), EdgeIdx: e1Idx}
		} else if int2 {
			currEdge = EdgeRef{TIdx: int(neighborIdx), EdgeIdx: e2Idx}
		} else {
			// Robustness: If neither edge strictly intersects, we must be hitting the vertex
			// between them (oppositeVertexIdx). Even if pointOnSegment strict check failed.

			// Prevent infinite recursion if we hit endpoints
			if int(oppositeVertexIdx) == v {
				return intersectingEdges, -1, nil
			}
			if int(oppositeVertexIdx) == u {
				return nil, -1, fmt.Errorf("walk circled back to start")
			}

			return intersectingEdges, int(oppositeVertexIdx), nil
		}
	}
	return nil, -1, fmt.Errorf("walk limit exceeded")
}

func (d *Delaunay) pointOnSegment(p, a, b Point) bool {
	// Check collinearity
	if d.orient2d(a, b, p) < -EPSILON || d.orient2d(a, b, p) > EPSILON {
		return false
	}
	// Check bounding box / betweenness
	// Dot product (p-a).(b-a) should be between 0 and |b-a|^2
	dp := (p.X-a.X)*(b.X-a.X) + (p.Y-a.Y)*(b.Y-a.Y)
	if dp < -EPSILON {
		return false
	}
	lenSq := (b.X-a.X)*(b.X-a.X) + (b.Y-a.Y)*(b.Y-a.Y)
	if dp > lenSq+EPSILON {
		return false
	}
	return true
}

func segmentsIntersect(a, b, c, d Point) bool {
	o1, o2 := orient(a, b, c), orient(a, b, d)
	o3, o4 := orient(c, d, a), orient(c, d, b)
	return ((o1 > 0 && o2 < 0) || (o1 < 0 && o2 > 0)) && ((o3 > 0 && o4 < 0) || (o3 < 0 && o4 > 0))
}

func orient(a, b, c Point) float64 {
	return (b.X-a.X)*(c.Y-a.Y) - (b.Y-a.Y)*(c.X-a.X)
}
