package logh

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewRotateFile(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	path += "/TestNewRotateFile"
	file, err := NewRotateFile(path, "test", 10, nil)
	require.NoError(t, err)
	require.NotNil(t, file)

	dir, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 1)
}

func TestRotateSize(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest-TestRotateSize")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	file, err := NewRotateFile(path, "test", 100)
	require.NoError(t, err)

	b := []byte("test\n")
	for range 100 {
		require.NoError(t, err)
		_, err = file.Write(b)
		require.NoError(t, err)
	}

	dir, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 5)
}

func TestRotateInterval(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest-TestRotateInterval")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	file, err := NewRotateFile(path, "test", 1024, WithCheckEveryN(1), WithRotateInterval(100*time.Millisecond))
	require.NoError(t, err)

	b := []byte("test\n")
	for range 5 {
		_, err = file.Write(b)
		require.NoError(t, err)
	}

	dir, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 1)

	time.Sleep(100 * time.Millisecond)
	for range 5 {
		_, err = file.Write(b)
		require.NoError(t, err)
	}

	dir, err = os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 2)
}
