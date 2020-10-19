package ddp

// Store is a template interface for storage
type Store interface {
	C(data interface{}) error
	R(id interface{}) error
	U() error
	D(id interface{}) error
}

// implementation, one of:
// - redis
// - other kv
// - cache

type mockStore struct{}

// NewMockStore returns new mockStore
func NewMockStore() *mockStore {
	return &mockStore{}
}

// C implements Store
func (m *mockStore) C(data interface{}) error {
	return nil
}

// R implements Store
func (m *mockStore) R(id interface{}) error {
	return nil
}

// U implements Store
func (m *mockStore) U() error {
	return nil
}

// D implements Store
func (m *mockStore) D(id interface{}) error {
	return nil
}
