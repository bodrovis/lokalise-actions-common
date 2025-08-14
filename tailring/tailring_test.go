package tailring

import (
	"bytes"
	"io"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRingKeepsTail_Basic(t *testing.T) {
	r := New(8)
	writeString(t, r, "abcd")
	if got := r.String(); got != "abcd" {
		t.Fatalf("got %q, want %q", got, "abcd")
	}

	writeString(t, r, "efgh")
	if got := r.String(); got != "abcdefgh" {
		t.Fatalf("got %q, want %q", got, "abcdefgh")
	}

	writeString(t, r, "XYZ")
	if got := r.String(); got != "defghXYZ" { // last 8 bytes
		t.Fatalf("got %q, want %q", got, "defghXYZ")
	}
}

func TestRingKeepsTail_SingleChunkBiggerThanLimit(t *testing.T) {
	r := New(5)
	writeString(t, r, "0123456789") // 10 bytes
	if got := r.String(); got != "56789" {
		t.Fatalf("got %q, want %q", got, "56789")
	}
}

func TestRingLenAndCap(t *testing.T) {
	r := New(3)
	if r.Cap() != 3 {
		t.Fatalf("Cap = %d, want 3", r.Cap())
	}
	if r.Len() != 0 {
		t.Fatalf("Len initial = %d, want 0", r.Len())
	}
	writeString(t, r, "ab")
	if r.Len() != 2 {
		t.Fatalf("Len after 'ab' = %d, want 2", r.Len())
	}
	writeString(t, r, "cd") // now "bcd"
	if r.Len() != 3 {
		t.Fatalf("Len after 'cd' = %d, want 3", r.Len())
	}
}

func TestRingReset(t *testing.T) {
	r := New(4)
	writeString(t, r, "wxyz")
	if r.Len() != 4 {
		t.Fatalf("Len before reset = %d, want 4", r.Len())
	}
	r.Reset()
	if r.Len() != 0 || r.String() != "" {
		t.Fatalf("after reset got Len=%d, String=%q; want 0, \"\"", r.Len(), r.String())
	}
}

func TestZeroLimit(t *testing.T) {
	r := New(0)
	writeString(t, r, "whatever")
	if r.Len() != 0 || r.String() != "" {
		t.Fatalf("zero limit should keep nothing; got Len=%d, String=%q", r.Len(), r.String())
	}
}

func TestNegativeLimit(t *testing.T) {
	// Spec: negative behaves like zero (keeps nothing).
	r := New(-10)
	writeString(t, r, "abc")
	if r.Len() != 0 || r.String() != "" {
		t.Fatalf("negative limit should keep nothing; got Len=%d, String=%q", r.Len(), r.String())
	}
}

func TestBytesReturnsCopy(t *testing.T) {
	r := New(4)
	writeString(t, r, "abcd")
	b := r.Bytes()
	if string(b) != "abcd" {
		t.Fatalf("Bytes() = %q, want %q", string(b), "abcd")
	}
	// mutate returned slice; should not affect internal buffer
	b[0] = 'Z'
	if r.String() != "abcd" {
		t.Fatalf("internal buffer should be unchanged; got %q", r.String())
	}
}

func TestTeeWritesBoth(t *testing.T) {
	r := New(4)
	var dst bytes.Buffer
	w := Tee(&dst, r)
	io.WriteString(w, "hello")

	if dst.String() != "hello" {
		t.Fatalf("tee dst mismatch: got %q, want %q", dst.String(), "hello")
	}
	if r.String() != "ello" {
		t.Fatalf("ring tail mismatch: got %q, want %q", r.String(), "ello")
	}
}

func TestTeeWithNilDst(t *testing.T) {
	r := New(3)
	w := Tee(nil, r)
	io.WriteString(w, "abcd")
	if r.String() != "bcd" {
		t.Fatalf("got %q, want %q", r.String(), "bcd")
	}
}

func TestInterleavedWrites(t *testing.T) {
	r := New(10)
	writeString(t, r, "12345")
	writeString(t, r, "678")
	writeString(t, r, "9")
	writeString(t, r, "0")
	if got := r.String(); got != "1234567890" {
		t.Fatalf("got %q, want %q", got, "1234567890")
	}
	writeString(t, r, "ABCDE")
	if got := r.String(); got != "67890ABCDE" {
		t.Fatalf("got %q, want %q", got, "67890ABCDE")
	}
}

func TestConcurrencyHammer(t *testing.T) {
	// Not asserting exact order (since concurrent), but we check:
	// - no panic / race
	// - final length <= limit
	// - final tail equals the last N bytes of the *sequential* model if we
	//   apply writes in a deterministic order. Since we can't know order,
	//   we assert len and that all bytes are from the input alphabet.
	limit := 1024
	r := New(limit)

	const goroutines = 32
	const perG = 200
	payloads := make([][]byte, goroutines)
	for i := range payloads {
		// build a distinct line per goroutine, repeated varying times
		line := strings.Repeat(string('A'+byte(i%26)), 31) + "\n"
		payloads[i] = []byte(line)
	}

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		g := g
		go func() {
			defer wg.Done()
			rnd := rand.New(rand.NewSource(time.Now().UnixNano() + int64(g)))
			for range make([]struct{}, perG) {
				times := 1 + rnd.Intn(3)
				for range make([]struct{}, times) {
					_, _ = r.Write(payloads[g])
				}
			}
		}()
	}
	wg.Wait()

	if r.Len() > limit {
		t.Fatalf("Len=%d exceeds limit=%d", r.Len(), limit)
	}

	// sanity: all bytes should be either '\n' or 'A'..'Z'
	for _, b := range r.Bytes() {
		if b != '\n' && (b < 'A' || b > 'Z') {
			t.Fatalf("unexpected byte %q in ring tail", b)
		}
	}
}

func TestHugeChunkFollowedBySmall_WrapsCorrectly(t *testing.T) {
	r := New(6)
	writeString(t, r, "XXXXXXXXXXXX") // 12
	if got := r.String(); got != "XXXXXX" {
		t.Fatalf("got %q, want %q", got, "XXXXXX")
	}
	writeString(t, r, "12")
	if got := r.String(); got != "XXXX12" {
		t.Fatalf("got %q, want %q", got, "XXXX12")
	}
}

func writeString(t *testing.T, w io.Writer, s string) {
	t.Helper()
	if _, err := io.WriteString(w, s); err != nil {
		t.Fatalf("write %q failed: %v", s, err)
	}
}
