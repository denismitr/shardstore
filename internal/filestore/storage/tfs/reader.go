package tfs

import (
	"fmt"
	"os"
	"path"
)

type tmpFileReader struct {
	key  string
	dir  string
	file *os.File
}

func newTmpFileReader(serverName, key string) (*tmpFileReader, error) {
	dir := fmt.Sprintf("tmp/%s/filestore", serverName)
	filePath := path.Join(dir, key)

	f, err := os.OpenFile(filePath, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &tmpFileReader{
		key:  key,
		dir:  dir,
		file: f,
	}, nil
}

func (t *tmpFileReader) Read(p []byte) (n int, err error) {
	return t.file.Read(p)
}

func (t *tmpFileReader) Close() error {
	return t.file.Close()
}
