package remotestore

import (
	"context"
	"errors"
	"fmt"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	storeserverv1 "github.com/denismitr/shardstore/pkg/storeserver/v1"
	"io"
	"sync"
)

type GRPCStore struct {
	cfg    *config.Config
	client map[multishard.ServerIdx]storeserverv1.FileServiceClient
	mx     sync.RWMutex
	lg     logger.Logger
}

func NewGRPCStore(
	cfg *config.Config,
	lg logger.Logger,
) (*GRPCStore, error) {
	clients, err := bootstrapClients(cfg)
	if err != nil {
		return nil, err
	}
	return &GRPCStore{cfg: cfg, client: clients, lg: lg}, nil
}

var (
	ErrServerIDInvalid = errors.New("server id is invalid")
)

const bufSize = 4 * 1024

func (s *GRPCStore) Put(
	ctx context.Context,
	key multishard.Key,
	serverIdx multishard.ServerIdx,
	r io.Reader,
) error {
	s.mx.RLock()
	client, ok := s.client[serverIdx]
	if !ok {
		s.mx.RUnlock()
		return ErrServerIDInvalid // todo: wrap
	}
	s.mx.RUnlock()

	// todo: key into the outgoing context
	upload, err := client.Upload(ctx)
	if err != nil {
		return fmt.Errorf("failed to obtain upload client: %w", err)
	}

	if err := s.doUpload(ctx, key, r, upload); err != nil {
		s.lg.Error(err)
		if errCloseSend := upload.CloseSend(); errCloseSend != nil {
			s.lg.Error(errCloseSend)
		}
		return err
	}

	if _, err := upload.CloseAndRecv(); err != nil {
		s.lg.Error(fmt.Errorf("failed to close and recv the upload: %w", err))
	}

	return nil
}

func (s *GRPCStore) doUpload(
	ctx context.Context,
	key multishard.Key,
	r io.Reader,
	upload storeserverv1.FileService_UploadClient,
) error {
	buf := make([]byte, bufSize)
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		n, err := r.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed to read payload: %w", err)
		}

		if n == 0 {
			return nil
		}

		if err := upload.Send(&storeserverv1.UploadRequest{
			Key:     string(key),
			Payload: buf[:n],
		}); err != nil {
			s.lg.Error(err)
			return fmt.Errorf("failed to send a chunk of data for key %s: %w", key, err)
		}
	}
}

func (s *GRPCStore) Get(
	ctx context.Context,
	key multishard.Key,
	serverID multishard.ServerIdx,
	w io.Writer,
) (int, error) {
	s.mx.RLock()
	client, ok := s.client[serverID]
	if !ok {
		s.mx.RUnlock()
		return 0, ErrServerIDInvalid // todo: wrap
	}
	s.mx.RUnlock()

	stream, err := client.Download(ctx, &storeserverv1.DownloadRequest{
		Key: string(key),
	})
	if err != nil {
		return 0, fmt.Errorf("could not get the download stream for server %d: %w", serverID, err) // todo: wrap
	}

	downloaded := 0
	for {
		if ctx.Err() != nil {
			return downloaded, ctx.Err()
		}

		req, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return downloaded, nil
			}

			s.lg.Error(err)
			return downloaded, fmt.Errorf("download stream failed: %w", err)
		}

		if n, err := w.Write(req.Payload); err != nil {
			return downloaded, fmt.Errorf("could not write downloaded content: %w", err)
		} else {
			downloaded += n
		}
	}
}
