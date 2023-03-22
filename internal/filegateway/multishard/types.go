package multishard

import (
	"errors"
	"fmt"
	"strings"
)

type (
	Key        string
	ChunkIDX   int
	ServerIDX  int
	MultiShard map[ChunkIDX]ServerIDX
)

var (
	ErrInvalidFilename = errors.New("invalid file name")
)

func ResolveKey(fileName string) (Key, error) {
	if fileName == "" {
		return "", fmt.Errorf("cannot build a storage key without filename: %w", ErrInvalidFilename)
	}

	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")

	// logic with buckets would require extra work - this is a temp solution
	return Key(fmt.Sprintf("%s", fileName)), nil
}
