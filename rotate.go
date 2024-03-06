package logh

import (
	"log/slog"
)

func NewRotateJSONHandler(directory string, basename string, rollSize int, opts *slog.HandlerOptions, options ...Option) (slog.Handler, error) {
	file, err := NewRotateFile(directory, basename, rollSize, options...)
	if err != nil {
		return nil, err
	}

	return slog.NewJSONHandler(file, opts), nil
}
