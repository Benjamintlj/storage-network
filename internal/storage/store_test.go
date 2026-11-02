package store_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	store "example.com/storage-workshop/internal/storage"
)

func TestBehaviour(t *testing.T) {
	type putStep struct {
		id            string
		content       []byte
		expectedError error
	}
	type getStep struct {
		id              string
		expectedContent []byte
		expectedError   error
	}
	type step struct {
		put *putStep
		get *getStep
	}

	tests := []struct {
		name  string
		steps []step
	}{
		{
			name: "put_then_get_returns_same_bytes",
			steps: []step{
				{put: &putStep{id: "id", content: []byte("test input"), expectedError: nil}},
				{get: &getStep{id: "id", expectedContent: []byte("test input"), expectedError: nil}},
			},
		},
		{
			name: "put_empty_array_returns_same_bytes",
			steps: []step{
				{put: &putStep{id: "id", content: make([]byte, 10), expectedError: nil}},
				{get: &getStep{id: "id", expectedContent: make([]byte, 10), expectedError: nil}},
			},
		},
		{
			name: "duplicate_write_returns_err",
			steps: []step{
				{put: &putStep{id: "id", content: []byte("test input"), expectedError: nil}},
				{put: &putStep{id: "id", content: []byte("test input"), expectedError: store.ErrAlreadyExists}},
			},
		},
		{
			name: "duplicate_write_different_bytes_returns_conflict_err",
			steps: []step{
				{put: &putStep{id: "id", content: []byte("test1 input"), expectedError: nil}},
				{put: &putStep{id: "id", content: []byte("test2 input"), expectedError: store.ErrConflict}},
			},
		},
		{
			name: "write_over_max_size_returns_write_too_large_err",
			steps: []step{
				{put: &putStep{id: "id", content: make([]byte, store.MaxSize+1), expectedError: store.ErrWriteTooLarge}},
			},
		},
		{
			name: "unknown_identifier_returns_error",
			steps: []step{
				{get: &getStep{id: "id", expectedContent: nil, expectedError: store.ErrUnknownId}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			const dir = "./test-store"
			if err := os.MkdirAll(dir, 0o700); err != nil {
				t.Fatalf("Failed to create the temp dir %q", err)
			}
			defer func() {
				err := os.RemoveAll(dir)
				if err != nil {
					t.Fatalf("Failed to remove the temp dir %q", err)
				}
			}()

			nonVolatileStore, err := store.Open(dir)
			if err != nil {
				t.Fatalf("Failed to open test directory %q", err)
			}

			for stepIndex, step := range test.steps {
				switch {
				case step.put != nil:
					err := nonVolatileStore.Put(step.put.id, step.put.content)
					if !errors.Is(err, step.put.expectedError) {
						t.Fatalf("step %d Put: want error %v, got %v", stepIndex+1, step.put.expectedError, err)
					}
				case step.get != nil:
					got, err := nonVolatileStore.Get(step.get.id)
					if !errors.Is(err, step.get.expectedError) {
						t.Fatalf("step %d Get: want error %v, got %v", stepIndex+1, step.get.expectedError, err)
					}
					if !bytes.Equal(got, step.get.expectedContent) {
						t.Fatalf("step %d Get: want content %q, got %q", stepIndex+1, step.get.expectedContent, got)
					}
				default:
					t.Fatalf("step %d has no Put or Get call", stepIndex+1)
				}
			}
		})
	}
}

func ExampleNonVolatileStore() {
	dir, _ := os.MkdirTemp("", "demo")
	nonVolatileStore, _ := store.Open(dir)

	_ = nonVolatileStore.Put("id", []byte("hello there"))

	content, _ := nonVolatileStore.Get("id")
	fmt.Println(string(content))

	// output: hello there
}
