package logh

import (
	"os"
	"path/filepath"
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

func TestPruneByMaxAge(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest-TestPruneByMaxAge")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	file, err := NewRotateFile(path, "test", 1024,
		WithCheckEveryN(1),
		WithRotateInterval(20*time.Millisecond),
		WithMaxAge(50*time.Millisecond),
	)
	require.NoError(t, err)

	// First file created by NewRotateFile.
	dir, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 1)

	// Wait past rotate interval, write to trigger second file.
	time.Sleep(30 * time.Millisecond)
	_, err = file.Write([]byte("a\n"))
	require.NoError(t, err)

	dir, err = os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 2)

	// Wait until the first two files are older than maxAge, then write again.
	// The third rotate should prune both predecessors.
	time.Sleep(80 * time.Millisecond)
	_, err = file.Write([]byte("a\n"))
	require.NoError(t, err)

	dir, err = os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 1, "only the current file should remain")
}

func TestPruneIgnoresOtherHosts(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest-TestPruneIgnoresOtherHosts")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	// An old file from a different host that prune must NOT touch.
	foreign := filepath.Join(path, "test.20000101-000000.000000000.otherhost.99999.log")
	require.NoError(t, os.WriteFile(foreign, []byte("foreign\n"), 0o644))

	file, err := NewRotateFile(path, "test", 1024,
		WithCheckEveryN(1),
		WithRotateInterval(10*time.Millisecond),
		WithMaxAge(5*time.Millisecond),
	)
	require.NoError(t, err)

	// Trigger several rotates so prune runs repeatedly.
	for range 3 {
		time.Sleep(15 * time.Millisecond)
		_, err = file.Write([]byte("a\n"))
		require.NoError(t, err)
	}

	_, err = os.Stat(foreign)
	require.NoError(t, err, "foreign-host file must be left alone")
}

func TestPruneIgnoresMalformedNames(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest-TestPruneIgnoresMalformedNames")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	hostname, err := os.Hostname()
	require.NoError(t, err)

	// File matches the glob (same basename and hostname) but the timestamp
	// segment is garbage and cannot be parsed; prune must skip it.
	malformed := filepath.Join(path, "test.not-a-timestamp."+hostname+".99999.log")
	require.NoError(t, os.WriteFile(malformed, []byte("garbage\n"), 0o644))

	file, err := NewRotateFile(path, "test", 1024,
		WithCheckEveryN(1),
		WithRotateInterval(10*time.Millisecond),
		WithMaxAge(5*time.Millisecond),
	)
	require.NoError(t, err)

	for range 3 {
		time.Sleep(15 * time.Millisecond)
		_, err = file.Write([]byte("a\n"))
		require.NoError(t, err)
	}

	_, err = os.Stat(malformed)
	require.NoError(t, err, "malformed-name file must be left alone")
}

func TestCreateFileAfterRemove(t *testing.T) {
	path, err := os.MkdirTemp("", "loghtest-TestRotateInterval")
	require.NoError(t, err)
	defer os.RemoveAll(path)

	file, err := NewRotateFile(path, "test", 1024*1024*1024, WithCheckEveryN(3))
	require.NoError(t, err)

	b := []byte("test\n")
	for range 5 {
		_, err = file.Write(b)
		require.NoError(t, err)
	}

	os.Remove(file.file.Name())
	dir, err := os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 0)

	for range 5 {
		_, err = file.Write(b)
		require.NoError(t, err)
	}

	dir, err = os.ReadDir(path)
	require.NoError(t, err)
	require.Len(t, dir, 1)
}
