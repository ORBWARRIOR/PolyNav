package algo

import (
	"encoding/json"
	"fmt"
)

// DebugJSON exports the current mesh state as simple JSON for visualisation.
// Useful for debugging triangulation results.
func (d *Delaunay) DebugJSON() (string, error) {
	type Point struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
	}

	type TriangleData struct {
		ID         int     `json:"id"`
		Points     []Point `json:"points"`
		Neighbours []int   `json:"neighbours"`
	}

	triangles := make([]TriangleData, 0, len(d.Triangles))

	for i, t := range d.Triangles {
		if !t.Active {
			continue
		}

		p1 := d.Points[int(t.A)]
		p2 := d.Points[int(t.B)]
		p3 := d.Points[int(t.C)]

		triangles = append(triangles, TriangleData{
			ID: i,
			Points: []Point{
				{X: p1.X, Y: p1.Y},
				{X: p2.X, Y: p2.Y},
				{X: p3.X, Y: p3.Y},
			},
			Neighbours: []int{int(t.T1), int(t.T2), int(t.T3)},
		})
	}

	bytes, err := json.MarshalIndent(triangles, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error()), err
	}
	return string(bytes), nil
}
