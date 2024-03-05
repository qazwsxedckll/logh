package logh

import (
	"io"
	"log/slog"
)

func NewRotateJSONHandler(rollSize int, flushInterval int, opts *slog.HandlerOptions) (slog.Handler, error) {
	var w io.Writer

	return slog.NewJSONHandler(w, opts), nil
}
