package artutil

import (
	"strings"

	"github.com/pkg/errors"
)

var (
	ErrInvalidRegistry = errors.New("invalid registry")
)

func ArtifactTypeFromHost(host string) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" {
		return "", errors.New("illegal argument: host should not be empty")
	}

	chunks := strings.Split(host, ".")
	if len(chunks) < 4 {
		return "", ErrInvalidRegistry
	}

	first := chunks[0]
	firstChunks := strings.Split(first, "-")
	firstChunkSize := len(firstChunks)
	if firstChunkSize < 2 {
		return "", ErrInvalidRegistry
	}

	return firstChunks[firstChunkSize-1], nil
}
