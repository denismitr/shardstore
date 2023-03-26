package tfs

import "io"

// KeyDir - a storage implementation based on a local filesystem
type KeyDir struct {
	// todo: mutex by key
}

func NewKeyDir() *KeyDir {
	return &KeyDir{}
}

func (kd *KeyDir) GetReader(appName, key string) (io.Reader, func() error, error) {
	// todo: lock the key for read
	f, err := newTmpFileReader(appName, key)
	if err != nil {
		return nil, nil, err
	}
	closer := func() error {
		// todo: release lock
		return f.Close()
	}
	return f, closer, nil
}

func (kd *KeyDir) GetWriter(appName, key string) (io.Writer, func() error, error) {
	// todo: lock the key for read
	f, err := newTmpFileWriter(appName, key)
	if err != nil {
		return nil, nil, err
	}
	closer := func() error {
		// todo: release lock
		return f.Close()
	}
	return f, closer, nil
}
