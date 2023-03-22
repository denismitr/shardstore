package storage

import (
	"bytes"
	"context"
)

// BufferStorage just a fake buffer storage for testing
type BufferStorage struct {
	key    string
	buffer *bytes.Buffer
}

func NewBufferStorage(key string) *BufferStorage {
	return &BufferStorage{
		key:    key,
		buffer: &bytes.Buffer{},
	}
}

func (f *BufferStorage) Write(ctx context.Context, chunk []byte) error {
	_, err := f.buffer.Write(chunk)

	return err
}

func (f *BufferStorage) Flush(ctx context.Context) error {
	return nil
}
