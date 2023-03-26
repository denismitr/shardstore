package downloader

import (
	"context"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/metastore"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	"io"
)

type metaStorage interface {
	GetShardPlan(ctx context.Context, key multishard.Key) (*metastore.ShardPlan, error)
}

type remoteStorage interface {
	Get(
		ctx context.Context,
		key multishard.Key,
		serverID multishard.ServerIdx,
		w io.Writer,
	) (int, error)
}

type Downloader struct {
	cfg         *config.Config
	lg          logger.Logger
	remoteStore remoteStorage
	metaStore   metaStorage
}

func NewDownloader(
	cfg *config.Config,
	remoteStore remoteStorage,
	metaStore metaStorage,
	lg logger.Logger,
) *Downloader {
	return &Downloader{cfg: cfg, remoteStore: remoteStore, metaStore: metaStore, lg: lg}
}

func (d *Downloader) Download(
	ctx context.Context,
	fileName string,
	w io.Writer,
) (int, error) {
	key, err := multishard.ResolveKey(fileName)
	if err != nil {
		return 0, err // todo: wrap
	}

	plan, err := d.metaStore.GetShardPlan(ctx, key)
	if err != nil {
		return 0, err
	}

	totalDownloaded := 0
	for _, shard := range plan.Shards {
		d.lg.Debugf("getting shard for chunk %d from server %d", shard.ChunkIdx, shard.ServerIdx)
		n, err := d.remoteStore.Get(ctx, key, multishard.ServerIdx(shard.ServerIdx), w)
		if err != nil {
			return totalDownloaded, err // todo: wrap
		}
		totalDownloaded += n
	}

	return totalDownloaded, nil
}
