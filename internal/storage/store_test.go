package store_test

import (
	"bytes"
	"os"
	"testing"

	store "example.com/storage-workshop/internal/storage"
)

func TestBehaviour(t *testing.T) {
	dir := "./test-store"
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatalf("Failed to create the temp dir %q", err)
	}
	defer func(name string) {
		err := os.RemoveAll(name)
		if err != nil {
			t.Fatalf("Failed to remove the temp dir %q", err)
		}
	}(dir)

	nonVolatileStore, err := store.Open("./test-store")
	if err != nil {
		t.Fatalf("Failed to open test directory %q", err)
	}

	// store with Put
	id := "id"
	content := []byte("test input")
	err = nonVolatileStore.Put(id, content)
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
}

func TestPut(t *testing.T) {

}
