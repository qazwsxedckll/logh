package logh

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const rotateTimestampLayout = "20060102-150405.000000000"

type RotateFile struct {
	filepath string
	hostname string
	file     *os.File

	rotateSize       int
	rotateInterval   time.Duration
	rotateAtMidnight bool
	checkEveryN      int
	maxAge           time.Duration

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

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknownhost"
	}

	rf := &RotateFile{
		filepath:         path,
		hostname:         hostname,
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
	now := time.Now()
	return r.filepath + "." + now.Format(rotateTimestampLayout) + "." + r.hostname + "." + fmt.Sprint(os.Getpid()) + ".log", now
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

		r.prune()
	}
}

// prune removes rotated log files belonging to this instance (same basename and
// hostname) whose embedded timestamp is older than maxAge. It is best-effort:
// any glob, parse, or remove error is silently dropped so it never disrupts the
// write path. Files whose names cannot be parsed against rotateTimestampLayout
// (e.g. unrelated files or files from a different format version) are skipped.
func (r *RotateFile) prune() {
	if r.maxAge <= 0 {
		return
	}

	pattern := r.filepath + ".*." + r.hostname + ".*.log"
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-r.maxAge)
	prefix := r.filepath + "."
	hostMarker := "." + r.hostname + "."
	current := ""
	if r.file != nil {
		current = r.file.Name()
	}

	for _, name := range matches {
		if name == current {
			continue
		}

		base := strings.TrimPrefix(name, prefix)
		idx := strings.Index(base, hostMarker)
		if idx <= 0 {
			continue
		}
		ts, err := time.ParseInLocation(rotateTimestampLayout, base[:idx], time.Local)
		if err != nil {
			continue
		}

		if ts.Before(cutoff) {
			_ = os.Remove(name)
		}
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

// WithMaxAge enables pruning of rotated log files older than d. Pruning runs
// after each successful rotate. A non-positive d disables pruning. Only files
// matching this instance's basename and hostname are considered; files whose
// timestamp segment cannot be parsed are left alone. Errors during pruning are
// silently ignored to avoid disrupting the write path.
func WithMaxAge(d time.Duration) Option {
	return func(r *RotateFile) {
		r.maxAge = d
	}
}
