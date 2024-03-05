package logh

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRotateFile(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	path += "/TestNewRotateFile"
	file, err := NewRotateFile(path, "test", 10)
	require.NoError(t, err)
	require.NotNil(t, file)

	dir, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 1)
}

func TestRotateSize(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest-TestWrite")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	file, err := NewRotateFile(path, "test", 10)
	require.NoError(t, err)

	_, err = file.Write([]byte("12345678901234567890"))
	require.NoError(t, err)
	_, err = file.Write([]byte("12345678901234567890"))
	require.NoError(t, err)

	dir, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 3)
}
