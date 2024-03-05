package logh

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRotateFile(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest")
	require.NoError(t, err)

	path += "/TestNewRotateFile"
	file, err := NewRotateFile(path, "test", 10)
	require.NoError(t, err)
	require.NotNil(t, file)

	dir, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 1)

	err = os.RemoveAll(path)
	require.NoError(t, err)
}
