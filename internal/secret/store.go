package secret

import (
	"fmt"
	"sync"

	"github.com/zalando/go-keyring"
)

type Store interface {
	Set(provider, profile, name, value string) error
	Get(provider, profile, name string) (string, error)
	Delete(provider, profile, name string) error
}

const serviceName = "api-cli"

var (
	storeMu sync.RWMutex
	store   Store = keychainStore{}
)

type keychainStore struct{}

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: map[string]string{}}
}

func SetStore(s Store) {
	storeMu.Lock()
	defer storeMu.Unlock()
	if s == nil {
		store = keychainStore{}
		return
	}
	store = s
}

func Set(provider, profile, name, value string) error {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return store.Set(provider, profile, name, value)
}

func Get(provider, profile, name string) (string, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return store.Get(provider, profile, name)
}

func Delete(provider, profile, name string) error {
	storeMu.RLock()
	defer storeMu.RUnlock()
	return store.Delete(provider, profile, name)
}

func (keychainStore) Set(provider, profile, name, value string) error {
	return keyring.Set(serviceName, key(provider, profile, name), value)
}

func (keychainStore) Get(provider, profile, name string) (string, error) {
	return keyring.Get(serviceName, key(provider, profile, name))
}

func (keychainStore) Delete(provider, profile, name string) error {
	return keyring.Delete(serviceName, key(provider, profile, name))
}

func (m *MemoryStore) Set(provider, profile, name, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		m.data = map[string]string{}
	}
	m.data[key(provider, profile, name)] = value
	return nil
}

func (m *MemoryStore) Get(provider, profile, name string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.data == nil {
		return "", fmt.Errorf("secret not found: %s", name)
	}
	val, ok := m.data[key(provider, profile, name)]
	if !ok {
		return "", fmt.Errorf("secret not found: %s", name)
	}
	return val, nil
}

func (m *MemoryStore) Delete(provider, profile, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		return nil
	}
	delete(m.data, key(provider, profile, name))
	return nil
}

func key(provider, profile, name string) string {
	return fmt.Sprintf("%s/%s/%s", provider, profile, name)
}
