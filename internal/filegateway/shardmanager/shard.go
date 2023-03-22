package shardmanager

import (
	"errors"
	"fmt"
	hash "github.com/cespare/xxhash/v2"
	"github.com/denismitr/shardstore/internal/filegateway/config"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
)

type ShardManager struct {
	cfg     *config.Config
	servers int
}

func NewShardManager(cfg *config.Config) *ShardManager {
	return &ShardManager{
		cfg:     cfg,
		servers: len(cfg.StorageServers),
	}
}

var ErrInvalidNumberOfServers = errors.New("invalid number of servers")

func (sm *ShardManager) GetMultiShard(key string) (multishard.MultiShard, error) {
	chunks := int(sm.cfg.NumberOfChunks)
	if sm.servers < chunks {
		return nil, fmt.Errorf("less servers than chunks: %w", ErrInvalidNumberOfServers) // todo: wrap
	}

	serverIdx := hash.Sum64String(key) % uint64(sm.servers)
	ms := make(multishard.MultiShard, chunks)

	for chunkIdx := 0; chunkIdx < chunkIdx; chunkIdx++ {
		ms[multishard.ChunkIDX(chunkIdx)] = multishard.ServerIDX(serverIdx)
		if serverIdx < uint64(sm.servers)-1 {
			serverIdx++
		} else {
			serverIdx = 0
		}
	}

	return ms, nil
}
