package server

import (
	"context"

	"github.com/ORBWARRIOR/PolyNav/backend/internal/algo"
	pb "github.com/ORBWARRIOR/PolyNav/backend/pkg/proto"
	"github.com/rs/zerolog/log"
)

type server struct {
	pb.UnimplementedGeometryServiceServer
}

func (s *server) Triangulate(ctx context.Context, in *pb.MapData) (*pb.TriangulationResult, error) {
	log.Info().Int("obstacles", len(in.Obstacles)).Msg("Received Triangulate request")

	var allPoints []algo.Point

	// Collect all points for triangulation
	for _, obs := range in.Obstacles {
		for _, p := range obs.Points {
			allPoints = append(allPoints, algo.Point{X: p.X, Y: p.Y})
		}
	}
	if in.Start != nil {
		allPoints = append(allPoints, algo.Point{X: in.Start.X, Y: in.Start.Y})
	}
	if in.Goal != nil {
		allPoints = append(allPoints, algo.Point{X: in.Goal.X, Y: in.Goal.Y})
	}

	if len(allPoints) < 3 {
		return &pb.TriangulationResult{}, nil
	}

	dt, err := algo.NewDelaunay(allPoints)
	if err != nil {
		log.Err(err).Msg("Delaunay initialisation failed")
		return nil, err
	}
	dt.Triangulate()

	// Optimization: Create a lookup map for O(1) access
	// We rely on the fact that input points are copied exactly.
	pointMap := make(map[algo.Point]int, len(dt.Points))
	for i, p := range dt.Points {
		pointMap[p] = i
	}

	// Helper to find index in dt.Points using the map
	getIdx := func(p *pb.Point) int {
		pt := algo.Point{X: p.X, Y: p.Y}
		if idx, exists := pointMap[pt]; exists {
			return idx
		}
		return -1
	}

	// Add Constraints for each obstacle (assuming they are closed loops)
	for _, obs := range in.Obstacles {
		if len(obs.Points) < 3 {
			continue
		}
		for i := 0; i < len(obs.Points); i++ {
			p1 := obs.Points[i]
			p2 := obs.Points[(i+1)%len(obs.Points)]

			idx1 := getIdx(p1)
			idx2 := getIdx(p2)

			if idx1 != -1 && idx2 != -1 {
				err := dt.AddConstraint(idx1, idx2)
				if err != nil {
					log.Warn().Err(err).Msg("Failed to add constraint")
				}
			}
		}
	}

	// Carve outside triangles
	dt.ClassifyRegions()

	var triangles []*pb.Triangle
	for _, t := range dt.Triangles {
		p1 := dt.Points[t.A]
		p2 := dt.Points[t.B]
		p3 := dt.Points[t.C]

		triangles = append(triangles, &pb.Triangle{
			A:                &pb.Point{X: p1.X, Y: p1.Y},
			B:                &pb.Point{X: p2.X, Y: p2.Y},
			C:                &pb.Point{X: p3.X, Y: p3.Y},
			ConstrainedEdges: []bool{t.Constrained[0], t.Constrained[1], t.Constrained[2]},
		})
	}

	return &pb.TriangulationResult{Triangles: triangles}, nil
}

func (s *server) SaveMap(ctx context.Context, in *pb.MapData) (*pb.SaveMapResponse, error) {
	log.Info().Msg("Received SaveMap request")
	return &pb.SaveMapResponse{Success: true, Message: "Map saved successfully (mock)"}, nil
}
