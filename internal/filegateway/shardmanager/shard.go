package shardmanager

import (
	"errors"
	"fmt"
	hash "github.com/cespare/xxhash/v2"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
)

type ShardManager struct {
	cfg     *config.Config
	lg      logger.Logger
	servers int
}

func NewShardManager(cfg *config.Config, lg logger.Logger) (*ShardManager, error) {
	if len(cfg.StorageServers) < int(cfg.NumberOfChunks) {
		return nil, fmt.Errorf("less servers than number of chunks: %w", ErrInvalidNumberOfServers)
	}

	return &ShardManager{
		cfg:     cfg,
		lg:      lg,
		servers: len(cfg.StorageServers),
	}, nil
}

var ErrInvalidNumberOfServers = errors.New("invalid number of servers")

func (sm *ShardManager) GetMultiShard(key multishard.Key) (multishard.MultiShard, error) {
	chunks := int(sm.cfg.NumberOfChunks)
	if sm.servers < chunks {
		return nil, fmt.Errorf("less servers than chunks: %w", ErrInvalidNumberOfServers) // todo: wrap
	}

	serverIdx := hash.Sum64String(string(key)) % uint64(sm.servers)
	ms := make(multishard.MultiShard, chunks)

	for chunkIdx := 0; chunkIdx < chunks; chunkIdx++ {
		ms[multishard.ChunkIDX(chunkIdx)] = multishard.ServerIDX(serverIdx)
		if serverIdx < uint64(sm.servers)-1 {
			serverIdx++
		} else {
			serverIdx = 0
		}
	}

	return ms, nil
}
