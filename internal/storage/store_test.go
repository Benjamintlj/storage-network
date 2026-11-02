package store_test

import (
	"bytes"
	"errors"
	"os"
	"testing"

	store "example.com/storage-workshop/internal/storage"
)

func TestBehaviour(t *testing.T) {
	t.Run("put_then_get_returns_same_bytes", func(t *testing.T) {
		nonVolatileStore, closer := CreateStore(t)
		defer closer()

		// store with Put
		id := "id"
		content := []byte("test input")
		err := nonVolatileStore.Put(id, content)
		if err != nil {
			t.Fatalf("Put failed with error %q", err)
		}

		// get it
		storedContent, err := nonVolatileStore.Get(id)
		if err != nil {
			t.Fatalf("Get failed with error %q", err)
		}
		if !bytes.Equal(storedContent, content) {
			t.Errorf("Content returned did not match input content want %q, got %q", content, storedContent)
		}
	})
}

func TestPutExpectError(t *testing.T) {
	type putCall struct {
		value         []byte
		expectedError error
	}

	tests := []struct {
		name  string
		calls []putCall
	}{
		{
			name: "duplicate_write_returns_err",
			calls: []putCall{
				{value: []byte("test input"), expectedError: nil},
				{value: []byte("test input"), expectedError: store.ErrAlreadyExists},
			},
		},
		{
			name: "duplicate_write_different_bytes_returns_conflict_err",
			calls: []putCall{
				{value: []byte("test1 input"), expectedError: nil},
				{value: []byte("test2 input"), expectedError: store.ErrConflict},
			},
		},
		{
			name: "write_over_max_size_returns_write_too_large_err",
			calls: []putCall{
				{value: make([]byte, store.MaxSize), expectedError: store.ErrWriteTooLarge},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			nonVolatileStore, closer := CreateStore(t)
			defer closer()

			for callIndex, call := range test.calls {
				err := nonVolatileStore.Put("id", call.value)
				if !errors.Is(err, call.expectedError) {
					t.Fatalf("Put call %d: want error %v, got %v", callIndex+1, call.expectedError, err)
				}
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Run("unknown_identifier_returns_error", func(t *testing.T) {
		nonVolatileStore, closer := CreateStore(t)
		defer closer()

		_, err := nonVolatileStore.Get("id")
		if !errors.Is(err, store.ErrUnknownId) {
			t.Fatalf("Get failed with the wrong error got %q, want %q", err, store.ErrUnknownId)
		}
	})
}

// Helpers
func CreateStore(tb testing.TB) (*store.NonVolatileStore, func()) {
	const dir = "./test-store"
	if err := os.MkdirAll(dir, 0o700); err != nil {
		tb.Fatalf("Failed to create the temp dir %q", err)
	}
	closer := func() {
		err := os.RemoveAll(dir)
		if err != nil {
			tb.Fatalf("Failed to remove the temp dir %q", err)
		}
	}

	nonVolatileStore, err := store.Open(dir)
	if err != nil {
		tb.Fatalf("Failed to open test directory %q", err)
	}

	return nonVolatileStore, closer
}
