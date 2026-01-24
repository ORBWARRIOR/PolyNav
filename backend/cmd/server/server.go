package server

import (
	"fmt"
	"net"

	pb "github.com/ORBWARRIOR/PolyNav/backend/pkg/proto"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

type GrpcServer struct {
	server *grpc.Server
}

func (s *GrpcServer) Run(webAddress string, errorChannel chan<- error) {
	defer func() {
		if r := recover(); r != nil {
			errorChannel <- fmt.Errorf("panic while running service: %v", r)
		}
	}()
	lis, err := net.Listen("tcp", webAddress)
	if err != nil {
		errorChannel <- fmt.Errorf("failed to listen: %v", err)
		return
	}
	if err := s.server.Serve(lis); err != nil {
		errorChannel <- fmt.Errorf("failed to serve: %v", err)
		return
	}
}

func (server *GrpcServer) Shutdown() {
	log.Info().Msg("Shutting down gRPC server...")
	server.server.GracefulStop()
}

func NewServer() (*GrpcServer, error) {
	s := grpc.NewServer()
	pb.RegisterGeometryServiceServer(s, &server{})
	return &GrpcServer{server: s}, nil
}
