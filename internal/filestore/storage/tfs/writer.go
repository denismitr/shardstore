package tfs

import (
	"fmt"
	"os"
	"path"
)

type tmpFileWriter struct {
	key  string
	dir  string
	file *os.File
}

func newTmpFileWriter(serverName, key string) (*tmpFileWriter, error) {
	dir := fmt.Sprintf("tmp/%s/filestore", serverName)
	if err := os.MkdirAll(dir, 0644); err != nil {
		return nil, err
	}

	filePath := path.Join(dir, key)

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	return &tmpFileWriter{
		key:  key,
		dir:  dir,
		file: f,
	}, nil
}

func (fs *tmpFileWriter) Write(chunk []byte) (int, error) {
	n, err := fs.file.Write(chunk)
	if err != nil {
		return 0, fmt.Errorf("failed to write to %s: %w", fs.file.Name(), err)
	}
	if err := fs.file.Sync(); err != nil {
		return n, fmt.Errorf("could not sync file %s: %w", fs.file.Name(), err)
	}
	return n, nil
}

func (fs *tmpFileWriter) Sync() error {
	if err := fs.file.Sync(); err != nil {
		return fmt.Errorf("could not sync file %s: %w", fs.file.Name(), err)
	}
	return nil
}

func (fs *tmpFileWriter) Close() error {
	if err := fs.file.Close(); err != nil {
		return fmt.Errorf("could not close file %s: %w", fs.file.Name(), err)
	}
	return nil
}
