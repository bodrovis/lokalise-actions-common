package tailring

import (
	"io"
	"sync"
)

// Ring is a thread-safe ring buffer that implements io.Writer
// and keeps only the last N bytes written.
type Ring struct {
	mu    sync.Mutex
	buf   []byte
	limit int
}

// New creates a Ring that keeps at most limit bytes.
func New(limit int) *Ring {
	return &Ring{limit: limit}
}

func NewKB(kb int) *Ring {
	return New(kb * 1024)
}

// Write appends p, keeping only the last limit bytes.
func (r *Ring) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.limit <= 0 {
		return len(p), nil
	}

	// If chunk alone exceeds limit, keep its tail.
	if len(p) >= r.limit {
		if cap(r.buf) < r.limit {
			r.buf = make([]byte, 0, r.limit)
		}
		r.buf = append(r.buf[:0], p[len(p)-r.limit:]...)
		return len(p), nil
	}

	need := len(r.buf) + len(p) - r.limit
	if need > 0 {
		r.buf = r.buf[need:]
	}
	r.buf = append(r.buf, p...)
	return len(p), nil
}

// Bytes returns a copy of the tail.
func (r *Ring) Bytes() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]byte, len(r.buf))
	copy(out, r.buf)
	return out
}

// String returns the tail as string (copy).
func (r *Ring) String() string {
	return string(r.Bytes())
}

// Reset clears the buffer.
func (r *Ring) Reset() {
	r.mu.Lock()
	r.buf = r.buf[:0]
	r.mu.Unlock()
}

// Len returns current stored bytes; Cap is the limit.
func (r *Ring) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.buf)
}

func (r *Ring) Cap() int { return r.limit }

// Tee returns an io.Writer that writes to both dst and ring.
// Handy for cmd.Stdout/StdErr: io.MultiWriter(dst, ring)
func Tee(dst io.Writer, ring *Ring) io.Writer {
	if dst == nil {
		return ring
	}
	return io.MultiWriter(dst, ring)
}
