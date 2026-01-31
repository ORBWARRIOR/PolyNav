package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ORBWARRIOR/PolyNav/backend/cmd/server"
	"github.com/rs/zerolog/log"
)

func main() {
	if err := setup(); err != nil {
		log.Err(err).Msg("server failed, exiting with code 1")
		os.Exit(1)
	}
}

func setup() (err error) {
	defer func() {
		if err != nil {
			log.Err(err).Msg("stopping with error")
		} else {
			log.Info().Msg("stopping")
		}
	}()

	webAddress := os.Getenv("GRPC_PORT")
	if webAddress == "" {
		webAddress = ":50051"
	}

	grpcServer, err := server.NewServer()
	if err != nil {
		return fmt.Errorf("failed to create gRPC Server, err: %v", err)
	}

	signalC := make(chan os.Signal, 10)
	signal.Notify(signalC, syscall.SIGINT)
	errorC := make(chan error)

	go grpcServer.Run(webAddress, errorC)
	defer grpcServer.Shutdown()

	select {
	case s := <-signalC:
		log.Info().Str("signal", s.String()).Msg("received signal, shutting down")
		return nil
	case err := <-errorC:
		return fmt.Errorf("server error: %v", err)
	}
}
