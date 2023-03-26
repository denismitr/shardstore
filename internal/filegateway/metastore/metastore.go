package metastore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	"io"
	"os"
	"sync"
)

// TmpMetaStore - silly implementation only for local testing
// meta store should be a normal database
// should also support file servers statistics
type TmpMetaStore struct {
	lg  logger.Logger
	mx  sync.Mutex // todo: lock should be key specific
	dir string
}

func NewTmpMetaStore(appName string, lg logger.Logger) (*TmpMetaStore, error) {
	dir := fmt.Sprintf("tmp/%s/metastore", appName)
	if err := os.MkdirAll(dir, 0644); err != nil {
		return nil, err
	}
	return &TmpMetaStore{lg: lg, dir: dir}, nil
}

type Shard struct {
	ChunkIdx  int    `json:"chunk_idx"`
	ServerIdx int    `json:"server_idx"`
	Size      int    `json:"size"`
	Checksum  uint32 `json:"checksum"`
}

type ShardPlan struct {
	OriginalSize int `json:"original_size"`

	// Shards represent a shard for every chunk
	Shards []Shard `json:"shards"`
}

type ShardPlanBuilder struct {
	key          string
	clusterEntry *ShardPlan
	mx           sync.Mutex
}

func NewShardPlanBuilder(key multishard.Key, originalSize, chunks int) *ShardPlanBuilder {
	return &ShardPlanBuilder{
		key: string(key),
		clusterEntry: &ShardPlan{
			OriginalSize: originalSize,
			Shards:       make([]Shard, chunks),
		},
	}
}

// AddShard - adds a new shard to a cluster map
func (b *ShardPlanBuilder) AddShard(chunkIdx, serverIdx, size int) error {
	b.mx.Lock()
	defer b.mx.Unlock()
	if chunkIdx > len(b.clusterEntry.Shards)-1 || chunkIdx < 0 {
		return fmt.Errorf("invalid chunk idx %d for key %s", chunkIdx, b.key)
	}

	b.clusterEntry.Shards[chunkIdx] = Shard{
		ChunkIdx:  chunkIdx,
		ServerIdx: serverIdx,
		Size:      size,
	}

	return nil
}

func (b *ShardPlanBuilder) Build() *ShardPlan {
	return b.clusterEntry
}

func (s *TmpMetaStore) Store(ctx context.Context, key multishard.Key, plan *ShardPlan) error {
	b, err := json.Marshal(plan)
	if err != nil {
		return err //todo: wrap
	}
	s.mx.Lock()
	defer s.mx.Unlock()
	filePath := fmt.Sprintf("%s/%s", s.dir, key)
	if err := os.WriteFile(filePath, b, 0644); err != nil {
		return err //todo: wrap
	}
	return nil
}

func (s *TmpMetaStore) GetShardPlan(ctx context.Context, key multishard.Key) (*ShardPlan, error) {
	s.mx.Lock()
	defer s.mx.Unlock()

	filePath := fmt.Sprintf("%s/%s", s.dir, key)
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err // todo: wrap
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, err // todo: wrap
	}

	var plan ShardPlan
	if err := json.Unmarshal(b, &plan); err != nil {
		return nil, err // todo: wrap
	}

	if len(plan.Shards) == 0 {
		return nil, fmt.Errorf("should have retrieved some shards")
	}

	return &plan, nil
}
