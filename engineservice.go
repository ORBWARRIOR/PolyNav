package main

import (
	"github.com/ORBWARRIOR/PolyNav/engine"
)

// EngineService exposes engine/ algorithms to the Wails frontend.
// Each exported method becomes an async callable in TypeScript via the
// auto-generated bindings (run `make bindings` after adding methods).
type EngineService struct {
	eng *engine.Engine
}

func NewEngineService() *EngineService {
	return &EngineService{eng: engine.NewEngine()}
}

// Triangulate runs Delaunay triangulation and returns a serialisable result.
// The frontend receives triangles as flat index arrays ready for WebGL / Canvas.
func (s *EngineService) Triangulate(points []engine.Point) (*engine.Mesh, error) {
	return engine.Triangulate(points)
}
