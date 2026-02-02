package algo

import (
	"encoding/json"
	"os"
	"testing"
)

type DragonMap struct {
	Points   []struct{ X, Y float64 } `json:"points"`
	Segments [][]int                  `json:"segments"`
}

func TestAddConstraint(t *testing.T) {
	// Setup a square
	points := []Point{
		{0, 0},   // A
		{10, 0},  // B
		{10, 10}, // C
		{0, 10},  // D
	}

	d, err := NewDelaunay(points)
	if err != nil {
		t.Fatalf("Failed to init Delaunay: %v", err)
	}

	d.Triangulate()

	// Helper to find index by coordinate
	findIdx := func(p Point) int {
		for i, dp := range d.Points {
			if dp.X == p.X && dp.Y == p.Y {
				return i
			}
		}
		return -1
	}

	idx00 := findIdx(Point{0, 0})
	idx1010 := findIdx(Point{10, 10})

	// Before constraint: findIntersectingEdges for (0,0)-(10,10) might be 0 or 1
	// depending on which diagonal Delaunay chose.
	// But after AddConstraint, it MUST be 0.
	err = d.AddConstraint(idx00, idx1010)
	if err != nil {
		t.Fatalf("AddConstraint failed: %v", err)
	}

	// Verify intersection is now 0
	edges, _, err := d.findIntersectingEdges(idx00, idx1010)
	if err != nil {
		t.Fatalf("findIntersectingEdges failed after constraint: %v", err)
	}
	if len(edges) != 0 {
		t.Errorf("Expected 0 intersecting edges after forcing constraint, got %d", len(edges))
	}

	// Verify the edge is marked as constrained
	found := false
	for _, tri := range d.Triangles {
		if !tri.Active {
			continue
		}
		// Check all 3 edges
		for i := 0; i < 3; i++ {
			if tri.Constrained[i] {
				// Get vertices of this edge
				var u, v int32
				if i == 0 {
					u, v = tri.B, tri.C
				} else if i == 1 {
					u, v = tri.C, tri.A
				} else {
					u, v = tri.A, tri.B
				}

				if (int(u) == idx00 && int(v) == idx1010) || (int(u) == idx1010 && int(v) == idx00) {
					found = true
				}
			}
		}
	}
	if !found {
		t.Error("Constraint was not found marked as true in any triangle")
	}
}

func TestDragonMap(t *testing.T) {
	// Read Dragon Map Data from file
	data, err := os.ReadFile("../../../dragon_map.json")
	if err != nil {
		t.Fatalf("Failed to read dragon_map.json: %v", err)
	}

	var dMap DragonMap
	if err := json.Unmarshal(data, &dMap); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// 2. Convert to algo.Point
	points := make([]Point, len(dMap.Points))
	for i, p := range dMap.Points {
		points[i] = Point{X: p.X, Y: p.Y}
	}

	// 3. Initialize Delaunay
	d, err := NewDelaunay(points)
	if err != nil {
		t.Fatalf("NewDelaunay failed: %v", err)
	}
	d.Triangulate()

	// 4. Helper to find exact index in d.Points (since they might be reordered/deduplicated)
	findIdx := func(originalIdx int) int {
		op := dMap.Points[originalIdx]
		for i, dp := range d.Points {
			if dp.X == op.X && dp.Y == op.Y {
				return i
			}
		}
		return -1
	}

	// 5. Add Constraints
	hasErrors := false
	for i, seg := range dMap.Segments {
		uIdx := findIdx(seg[0])
		vIdx := findIdx(seg[1])

		if uIdx == -1 || vIdx == -1 {
			t.Logf("Warning: Point not found for segment %d: %v -> %v", i, seg[0], seg[1])
			continue
		}

		// Self-loops or zero length check
		if uIdx == vIdx {
			continue
		}

		err := d.AddConstraint(uIdx, vIdx)
		if err != nil {
			t.Errorf("Failed to add constraint for segment %d (%d-%d): %v", i, uIdx, vIdx, err)
			hasErrors = true
		}
	}

	// 6. Carve Regions
	if !hasErrors {
		d.ClassifyRegions()
	} else {
		t.Log("Skipping ClassifyRegions due to constraint errors")
	}

	// 7. Verify Result
	activeCount := 0
	for _, t := range d.Triangles {
		if t.Active {
			activeCount++
		}
	}

	t.Logf("Dragon Map Triangulation complete. Active triangles: %d", activeCount)

	if activeCount == 0 {
		t.Error("Dragon Map resulted in 0 triangles after ClassifyRegions")
	}

	// 8. Save Debug Output
	jsonStr, _ := d.DebugJSON()
	_ = os.WriteFile("./savedTests/debug_dragon_map.json", []byte(jsonStr), 0644)
}
