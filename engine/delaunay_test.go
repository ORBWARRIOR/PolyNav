// Blackbox integration tests for the engine/delaunay package.
//
// "Blackbox" means package engine_test — we import engine as an external
// consumer would, testing the exported API only.  This catches interface
// drift and proves the public surface works as documented without relying
// on internal details.
//
// These are "integration" tests in the sense that they exercise real data
// flow through the normalise → deduplicate → triangulate pipeline, not
// that they spin up the Wails runtime.  Full GUI integration lives in
// ../tests/ with a build tag.

package engine_test

import (
	"testing"

	"github.com/ORBWARRIOR/PolyNav/engine"
)

// ---------------------------------------------------------------------------
// Triangulate
// ---------------------------------------------------------------------------

func TestTriangulateThreePointsReturnsOneTriangle(t *testing.T) {
	points := []engine.Point{
		{X: 0, Y: 0},
		{X: 100, Y: 0},
		{X: 0, Y: 100},
	}

	mesh, err := engine.Triangulate(points)
	if err != nil {
		t.Fatalf("Triangulate failed: %v", err)
	}
	if mesh == nil {
		t.Fatal("Triangulate returned nil mesh")
	}

	length := len(mesh.Points)
	if length != 3 {
		t.Errorf("Mesh mismatch. Expected length 3, got: %d", length)
	}
}

func TestTriangulateTwoPointsReturnsError(t *testing.T) {
	points := []engine.Point{
		{X: 0, Y: 0},
		{X: 1, Y: 1},
	}

	_, err := engine.Triangulate(points)
	if err == nil {
		t.Fatal("expected error for < 3 points, got nil")
	}
}

func TestTriangulateEmptyReturnsError(t *testing.T) {
	_, err := engine.Triangulate([]engine.Point{})
	if err == nil {
		t.Fatal("expected error for empty slice, got nil")
	}
}

// ---------------------------------------------------------------------------
// Circumcircle
// ---------------------------------------------------------------------------

func TestGetCircumcircleRightTriangle(t *testing.T) {
	cx, cy, rsq, ok := engine.GetCircumcircle(
		engine.Point{X: 0, Y: 0},
		engine.Point{X: 4, Y: 0},
		engine.Point{X: 0, Y: 4},
	)
	if !ok {
		t.Fatal("GetCircumcircle returned not-ok for a valid triangle")
	}
	if cx != 2 || cy != 2 {
		t.Fatalf("expected centre (2,2), got (%.2f,%.2f)", cx, cy)
	}
	expectedRSq := 8.0 // (2-0)² + (2-0)² = 8
	if rsq != expectedRSq {
		t.Fatalf("expected rsq=8, got %.2f", rsq)
	}
}

func TestGetCircumcircleCollinearReturnsNotOk(t *testing.T) {
	_, _, _, ok := engine.GetCircumcircle(
		engine.Point{X: 0, Y: 0},
		engine.Point{X: 1, Y: 1},
		engine.Point{X: 2, Y: 2},
	)
	if ok {
		t.Fatal("expected not-ok for collinear points")
	}
}

// ---------------------------------------------------------------------------
// MeshStats
// ---------------------------------------------------------------------------

func TestMeshStatsEmptyMesh(t *testing.T) {
	m := &engine.Mesh{}
	s := engine.MeshStats(m)
	if s.TriangleCount != 0 {
		t.Fatalf("expected 0 triangles, got %d", s.TriangleCount)
	}
}

func TestTriangulateDedupCoincident(t *testing.T) {
	// All points are the same — after dedup there are fewer than 3 distinct.
	points := []engine.Point{
		{X: 5, Y: 5},
		{X: 5, Y: 5},
		{X: 5, Y: 5},
	}
	_, err := engine.Triangulate(points)
	if err == nil {
		t.Fatal("expected error when all points are coincident")
	}
}
