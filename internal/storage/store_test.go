package store_test

import (
	"bytes"
	"testing"

	store "example.com/storage-workshop/internal/storage"
)

func TestBehaviour(t *testing.T) {
	// store with Put
	id := "id"
	content := []byte("test input")
	err := store.Put(id, content)
	if err != nil {
		t.Fatalf("Put failed with error %q", err)
	}

	// get it
	storedContent, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get failed wih error %q", err)
	}
	if !bytes.Equal(storedContent, content) {
		t.Errorf("Content returned did not match input content want %q, got %q", content, storedContent)
	}
}

func TestPut(t *testing.T) {

}
