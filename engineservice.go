package main

import (
	"fmt"

	"github.com/ORBWARRIOR/PolyNav/engine"
)

// EngineService exposes engine/ algorithms to the Wails frontend.
// Each exported method becomes an async callable in TypeScript via the
// auto-generated bindings (run `make bindings` after adding methods).
type EngineService struct{}

func (g *EngineService) Log(msg string) {
	fmt.Println("Frontend sent message:", msg)
}

// Triangulate runs Delaunay triangulation and returns a serialisable result.
func (s *EngineService) Triangulate(points []engine.Point) (*engine.Mesh, error) {
	return engine.Triangulate(points)
}
