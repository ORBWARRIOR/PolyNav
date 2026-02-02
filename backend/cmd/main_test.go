package main

import (
	"context"
	"testing"
	"time"

	"github.com/ORBWARRIOR/PolyNav/backend/cmd/server"
	pb "github.com/ORBWARRIOR/PolyNav/backend/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const testAddress = ":50052"

func TestIntegrationTriangulate(t *testing.T) {
	// Start the Server
	srv, err := server.NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	errorC := make(chan error, 1)
	go srv.Run(testAddress, errorC)

	defer srv.Shutdown()

	time.Sleep(100 * time.Millisecond)

	// Check if server crashed immediately
	select {
	case err := <-errorC:
		t.Fatalf("Server failed to start: %v", err)
	default:
	}

	// Create Client
	conn, err := grpc.NewClient("localhost"+testAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	client := pb.NewGeometryServiceClient(conn)

	// Test Case: Simple Square
	req := &pb.MapData{
		Obstacles: []*pb.Obstacle{
			{
				Points: []*pb.Point{
					{X: 0, Y: 0},
					{X: 10, Y: 0},
					{X: 10, Y: 10},
					{X: 0, Y: 10},
				},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.Triangulate(ctx, req)
	if err != nil {
		t.Fatalf("Triangulate RPC failed: %v", err)
	}

	if len(resp.Triangles) != 2 {
		t.Errorf("Expected 2 triangles, got %d", len(resp.Triangles))
	}
}

func TestIntegrationEmptyRequest(t *testing.T) {
	// Start Server
	srv, err := server.NewServer()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}
	errorC := make(chan error, 1)
	go srv.Run(testAddress, errorC)
	defer srv.Shutdown()
	time.Sleep(100 * time.Millisecond)

	// Create Client
	conn, err := grpc.NewClient("localhost"+testAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial server: %v", err)
	}
	defer conn.Close()

	client := pb.NewGeometryServiceClient(conn)

	// Test Case: Empty Map
	req := &pb.MapData{Obstacles: []*pb.Obstacle{}}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := client.Triangulate(ctx, req)
	if err != nil {
		t.Fatalf("RPC failed on empty input: %v", err)
	}

	if len(resp.Triangles) != 0 {
		t.Errorf("Expected 0 triangles for empty map, got %d", len(resp.Triangles))
	}
}
