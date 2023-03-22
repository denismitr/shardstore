package processor

import (
	"bytes"
	"context"
	"errors"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/metastore"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	"io"
	"mime/multipart"
	"sync"
)

const (
	maxBufSize = 4 * 1024
)

type shardManager interface {
	GetMultiShard(key multishard.Key) (multishard.MultiShard, error)
}

type remoteStorage interface {
	Put(ctx context.Context, key multishard.Key, serverID multishard.ServerIDX, size int, r io.Reader) error
}

type metaStorage interface {
	Store(ctx context.Context, req *metastore.StoreRequest) error
}

type Processor struct {
	cfg          *config.Config
	shardManager shardManager
	remoteStore  remoteStorage
	metaStore    metaStorage
}

func NewProcessor(
	cfg *config.Config,
	shardManager shardManager,
	remoteStore remoteStorage,
	metaStore metaStorage,
) *Processor {
	return &Processor{
		cfg:          cfg,
		shardManager: shardManager,
		remoteStore:  remoteStore,
		metaStore:    metaStore,
	}
}

func (p *Processor) Process(
	ctx context.Context,
	f multipart.File,
	h *multipart.FileHeader,
) error {
	errCh := make(chan error, 1)
	doneCh := make(chan struct{}, 1)

	key, err := multishard.ResolveKey(h.Filename)
	if err != nil {
		return err
	}

	storeMetaReq := metastore.NewStoreRequest(key, len(p.cfg.StorageServers))
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
				errCh <- err // todo: wrap
				return
			}

			// add shard info
			if err := storeMetaReq.SetShard(int(chunkIdx), metastore.Shard{
				ChunkIdx:  int(chunkIdx),
				ServerIdx: int(serverID),
				Size:      int(size),
			}); err != nil {
				errCh <- err // todo: wrap
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
		// save metadata about the key and associated shards
		if err := p.metaStore.Store(ctx, storeMetaReq); err != nil {
			return err // todo: wrap
		}
		return nil
	}
}

func (p *Processor) processChunk(
	parentCtx context.Context,
	key multishard.Key,
	f multipart.File,
	serverID multishard.ServerIDX,
	size int,
	offset int64,
) error {
	var wg sync.WaitGroup
	readyCh := make(chan struct{})
	errCh := make(chan error, 1)

	sendBuf := &bytes.Buffer{}

	// cancel will be called on function exit, thus remoteStorage.Put will receive done signal
	// in case write was not finished
	ctx, cancel := context.WithCancel(parentCtx)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := p.remoteStore.Put(ctx, key, serverID, size, sendBuf); err != nil {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := p.send(f, sendBuf, size, offset); err != nil {
			errCh <- err
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

func (p *Processor) send(f multipart.File, w io.Writer, size int, offset int64) error {
	maxReadLen := minBufSize(size, maxBufSize)
	readBuf := make([]byte, maxReadLen)

	written := 0
	bytesRead := 0
	for {
		n, err := f.ReadAt(readBuf, offset)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			} else {
				return err
			}
		}

		bytesRead += n
		if bytesRead < size {
			if _, err := w.Write(readBuf[:n]); err != nil {
				return err
			}
			written += n
		} else {
			if _, err := w.Write(readBuf[:size-written]); err != nil {
				return err
			}
			written = size
		}

		if written >= size {
			return nil
		}
	}
}

func minBufSize(chunkSize, maxBufSize int) int {
	if chunkSize < maxBufSize {
		return chunkSize
	}
	return maxBufSize
}
