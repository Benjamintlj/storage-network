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
			t.Fatalf("Get failed wih error %q", err)
		}
		if !bytes.Equal(storedContent, content) {
			t.Errorf("Content returned did not match input content want %q, got %q", content, storedContent)
		}
	})
}

func TestPut(t *testing.T) {
	t.Run("duplicate_write_returns_err", func(t *testing.T) {
		nonVolatileStore, closer := CreateStore(t)
		defer closer()

		id := "id"
		content := []byte("test input")
		err := nonVolatileStore.Put(id, content)
		if err != nil {
			t.Fatalf("Put failed early with error %q", err)
		}

		err = nonVolatileStore.Put(id, content)
		if !errors.Is(err, store.ErrAlreadyExists) {
			if err == nil {
				t.Fatalf("Put failed no error, want: %q", store.ErrAlreadyExists)
			}
			t.Fatalf("Put failed with wrong error, want: %q got: %q", store.ErrAlreadyExists, err)
		}
	})

	t.Run("duplicate_write_different_bytes_returns_conflict_err", func(t *testing.T) {
		nonVolatileStore, closer := CreateStore(t)
		defer closer()

		id := "id"
		content1 := []byte("test1 input")
		content2 := []byte("test2 input")
		err := nonVolatileStore.Put(id, content1)
		if err != nil {
			t.Fatalf("Put failed early with error %q", err)
		}

		err = nonVolatileStore.Put(id, content2)
		if !errors.Is(err, store.ErrConflict) {
			if err == nil {
				t.Fatalf("Put failed no error, want: %q", store.ErrConflict)
			}
			t.Fatalf("Put failed with wrong error, want: %q got: %q", store.ErrAlreadyExists, err)
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
