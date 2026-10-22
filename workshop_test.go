package workshop

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type tinyReader struct {
	r io.Reader
	n int
}

func (r tinyReader) Read(p []byte) (int, error) { return r.r.Read(p[:min(len(p), r.n)]) }

type shortWriter struct{ bytes.Buffer }

func (w *shortWriter) Write(p []byte) (int, error) { return w.Buffer.Write(p[:min(len(p), 2)]) }

type stuckWriter struct{}

func (stuckWriter) Write([]byte) (int, error) { return 0, nil }

type brokenWriter struct{}

func (brokenWriter) Write(p []byte) (int, error) {
	if len(p) > 0 {
		return 1, io.ErrClosedPipe
	}
	return 0, io.ErrClosedPipe
}

type endReader struct {
	done     bool
	terminal error
}

func (r *endReader) Read(p []byte) (int, error) {
	if r.done {
		return 0, io.EOF
	}
	r.done = true
	n := copy(p, []byte("abc"))
	return n, r.terminal
}

func TestFrames(t *testing.T) {
	for _, payload := range [][]byte{{}, []byte("hello"), {0, 255, 13, 10, 0}} {
		t.Run(fmt.Sprintf("roundtrip_%d", len(payload)), func(t *testing.T) {
			var w shortWriter
			if err := WriteFrame(&w, payload); err != nil {
				t.Fatal(err)
			}
			if err := WriteFrame(&w, []byte("second")); err != nil {
				t.Fatal(err)
			}
			r := tinyReader{&w, 1}
			got, err := ReadFrame(r)
			if err != nil || !bytes.Equal(got, payload) {
				t.Fatalf("%x %v", got, err)
			}
			got, err = ReadFrame(r)
			if err != nil || string(got) != "second" {
				t.Fatalf("second: %q %v", got, err)
			}
			if _, err = ReadFrame(r); !errors.Is(err, io.EOF) {
				t.Fatalf("expected clean EOF, got %v", err)
			}
		})
	}
	t.Run("every_truncation", func(t *testing.T) {
		var b bytes.Buffer
		_ = WriteFrame(&b, []byte("sample"))
		wire := b.Bytes()
		for i := 0; i < len(wire); i++ {
			_, err := ReadFrame(bytes.NewReader(wire[:i]))
			if i == 0 {
				if !errors.Is(err, io.EOF) {
					t.Fatal(err)
				}
			} else if !errors.Is(err, io.ErrUnexpectedEOF) {
				t.Fatalf("cut %d: %v", i, err)
			}
		}
	})
	t.Run("limits", func(t *testing.T) {
		var p [4]byte
		binary.BigEndian.PutUint32(p[:], MaxFrame+1)
		if _, err := ReadFrame(bytes.NewReader(p[:])); !errors.Is(err, ErrLimit) {
			t.Fatal(err)
		}
		if err := WriteFrame(io.Discard, make([]byte, MaxFrame+1)); !errors.Is(err, ErrLimit) {
			t.Fatal(err)
		}
	})
	t.Run("writer_errors", func(t *testing.T) {
		if err := WriteFrame(stuckWriter{}, []byte("x")); !errors.Is(err, io.ErrShortWrite) {
			t.Fatal(err)
		}
		if err := WriteFrame(brokenWriter{}, []byte("x")); !errors.Is(err, io.ErrClosedPipe) {
			t.Fatal(err)
		}
	})
}

func TestChunks(t *testing.T) {
	for _, length := range []int{0, 1, 6, 7, 8, 14, 15, 101} {
		t.Run(fmt.Sprint(length), func(t *testing.T) {
			input := make([]byte, length)
			for i := range input {
				input[i] = byte(i)
			}
			var retained [][]byte
			err := Chunks(tinyReader{bytes.NewReader(input), 2}, 7, func(index int, b []byte) error {
				if index != len(retained) {
					t.Fatalf("index %d", index)
				}
				retained = append(retained, b)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(retained) != (length+6)/7 {
				t.Fatalf("wrong chunk count %d", len(retained))
			}
			for i, b := range retained {
				if i < len(retained)-1 && len(b) != 7 {
					t.Fatal("short interior chunk")
				}
			}
			if !bytes.Equal(bytes.Join(retained, nil), input) {
				t.Fatal("content or buffer ownership")
			}
		})
	}
	t.Run("bytes_with_eof", func(t *testing.T) {
		var out []byte
		err := Chunks(&endReader{terminal: io.EOF}, 8, func(_ int, b []byte) error { out = append(out, b...); return nil })
		if err != nil || string(out) != "abc" {
			t.Fatalf("%q %v", out, err)
		}
	})
	t.Run("source_error", func(t *testing.T) {
		for _, want := range []error{io.ErrClosedPipe, io.ErrUnexpectedEOF} {
			err := Chunks(&endReader{terminal: want}, 8, func(int, []byte) error { return nil })
			if !errors.Is(err, want) {
				t.Fatalf("want %v got %v", want, err)
			}
		}
	})
	t.Run("callback_error", func(t *testing.T) {
		want := errors.New("stop")
		err := Chunks(bytes.NewBufferString("abc"), 2, func(int, []byte) error { return want })
		if !errors.Is(err, want) {
			t.Fatal(err)
		}
	})
	t.Run("bad_size", func(t *testing.T) {
		for _, size := range []int{-1, 0, MaxFrame + 1} {
			if err := Chunks(bytes.NewReader(nil), size, func(int, []byte) error { return nil }); err == nil {
				t.Fatal("accepted invalid size")
			}
		}
	})
}

func TestErasure(t *testing.T) {
	for _, length := range []int{0, 1, 2, 3, 17, 1025} {
		t.Run(fmt.Sprint(length), func(t *testing.T) {
			input := make([]byte, length)
			for i := range input {
				input[i] = byte(i * 73)
			}
			encoded, n := Encode2(input)
			for missing := -1; missing < 3; missing++ {
				shards := encoded
				if missing >= 0 {
					shards[missing] = nil
				}
				got, err := Recover2(shards, n)
				if err != nil || !bytes.Equal(got, input) {
					t.Fatalf("missing %d: %v", missing, err)
				}
			}
			for keep := 0; keep < 3; keep++ {
				var shards [3][]byte
				shards[keep] = encoded[keep]
				if _, err := Recover2(shards, n); err == nil {
					t.Fatal("accepted one shard")
				}
			}
		})
	}
	t.Run("malformed", func(t *testing.T) {
		shards, n := Encode2([]byte("abc"))
		if _, err := Recover2(shards, -1); err == nil {
			t.Fatal("negative size")
		}
		if _, err := Recover2(shards, n+10); err == nil {
			t.Fatal("oversized output")
		}
		shards[1] = append(shards[1], 0)
		if _, err := Recover2(shards, n); err == nil {
			t.Fatal("unequal shards")
		}
	})
	t.Run("does_not_mutate", func(t *testing.T) {
		shards, n := Encode2([]byte("abcdef"))
		before := append([]byte(nil), shards[1]...)
		shards[0] = nil
		_, _ = Recover2(shards, n)
		if !bytes.Equal(before, shards[1]) {
			t.Fatal("mutated input")
		}
	})
}

func TestEncryption(t *testing.T) {
	key := bytes.Repeat([]byte{42}, 32)
	aad := []byte("owner/version/stripe-1")
	for _, plain := range [][]byte{nil, []byte("private photograph"), make([]byte, 10001)} {
		t.Run(fmt.Sprint(len(plain)), func(t *testing.T) {
			a, err := Seal(key, plain, aad)
			if err != nil {
				t.Fatal(err)
			}
			b, err := Seal(key, plain, aad)
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Equal(a, b) {
				t.Fatal("fresh encryptions reused envelope")
			}
			got, err := Open(key, a, aad)
			if err != nil || !bytes.Equal(got, plain) {
				t.Fatal("round trip", err)
			}
			for _, pos := range []int{0, 12, len(a) - 1} {
				bad := append([]byte(nil), a...)
				bad[pos] ^= 1
				got, err := Open(key, bad, aad)
				if err == nil || got != nil {
					t.Fatal("tampering accepted")
				}
			}
			if got, err := Open(key, a, []byte("different context")); err == nil || got != nil {
				t.Fatal("wrong context")
			}
			wrong := append([]byte(nil), key...)
			wrong[0] ^= 1
			if _, err := Open(wrong, a, aad); err == nil {
				t.Fatal("wrong key")
			}
			shards, n := Encode2(a)
			shards[1] = nil
			recovered, err := Recover2(shards, n)
			if err != nil {
				t.Fatal(err)
			}
			got, err = Open(key, recovered, aad)
			if err != nil || !bytes.Equal(got, plain) {
				t.Fatal("encrypted erasure recovery")
			}
		})
	}
	if _, err := Seal([]byte{1}, nil, nil); err == nil {
		t.Fatal("bad key")
	}
	if _, err := Open(key, []byte{1}, aad); err == nil {
		t.Fatal("short envelope")
	}
}

func TestStore(t *testing.T) {
	t.Run("restart_and_integrity", func(t *testing.T) {
		dir := t.TempDir()
		s, err := NewStore(dir)
		if err != nil {
			t.Fatal(err)
		}
		b := []byte("immutable bytes")
		id := Digest(b)
		if err = s.Put(id, b); err != nil {
			t.Fatal(err)
		}
		if err = s.Put(id, b); err != nil {
			t.Fatal(err)
		}
		if err = s.Put(id, []byte("different")); err == nil {
			t.Fatal("conflict accepted")
		}
		s, err = NewStore(dir)
		if err != nil {
			t.Fatal(err)
		}
		got, err := s.Get(id)
		if err != nil || !bytes.Equal(got, b) {
			t.Fatal("restart", err)
		}
		if _, err = s.Get(Digest([]byte("absent"))); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("missing", err)
		}
		if _, err = s.Get("../escape"); err == nil {
			t.Fatal("invalid identifier")
		}
		if err = os.WriteFile(filepath.Join(dir, id), []byte("corrupt"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err = s.Get(id); !errors.Is(err, ErrIntegrity) {
			t.Fatal("corruption", err)
		}
	})
	for _, stage := range []string{"staged", "file-synced", "published", "directory-synced"} {
		t.Run("failure_"+stage, func(t *testing.T) {
			dir := t.TempDir()
			s, err := NewStore(dir)
			if err != nil {
				t.Fatal(err)
			}
			want := errors.New("injected crash boundary")
			s.Stage = func(at string) error {
				if at == stage {
					return want
				}
				return nil
			}
			b := []byte("recoverable")
			id := Digest(b)
			if err = s.Put(id, b); !errors.Is(err, want) {
				t.Fatal(err)
			}
			s.Stage = nil
			got, err := s.Get(id)
			if stage == "staged" || stage == "file-synced" {
				if !errors.Is(err, os.ErrNotExist) {
					t.Fatal("premature publication", err)
				}
			} else if err != nil || !bytes.Equal(got, b) {
				t.Fatal("published content", err)
			}
			if err = s.Put(id, b); err != nil {
				t.Fatal("retry", err)
			}
			entries, _ := os.ReadDir(dir)
			if len(entries) != 1 {
				t.Fatalf("expected one final object, got %d", len(entries))
			}
		})
	}
	t.Run("concurrent_duplicates", func(t *testing.T) {
		dir := t.TempDir()
		s, err := NewStore(dir)
		if err != nil {
			t.Fatal(err)
		}
		b := []byte("same object")
		id := Digest(b)
		var wg sync.WaitGroup
		for i := 0; i < 12; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if err := s.Put(id, b); err != nil {
					t.Error(err)
				}
			}()
		}
		wg.Wait()
		entries, _ := os.ReadDir(dir)
		if len(entries) != 1 {
			t.Fatal("duplicate physical objects")
		}
	})
}

func TestWorkers(t *testing.T) {
	t.Run("bounded_and_joined", func(t *testing.T) {
		var active, peak, completed atomic.Int32
		entered := make(chan struct{}, 3)
		release := make(chan struct{})
		done := make(chan error, 1)
		go func() {
			done <- Run(context.Background(), 3, []int{0, 1, 2, 3, 4, 5}, func(ctx context.Context, _ int) error {
				n := active.Add(1)
				defer active.Add(-1)
				for old := peak.Load(); n > old; old = peak.Load() {
					if peak.CompareAndSwap(old, n) {
						break
					}
				}
				select {
				case entered <- struct{}{}:
				default:
				}
				select {
				case <-release:
					completed.Add(1)
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})
		}()
		for i := 0; i < 3; i++ {
			select {
			case <-entered:
			case <-time.After(time.Second):
				t.Fatal("workers did not enter")
			}
		}
		if peak.Load() > 3 {
			t.Fatal("unbounded workers")
		}
		close(release)
		select {
		case err := <-done:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("did not join")
		}
		if active.Load() != 0 || completed.Load() != 6 {
			t.Fatal("unfinished workers")
		}
	})
	t.Run("cancel_blocked_callbacks", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		started := make(chan struct{})
		done := make(chan error, 1)
		var once sync.Once
		var active atomic.Int32
		go func() {
			done <- Run(ctx, 2, []int{1, 2, 3}, func(ctx context.Context, _ int) error {
				active.Add(1)
				defer active.Add(-1)
				once.Do(func() { close(started) })
				<-ctx.Done()
				return ctx.Err()
			})
		}()
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("not started")
		}
		cancel()
		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Fatal(err)
			}
		case <-time.After(time.Second):
			t.Fatal("cancellation did not join")
		}
		if active.Load() != 0 {
			t.Fatal("callback survived operation")
		}
	})
	if err := Run(context.Background(), 0, nil, func(context.Context, int) error { return nil }); err == nil {
		t.Fatal("invalid worker count")
	}
}

func TestHealth(t *testing.T) {
	start := time.Unix(1000, 0)
	for _, tc := range []struct {
		age  time.Duration
		want string
	}{{0, "alive"}, {9 * time.Second, "alive"}, {10 * time.Second, "suspect"}, {29 * time.Second, "suspect"}, {30 * time.Second, "unavailable"}} {
		got, err := Health(start.Add(tc.age), start, true, 10*time.Second, 30*time.Second)
		if err != nil || got != tc.want {
			t.Fatalf("age %v: %s %v", tc.age, got, err)
		}
	}
	if got, _ := Health(start, start, false, time.Second, 2*time.Second); got != "unknown" {
		t.Fatal(got)
	}
	if _, err := Health(start, start, true, 0, time.Second); err == nil {
		t.Fatal("invalid thresholds")
	}
	if _, err := Health(start.Add(-time.Second), start, true, time.Second, 2*time.Second); err == nil {
		t.Fatal("future observation")
	}
}

func TestLedger(t *testing.T) {
	events := []Event{{"store-1", 100}, {"use-1", -20}, {"store-1", 100}, {"reverse-use-1", 20}}
	if got, err := Balance(events); err != nil || got != 100 {
		t.Fatalf("%d %v", got, err)
	}
	for i, j := 0, len(events)-1; i < j; i, j = i+1, j-1 {
		events[i], events[j] = events[j], events[i]
	}
	if got, err := Balance(events); err != nil || got != 100 {
		t.Fatalf("reordered: %d %v", got, err)
	}
	if _, err := Balance([]Event{{"x", 1}, {"x", 2}}); err == nil {
		t.Fatal("conflicting ID")
	}
	if _, err := Balance([]Event{{"", 1}}); err == nil {
		t.Fatal("empty ID")
	}
	if _, err := Balance([]Event{{"x", math.MaxInt64}, {"y", 1}}); err == nil {
		t.Fatal("overflow")
	}
}

func FuzzFrames(f *testing.F) {
	f.Add([]byte("hello"))
	f.Add([]byte{0, 255})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, p []byte) {
		if len(p) > MaxFrame {
			t.Skip()
		}
		var b bytes.Buffer
		if err := WriteFrame(&b, p); err != nil {
			t.Fatal(err)
		}
		got, err := ReadFrame(tinyReader{&b, 3})
		if err != nil || !bytes.Equal(got, p) {
			t.Fatal("round trip", err)
		}
	})
}
