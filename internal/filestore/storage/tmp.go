package storage

import (
	"context"
	"fmt"
	"os"
	"path"
)

type TmpFileStorage struct {
	key  string
	dir  string
	file *os.File
}

func NewTmpFileStorage(serverName, key string) (*TmpFileStorage, error) {
	dir := fmt.Sprintf("tmp/%s/filestore", serverName)
	if err := os.MkdirAll(dir, 0644); err != nil {
		return nil, err
	}

	filePath := path.Join(dir, key)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return &TmpFileStorage{
		key:  key,
		dir:  dir,
		file: f,
	}, nil
}

func (fs *TmpFileStorage) Write(ctx context.Context, chunk []byte) error {
	_, err := fs.file.Write(chunk)
	if err != nil {
		return fmt.Errorf("failed to write to %s: %w", fs.file.Name(), err)
	}
	return nil
}

func (fs *TmpFileStorage) FSync() error {
	return fs.file.Sync()
}

func (fs *TmpFileStorage) Close() error {
	return fs.file.Close()
}
