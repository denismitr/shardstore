package grpcserver

import (
	"errors"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filestore/config"
	storeserverv1 "github.com/denismitr/shardstore/pkg/storeserver/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

const (
	readChunkSize = 4 * 1024
)

type storageFactory interface {
	GetWriter(appName, key string) (io.Writer, func() error, error)
	GetReader(appName, key string) (io.Reader, func() error, error)
}

type FileServer struct {
	storeserverv1.UnimplementedFileServiceServer

	cfg            *config.Config
	lg             logger.Logger
	storageFactory storageFactory
}

func NewFileServer(
	cfg *config.Config,
	lg logger.Logger,
	sf storageFactory,
) *FileServer {
	return &FileServer{cfg: cfg, lg: lg, storageFactory: sf}
}

func (fs *FileServer) Upload(stream storeserverv1.FileService_UploadServer) error {
	var writer io.Writer
	var wCloser func() error
	defer func() {
		if writer != nil {
			if errClose := wCloser(); errClose != nil {
				fs.lg.Error(errClose)
			}
		}
	}()

	ctx := stream.Context() // todo: key from headers
	fs.lg.Debugf("%s received upload request", fs.cfg.AppName)
	for {
		if ctx.Err() != nil {
			return status.Errorf(codes.Internal, ctx.Err().Error())
		}

		req, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return stream.SendAndClose(&storeserverv1.UploadResponse{})
			}

			fs.lg.Error(err)
			return status.Error(codes.Internal, err.Error())
		}

		if writer == nil {
			var errStore error
			// todo:  key should come from request(incoming context) header
			// todo: in that case writer can be instantiated in the beginning of the function
			writer, wCloser, errStore = fs.storageFactory.GetWriter(fs.cfg.AppName, req.Key)
			if errStore != nil {
				fs.lg.Error(errStore)
				return status.Error(codes.Internal, errStore.Error())
			}
			fs.lg.Debugf("writer created in %s for key %s", fs.cfg.AppName, req.Key)
		}

		if _, err := writer.Write(req.GetPayload()); err != nil {
			fs.lg.Error(err)
			return status.Error(codes.Internal, err.Error())
		}
	}
}

func (fs *FileServer) Download(
	req *storeserverv1.DownloadRequest,
	stream storeserverv1.FileService_DownloadServer,
) error {
	ctx := stream.Context()
	rc, closer, err := fs.storageFactory.GetReader(fs.cfg.AppName, req.Key)
	if err != nil {
		return status.Errorf(codes.Internal, "app %s failed to obtain reader for key: %s", fs.cfg.AppName, req.Key, err)
	}

	defer func() {
		if err := closer(); err != nil {
			fs.lg.Error(err)
		}
	}()

	chunk := &storeserverv1.DownloadResponse{Payload: make([]byte, readChunkSize)}
	for {
		if ctx.Err() != nil {
			return status.Errorf(codes.Internal, ctx.Err().Error())
		}

		n, err := rc.Read(chunk.Payload)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			} else {
				return status.Errorf(codes.Internal, "file read failed: %s", err.Error())
			}
		}

		chunk.Payload = chunk.Payload[:n]
		errStream := stream.Send(chunk)
		if errStream != nil {
			return status.Errorf(codes.Internal, "stream send failed: %s", errStream.Error())
		}
	}

	return nil
}
