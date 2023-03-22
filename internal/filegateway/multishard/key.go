package multishard

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrInvalidFilename = errors.New("invalid file name")
)

type Key string

// ResolveKey - makes a key out of a filename
// silly implementation
func ResolveKey(fileName string) (Key, error) {
	if fileName == "" {
		return "", fmt.Errorf("cannot build a storage key without filename: %w", ErrInvalidFilename)
	}

	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, ".", "_")

	// logic with buckets would require extra work - this is a temp solution
	return Key(fmt.Sprintf("%s", fileName)), nil
}
