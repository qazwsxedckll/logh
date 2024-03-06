package logh

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRotateJSONHandler(t *testing.T) {
	handler, err := NewRotateJSONHandler("test", "test", 10, nil, nil)
	require.NoError(t, err)

	logger := slog.New(handler)
	logger.Info("test")
}
