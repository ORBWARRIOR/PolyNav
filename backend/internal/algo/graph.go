package algo

import "math"

// ExportGraph generates the Voronoi Graph (Dual of Delaunay).
// See docs/MATHEMATICS.md#3-circumcenter-calculation
func (d *Delaunay) ExportGraph() map[int]*GraphNode {
	graph := make(map[int]*GraphNode)

	for i, t := range d.Triangles {
		if !t.Active {
			continue
		}

		// Calculate Circumcenter
		p1, p2, p3 := d.Points[t.A], d.Points[t.B], d.Points[t.C]

		D := 2 * (p1.X*(p2.Y-p3.Y) + p2.X*(p3.Y-p1.Y) + p3.X*(p1.Y-p2.Y))
		if math.Abs(D) < EPSILON {
			continue
		} // Degenerate triangle

		// See docs/MATHEMATICS.md#3-circumcenter-calculation
		Ux := ((p1.X*p1.X+p1.Y*p1.Y)*(p2.Y-p3.Y) +
			(p2.X*p2.X+p2.Y*p2.Y)*(p3.Y-p1.Y) +
			(p3.X*p3.X+p3.Y*p3.Y)*(p1.Y-p2.Y)) / D

		Uy := ((p1.X*p1.X+p1.Y*p1.Y)*(p3.X-p2.X) +
			(p2.X*p2.X+p2.Y*p2.Y)*(p1.X-p3.X) +
			(p3.X*p3.X+p3.Y*p3.Y)*(p2.X-p1.X)) / D

		node := &GraphNode{
			ID:        i,
			X:         Ux,
			Y:         Uy,
			Neighbors: []int{},
		}

		// Add neighbors if they are active and valid
		addNeigh := func(nIdx int) {
			if nIdx != -1 && d.Triangles[nIdx].Active {
				node.Neighbors = append(node.Neighbors, nIdx)
			}
		}
		addNeigh(t.T1)
		addNeigh(t.T2)
		addNeigh(t.T3)

		graph[i] = node
	}
	return graph
}
