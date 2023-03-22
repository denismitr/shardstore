package uploadserver

import (
	"context"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filestore/config"
	"github.com/denismitr/shardstore/internal/filestore/storage"
	storeserverv1 "github.com/denismitr/shardstore/pkg/storeserver/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type Server struct {
	cfg *config.Config
	storeserverv1.UnimplementedUploadServiceServer
	lg logger.Logger
	// todo: storage factory
}

func NewServer(cfg *config.Config, lg logger.Logger) *Server {
	return &Server{cfg: cfg, lg: lg}
}

type destinationStorage interface {
	Write(ctx context.Context, chunk []byte) error
	FSync() error
	Close() error
}

func (s *Server) Upload(stream storeserverv1.UploadService_UploadServer) error {
	var store destinationStorage
	defer func() {
		if store != nil {
			if errClose := store.Close(); errClose != nil {
				s.lg.Error(errClose)
			}
		}
	}()

	for {
		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF && store != nil {
				if errSync := store.FSync(); errSync != nil {
					s.lg.Error(errSync)
					return status.Error(codes.Internal, errSync.Error())
				}

				return stream.SendAndClose(&storeserverv1.UploadResponse{
					Checksum: 0, // todo
				})
			}

			s.lg.Error(err)
			return status.Error(codes.Internal, err.Error())
		}

		if store == nil {
			var errStore error
			// todo: factory that creates a storage instance from config settings
			// todo: maybe key should come from request header
			store, errStore = storage.NewTmpFileStorage(s.cfg.AppName, req.Key)
			if errStore != nil {
				s.lg.Error(errStore)
				return status.Error(codes.Internal, errStore.Error())
			}
		}

		if err := store.Write(stream.Context(), req.GetPayload()); err != nil {
			s.lg.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
	}
}
