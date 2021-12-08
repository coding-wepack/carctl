package artutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArtifactTypeFromHost(t *testing.T) {
	const (
		h1 = "team-maven.pkg.coding.net"
		e1 = "maven"

		h2 = "team-maven.coding.net"
		e2 = ""

		h3 = "team-another-docker.pkg.test.coding.io"
		e3 = "docker"

		h4 = "auth.pkg.coding.com"
		e4 = ""
	)

	t1, err := ArtifactTypeFromHost(h1)
	assert.NoError(t, err)
	assert.Equal(t, e1, t1)

	t2, err := ArtifactTypeFromHost(h2)
	assert.ErrorIs(t, err, ErrInvalidRegistry)
	assert.Equal(t, e2, t2)

	t3, err := ArtifactTypeFromHost(h3)
	assert.NoError(t, err)
	assert.Equal(t, e3, t3)

	t4, err := ArtifactTypeFromHost(h4)
	assert.ErrorIs(t, err, ErrInvalidRegistry)
	assert.Equal(t, e4, t4)
}
