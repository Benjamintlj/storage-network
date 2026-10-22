// Package workshop is your optional exercise pad. Implement these mechanisms
// independently of your main project, or adapt the tests to your own API.
// Unimplemented exercises intentionally fail. Select the chapter's test group.
package workshop

import (
	"context"
	"errors"
	"io"
	"time"
)

const MaxFrame = 8 << 20

var ErrLimit = errors.New("size limit exceeded")
var ErrShards = errors.New("invalid or insufficient shards")
var ErrIntegrity = errors.New("object integrity failure")

type Store struct {
	Stage func(string) error /* Add fields you need. */
}
type Event struct {
	ID    string
	Delta int64
}

func ReadFrame(r io.Reader) ([]byte, error)     { panic("implement Chapter 4 framing") }
func WriteFrame(w io.Writer, body []byte) error { panic("implement Chapter 4 framing") }
func Chunks(r io.Reader, size int, emit func(int, []byte) error) error {
	panic("implement Chapter 6 chunking")
}
func Encode2(data []byte) ([3][]byte, int)                    { panic("implement Chapter 7 encoding") }
func Recover2(shards [3][]byte, original int) ([]byte, error) { panic("implement Chapter 7 recovery") }
func Seal(key, plain, aad []byte) ([]byte, error)             { panic("implement Chapter 11 encryption") }
func Open(key, envelope, aad []byte) ([]byte, error)          { panic("implement Chapter 11 decryption") }
func Digest(b []byte) string                                  { panic("implement content digest") }
func NewStore(dir string) (*Store, error)                     { panic("implement Chapter 3 store") }
func (s *Store) Put(id string, b []byte) error                { panic("implement Chapter 3 put") }
func (s *Store) Get(id string) ([]byte, error)                { panic("implement Chapter 3 get") }
func Run(ctx context.Context, workers int, jobs []int, fn func(context.Context, int) error) error {
	panic("implement Chapters 13–14 workers")
}
func Health(now, last time.Time, known bool, suspect, unavailable time.Duration) (string, error) {
	panic("implement Chapter 15 health")
}
func Balance(events []Event) (int64, error) { panic("implement Chapter 23 ledger") }
