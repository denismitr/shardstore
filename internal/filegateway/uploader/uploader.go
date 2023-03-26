package uploader

import (
	"context"
	"errors"
	"fmt"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/metastore"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	"io"
	"mime/multipart"
	"sync"
	"time"
)

const (
	maxBufSize = 4 * 1024
)

type shardManager interface {
	// ResolveShardMap - resolves a shard map for a given key
	// todo: in case we need to distribute data according to the current servers capacity
	// todo: we probably should pass the metadata summary with statistics on server data distribution
	ResolveShardMap(key multishard.Key) (multishard.ShardMap, error)
}

type remoteStorage interface {
	Put(
		ctx context.Context,
		key multishard.Key,
		serverID multishard.ServerIdx,
		r io.Reader,
	) error
}

// metaStorage is a gateway to a database (e.g. MongoDB or Cassandra) that stores metadata on files
// and statistics on servers
type metaStorage interface {
	Store(ctx context.Context, key multishard.Key, entry *metastore.ShardPlan) error
}

type Uploader struct {
	cfg          *config.Config
	lg           logger.Logger
	shardManager shardManager
	remoteStore  remoteStorage
	metaStore    metaStorage
}

func NewUploader(
	cfg *config.Config,
	shardManager shardManager,
	remoteStore remoteStorage,
	metaStore metaStorage,
	lg logger.Logger,
) *Uploader {
	return &Uploader{
		cfg:          cfg,
		lg:           lg,
		shardManager: shardManager,
		remoteStore:  remoteStore,
		metaStore:    metaStore,
	}
}

func (u *Uploader) Upload(
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

	// build the shard information with chunks and corresponding servers
	planBuilder := metastore.NewShardPlanBuilder(key, int(h.Size), len(u.cfg.StorageServers))

	multiShard, err := u.shardManager.ResolveShardMap(key)
	if err != nil {
		return err
	}
	chunkSize := h.Size / u.cfg.NumberOfChunks
	remainder := (h.Size % u.cfg.NumberOfChunks) != 0

	var wg sync.WaitGroup
	wg.Add(int(u.cfg.NumberOfChunks))
	for i := int64(0); i < u.cfg.NumberOfChunks; i++ {
		hasResidualBytes := i == u.cfg.NumberOfChunks-1 && remainder

		go func(chunkIdx int64, hasResidualBytes bool) {
			defer wg.Done()

			serverIdx := multiShard[multishard.ChunkIdx(chunkIdx)]
			offset := chunkSize * chunkIdx
			size := chunkSize
			if hasResidualBytes {
				size += h.Size - (u.cfg.NumberOfChunks * chunkSize)
			}
			if err := u.uploadChunk(
				ctx,
				key,
				f,
				serverIdx,
				int(size),
				offset,
			); err != nil {
				errCh <- err
				return
			}

			// add shard info
			if err := planBuilder.AddShard(int(chunkIdx), int(serverIdx), int(size)); err != nil {
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
		return fmt.Errorf("upload failed: %w", err)
	case <-ctx.Done():
		return ctx.Err()
	case <-doneCh:
		// save metadata about the key and associated shards
		if err := u.metaStore.Store(ctx, key, planBuilder.Build()); err != nil {
			return fmt.Errorf("upload could not be accomplished: %w", err)
		}
		return nil
	}
}

func (u *Uploader) uploadChunk(
	parentCtx context.Context,
	key multishard.Key,
	f multipart.File,
	serverID multishard.ServerIdx,
	size int,
	offset int64,
) error {
	if size == 0 {
		return fmt.Errorf("how can size be 0")
	}

	var wg sync.WaitGroup
	readyCh := make(chan struct{})
	errCh := make(chan error, 1)
	r, w := io.Pipe()

	// cancel will be called on function exit, thus remoteStorage.Put will receive done signal
	// in case write was not finished
	ctx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := u.remoteStore.Put(ctx, key, serverID, r); err != nil {
			u.lg.Error(err)
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer func() {
			if err := w.Close(); err != nil {
				u.lg.Error(fmt.Errorf("failed closing the writer: %w", err))
			}
			wg.Done()
		}()

		if err := u.send(f, w, size, offset); err != nil {
			u.lg.Error(err)
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

func (u *Uploader) send(f multipart.File, w io.Writer, size int, offset int64) error {
	bufSize := minBufSize(size, maxBufSize)
	readBuf := make([]byte, bufSize)

	totalBytesWritten := 0
	for {
		if totalBytesWritten >= size {
			return nil
		}

		n, err := f.ReadAt(readBuf, offset)
		if err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("failed reading from offset %d: %w", offset, err)
		}

		if size < totalBytesWritten+n {
			u.lg.Debugf("writing last bytes")
			n = size - totalBytesWritten
		}

		if _, err := w.Write(readBuf[:n]); err != nil {
			return fmt.Errorf("cannot write payload: %w", err)
		}

		totalBytesWritten += n
		offset += int64(n)
	}
}

func minBufSize(chunkSize, maxBufSize int) int {
	if chunkSize < maxBufSize {
		return chunkSize
	}
	return maxBufSize
}
