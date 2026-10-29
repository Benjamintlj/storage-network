package store

type NonVolatileStore struct {
}

func Open(dir string) (*NonVolatileStore, error) {
	return nil, nil
}

func (nvs *NonVolatileStore) Get(id string) ([]byte, error) {
	return nil, nil
}

// Todo: Write persistently
// Todo: Duplicate write (same bytes & id), are successful, one object
// Todo: Different bytes same identifier, conflict
// Todo: Unknown identifier, unknown
// Todo: Do not allow writes over the size of 8MiB, this memory must be pre-allocated

// Todo: Empty object, binary object, text
// Todo: Stop the program completely, objects should still be readable
// Todo: Block ids that may act as paths such as `../`
func (nvs *NonVolatileStore) Put(id string, b []byte) error {
	return nil
}
