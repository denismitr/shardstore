package processor

import (
	"bytes"
	"context"
	"errors"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	"io"
	"mime/multipart"
	"sync"
)

const (
	maxBufSize = 4 * 1024
)

type shardManager interface {
	GetMultiShard(key string) (multishard.MultiShard, error)
}

type store interface {
	Put(ctx context.Context, key string, serverID multishard.ServerIDX, size int, r io.Reader) error
}

type Processor struct {
	cfg          *config.Config
	shardManager shardManager
	store        store
}

func NewProcessor(cfg *config.Config, shardManager shardManager, store store) *Processor {
	return &Processor{cfg: cfg, shardManager: shardManager, store: store}
}

func (p *Processor) Process(
	ctx context.Context,
	f multipart.File,
	h *multipart.FileHeader,
) error {
	errCh := make(chan error, 1)
	doneCh := make(chan struct{}, 1)

	key := h.Filename
	multiShard, err := p.shardManager.GetMultiShard(key)
	if err != nil {
		return err
	}
	chunkSize := h.Size / p.cfg.NumberOfChunks
	remainder := (h.Size % p.cfg.NumberOfChunks) != 0

	var wg sync.WaitGroup
	wg.Add(int(p.cfg.NumberOfChunks))
	for i := int64(0); i < p.cfg.NumberOfChunks; i++ {
		hasResidualBytes := i == p.cfg.NumberOfChunks-1 && remainder
		go func(chunkIdx int64, hasResidualBytes bool) {
			defer wg.Done()

			serverID := multiShard[multishard.ChunkIDX(chunkIdx)]
			offset := chunkSize * chunkIdx
			size := chunkSize
			if hasResidualBytes {
				size += h.Size - (p.cfg.NumberOfChunks * chunkSize)
			}
			if err := p.processChunk(
				ctx,
				key,
				f,
				serverID,
				int(size),
				offset,
			); err != nil {
				errCh <- err
				return
			}
		}(i, hasResidualBytes)
	}

	go func() {
		wg.Wait()
		close(doneCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-doneCh:
		return nil
	}
}

func (p *Processor) processChunk(
	parentCtx context.Context,
	key string,
	f multipart.File,
	serverID multishard.ServerIDX,
	size int,
	offset int64,
) error {
	maxReadLen := minBufSize(size, maxBufSize)
	readBuf := make([]byte, maxReadLen)
	sendBuf := &bytes.Buffer{}

	var wg sync.WaitGroup
	readyCh := make(chan struct{})
	errCh := make(chan error, 1)

	// cancel will be called on function exit, thus store.Put will receive done signal
	// in case write was not finished
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := p.store.Put(ctx, key, serverID, size, sendBuf); err != nil {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		written := 0
		bytesRead := 0
		for {
			n, err := f.ReadAt(readBuf, offset)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				} else {
					errCh <- err
				}
			}

			bytesRead += n
			if bytesRead < size {
				sendBuf.Write(readBuf[:n])
				written += n
			} else {
				sendBuf.Write(readBuf[:size-written])
				written = size
			}

			if written >= size {
				return
			}
		}
	}()

	go func() {
		wg.Wait()
		close(readyCh)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-readyCh:
		return nil
	case err := <-errCh:
		return err
	}
}

func minBufSize(chunkSize, maxBufSize int) int {
	if chunkSize < maxBufSize {
		return chunkSize
	}
	return maxBufSize
}
