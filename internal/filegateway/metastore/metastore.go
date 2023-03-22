package metastore

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/denismitr/shardstore/internal/common/logger"
	"github.com/denismitr/shardstore/internal/filegateway/multishard"
	"os"
	"sync"
)

type TmpMetaStore struct {
	lg  logger.Logger
	mx  sync.Mutex
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

type StoreRequest struct {
	Key string `json:"key"`

	// Shards represent a shard for every chunk
	Shards []Shard `json:"shards"`
}

func NewStoreRequest(key multishard.Key, chunks int) *StoreRequest {
	return &StoreRequest{
		Key:    string(key),
		Shards: make([]Shard, chunks),
	}
}

func (req *StoreRequest) SetShard(chunkIdx int, shard Shard) error {
	if chunkIdx > len(req.Shards)-1 || chunkIdx < 0 {
		return fmt.Errorf("invalid chunk idx %d", chunkIdx)
	}

	req.Shards[chunkIdx] = shard
	return nil
}

func (s *TmpMetaStore) Store(ctx context.Context, req *StoreRequest) error {
	b, err := json.Marshal(req)
	if err != nil {
		return err //todo: wrap
	}
	s.mx.Lock()
	filePath := fmt.Sprintf("%s/%s", s.dir, req.Key)
	if err := os.WriteFile(filePath, b, 0644); err != nil {
		return err //todo: wrap
	}
	return nil
}
