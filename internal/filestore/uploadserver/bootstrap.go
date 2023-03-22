package uploadserver

import (
	"fmt"
	"github.com/denismitr/shardstore/internal/filestore/config"
	storeserverv1 "github.com/denismitr/shardstore/pkg/storeserver/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func StartGRPCServer(cfg *config.Config, uploadSrv *Server) error {
	s := grpc.NewServer()
	if cfg.ReflectionAPI {
		reflection.Register(s)
	}

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("failed to listen tcp %d: %w", cfg.GRPCPort, err)
	}

	storeserverv1.RegisterUploadServiceServer(s, uploadSrv)

	go func() {
		if err := s.Serve(l); err != nil {
			log.Fatalf("error service grpc server, err: %v", err)
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
