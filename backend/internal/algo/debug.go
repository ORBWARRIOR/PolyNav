package algo

import (
	"encoding/json"
	"fmt"
)

// DebugJSON exports the current mesh state as GeoJSON for visualisation.
// Useful for debugging triangulation results with GIS tools.
func (d *Delaunay) DebugJSON() (string, error) {
	type Geometry struct {
		Type        string        `json:"type"`
		Coordinates [][][]float64 `json:"coordinates"`
	}

	type Feature struct {
		Type       string                 `json:"type"`
		Geometry   Geometry               `json:"geometry"`
		Properties map[string]interface{} `json:"properties"`
	}

	type FeatureCollection struct {
		Type     string    `json:"type"`
		Features []Feature `json:"features"`
	}

	fc := FeatureCollection{
		Type:     "FeatureCollection",
		Features: make([]Feature, 0, len(d.Triangles)),
	}

	for i, t := range d.Triangles {

		if !t.Active {
			continue
		}

		p1 := d.Points[int(t.A)]
		p2 := d.Points[int(t.B)]
		p3 := d.Points[int(t.C)]

		coords := [][][]float64{{
			{p1.X, p1.Y},
			{p2.X, p2.Y},
			{p3.X, p3.Y},
			{p1.X, p1.Y}, // Closing the loop
		}}

		fc.Features = append(fc.Features, Feature{
			Type: "Feature",
			Geometry: Geometry{
				Type:        "Polygon",
				Coordinates: coords,
			},
			Properties: map[string]interface{}{
				"id":         i,
				"neighbours": []int{int(t.T1), int(t.T2), int(t.T3)},
			},
		})
	}

	bytes, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "%s"}`, err.Error()), err
	}
	return string(bytes), nil
}
