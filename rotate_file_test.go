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

func TestRotateInterval(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest-TestRotateInterval")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	file, err := NewRotateFile(path, "test", 1024, WithCheckEveryN(2), WithRotateInterval(1*time.Second))
	require.NoError(t, err)

	_, err = file.Write([]byte("1"))
	require.NoError(t, err)
	_, err = file.Write([]byte("2"))
	require.NoError(t, err)

	dir, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 1)

	time.Sleep(1 * time.Second)
	_, err = file.Write([]byte("1"))
	require.NoError(t, err)
	_, err = file.Write([]byte("2"))
	require.NoError(t, err)

	dir, err = os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 2)
}
