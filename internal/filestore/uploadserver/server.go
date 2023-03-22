package uploadserver

import (
	"context"
	"errors"
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
	// todo: storage factory
}

func NewServer(cfg *config.Config) *Server {
	return &Server{cfg: cfg}
}

type destinationStorage interface {
	Write(ctx context.Context, chunk []byte) error
	Flush(ctx context.Context) error
}

func (s *Server) Upload(stream storeserverv1.UploadService_UploadServer) error {
	var store destinationStorage
	for {
		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) && store != nil {
				if err := store.Flush(stream.Context()); err != nil {
					return status.Error(codes.Internal, err.Error())
				}
				return stream.SendAndClose(&storeserverv1.UploadResponse{
					Checksum: 0, // todo
				})
			}
			return status.Error(codes.Internal, err.Error())
		}

		if store == nil {
			// todo: factory that creates a storage instance from config settings
			// todo: maybe key should come from request header
			store = storage.NewBufferStorage(req.Key)
		}

		if err := store.Write(stream.Context(), req.GetPayload()); err != nil {
			return status.Error(codes.Internal, err.Error())
		}
	}
}
