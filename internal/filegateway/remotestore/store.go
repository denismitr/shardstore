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
	client map[multishard.ServerIDX]storeserverv1.UploadServiceClient
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
	ErrInvalidSize     = errors.New("invalid size")
)

const bufSize = 4 * 1024

func (s *GRPCStore) Put(
	ctx context.Context,
	key multishard.Key,
	serverID multishard.ServerIDX,
	size int,
	r io.Reader,
) error {
	if size <= 0 {
		return ErrInvalidSize
	}

	s.mx.RLock()
	client, ok := s.client[serverID]
	if !ok {
		s.mx.RUnlock()
		return ErrServerIDInvalid // todo: wrap
	}
	s.mx.RUnlock()

	upload, err := client.Upload(ctx)
	if err != nil {
		return fmt.Errorf("failed to obtain upload client: %w", err)
	}

	buf := make([]byte, bufSize)
	bytesSend := 0
	for {
		n, err := r.Read(buf)
		if err != nil {
			s.lg.Error(err)
			break
		}

		if err := upload.Send(&storeserverv1.UploadRequest{
			Key:     string(key),
			Payload: buf[:n],
		}); err != nil {
			if errCloseSend := upload.CloseSend(); errCloseSend != nil {
				s.lg.Error(err)
			}
			return fmt.Errorf("failed to send a chunk of data for key %s: %w", key, err) // todo: wrap
		}

		bytesSend += n
		if bytesSend >= size {
			break
		}
	}

	if _, err := upload.CloseAndRecv(); err != nil {
		return err
	}

	return nil
}
