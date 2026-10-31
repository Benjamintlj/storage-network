package store

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type nvsError string

func (err nvsError) Error() string {
	return string(err)
}

var (
	ErrAlreadyExists = nvsError("Id already exists")
	ErrDirNotExists  = nvsError("Directory does not exist")
	ErrIsFileNotDir  = nvsError("Expected directory not a file")
	ErrConflict      = nvsError("Conflict, identifier already exists with different data")
)

type NonVolatileStore struct {
	dir string
}

func Open(dir string) (*NonVolatileStore, error) {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrDirNotExists
	} else if err != nil {
		return nil, err
	} else if !info.IsDir() {
		return nil, ErrIsFileNotDir
	}

	return &NonVolatileStore{dir: dir}, nil
}

func (nvs *NonVolatileStore) Get(id string) ([]byte, error) {
	file, err := os.Open(filepath.Join(nvs.dir, id))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	data, err := io.ReadAll(file)
	return data, err
}

func (nvs *NonVolatileStore) Put(id string, b []byte) error {
	file, err := os.OpenFile(filepath.Join(nvs.dir, id), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return ErrAlreadyExists
	}
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(b)
	return err
}
