package logh

import (
	"log/slog"
)

func NewRotateJSONHandler(directory string, basename string, rotateSize int, opts *slog.HandlerOptions, options ...Option) (slog.Handler, error) {
	file, err := NewRotateFile(directory, basename, rotateSize, options...)
	if err != nil {
		return nil, err
	}

	return slog.NewJSONHandler(file, opts), nil
}
