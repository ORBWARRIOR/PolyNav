package server

import (
	"context"
	"log"

	"github.com/ORBWARRIOR/PolyNav/backend/internal/algo"
	pb "github.com/ORBWARRIOR/PolyNav/backend/pkg/proto"
)

type server struct {
	pb.UnimplementedGeometryServiceServer
}

func (s *server) Triangulate(ctx context.Context, in *pb.MapData) (*pb.TriangulationResult, error) {
	log.Printf("Received Triangulate request with %d obstacles", len(in.Obstacles))

	var algoPoints []algo.Point
	for _, obs := range in.Obstacles {
		for _, p := range obs.Points {
			algoPoints = append(algoPoints, algo.Point{X: p.X, Y: p.Y})
		}
	}

	// Add start and goal points to triangulation if they exist
	if in.Start != nil {
		algoPoints = append(algoPoints, algo.Point{X: in.Start.X, Y: in.Start.Y})
	}
	if in.Goal != nil {
		algoPoints = append(algoPoints, algo.Point{X: in.Goal.X, Y: in.Goal.Y})
	}

	if len(algoPoints) < 3 {
		return &pb.TriangulationResult{}, nil
	}

	dt, err := algo.NewDelaunay(algoPoints)
	if err != nil {
		log.Printf("Delaunay initialization failed: %v", err)
		return nil, err
	}
	dt.Triangulate()

	var triangles []*pb.Triangle
	for _, t := range dt.Triangles {
		p1 := dt.Points[t.A]
		p2 := dt.Points[t.B]
		p3 := dt.Points[t.C]

		triangles = append(triangles, &pb.Triangle{
			A: &pb.Point{X: p1.X, Y: p1.Y},
			B: &pb.Point{X: p2.X, Y: p2.Y},
			C: &pb.Point{X: p3.X, Y: p3.Y},
		})
	}

	return &pb.TriangulationResult{Triangles: triangles}, nil
}

func (s *server) SaveMap(ctx context.Context, in *pb.MapData) (*pb.SaveMapResponse, error) {
	log.Printf("Received SaveMap request")
	return &pb.SaveMapResponse{Success: true, Message: "Map saved successfully (mock)"}, nil
}
