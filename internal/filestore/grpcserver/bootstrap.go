package grpcserver

import (
	"fmt"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filestore/config"
	storeserverv1 "github.com/denismitr/shardstore/pkg/storeserver/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func StartGRPCServer(
	cfg *config.Config,
	lg logger.Logger,
	fileSrv *FileServer,
) error {
	s := grpc.NewServer()
	if cfg.ReflectionAPI {
		reflection.Register(s)
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen tcp %d: %w", cfg.GRPCPort, err)
	}

	storeserverv1.RegisterFileServiceServer(s, fileSrv)

	go func() {
		if err := s.Serve(l); err != nil {
			lg.Error(fmt.Errorf("error service grpc server, err: %v", err))
		}
	}()

	gracefulShutDown(s)

	return nil
}

func gracefulShutDown(s *grpc.Server) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	<-ch
	s.GracefulStop()
}
