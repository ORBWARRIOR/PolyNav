package algo

import "math"

// ExportGraph generates the Voronoi diagram as a navigation graph.
// Each Delaunay triangle becomes a graph node at its circumcentre.
// See docs/MATHEMATICS.md#2-voronoi-duality and #3-circumcentre-calculation
func (d *Delaunay) ExportGraph() map[int]*GraphNode {
	graph := make(map[int]*GraphNode)

	for i, t := range d.Triangles {
		if !t.Active {
			continue
		}

		p1, p2, p3 := d.Points[int(t.A)], d.Points[int(t.B)], d.Points[int(t.C)]

		D := 2 * (p1.X*(p2.Y-p3.Y) + p2.X*(p3.Y-p1.Y) + p3.X*(p1.Y-p2.Y))
		if math.Abs(D) < EPSILON {
			continue // Skip degenerate triangles (collinear vertices)
		}

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

		addNeigh := func(nIdx int) {
			if nIdx != -1 && nIdx < len(d.Triangles) && d.Triangles[nIdx].Active {
				node.Neighbors = append(node.Neighbors, nIdx)
			}
		}
		addNeigh(int(t.T1))
		addNeigh(int(t.T2))
		addNeigh(int(t.T3))

		graph[i] = node
	}
	return graph
}
