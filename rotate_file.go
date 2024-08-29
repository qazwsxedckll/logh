package logh

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type RotateFile struct {
	filepath string
	file     *os.File

	rotateSize       int
	rotateInterval   time.Duration
	rotateAtMidnight bool
	checkEveryN      int

	written    int
	lastRotate time.Time
	count      int
}

func NewRotateFile(directory string, basename string, rotateSize int, opts ...Option) (*RotateFile, error) {
	if directory != "" {
		err := os.MkdirAll(directory, 0o755)
		if err != nil {
			return nil, err
		}
	}

	if basename == "" {
		basename = "log"
	}

	path := filepath.Join(directory, basename)

	rf := &RotateFile{
		filepath:         path,
		rotateSize:       rotateSize,
		rotateInterval:   time.Hour * 24,
		rotateAtMidnight: false,
		checkEveryN:      1024,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(rf)
		}
	}

	rf.rotate()

	return rf, nil
}

func (r *RotateFile) Write(p []byte) (int, error) {
	n, err := r.file.Write(p)
	r.written += n

	if r.written > r.rotateSize {
		r.rotate()
	} else {
		r.count++
		if r.count >= r.checkEveryN {
			r.count = 0

			_, err := os.Stat(r.file.Name())
			if errors.Is(err, os.ErrNotExist) {
				r.rotate()
			}

			if r.rotateAtMidnight {
				if time.Now().Day() != r.lastRotate.Day() {
					r.rotate()
				}
			} else {
				if time.Now().After(r.lastRotate.Add(r.rotateInterval)) {
					r.rotate()
				}
			}
		}
	}

	return n, err
}

func (r *RotateFile) logFileName() (string, time.Time) {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknownhost"
	}
	now := time.Now()
	return r.filepath + "." + now.Format("20060102-150405.000000000") + "." + hostname + "." + fmt.Sprint(os.Getpid()) + ".log", now
}

func (r *RotateFile) rotate() {
	filename, now := r.logFileName()

	if now.After(r.lastRotate) {
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			panic(fmt.Sprintf("Failed to open log file %s: %s", filename, err))
		}

		if r.file != nil {
			// TODO: error handling?
			r.file.Close()
		}
		r.file = file
		r.written = 0
		r.lastRotate = now
	}
}

type Option func(*RotateFile)

func WithCheckEveryN(n int) Option {
	return func(r *RotateFile) {
		r.checkEveryN = n
	}
}

func WithRotateInterval(d time.Duration) Option {
	return func(r *RotateFile) {
		r.rotateInterval = d
	}
}

// WithRotateAtMidnight will suppress the rotateInterval
func WithRotateAtMidnight() Option {
	return func(r *RotateFile) {
		r.rotateAtMidnight = true
	}
}
