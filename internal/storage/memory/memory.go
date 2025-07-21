package memory

import (
	"errors"
	"sync"

	"github.com/pragneshbagary/go-auth/pkg/models"
)

// InMemoryStorage is a simple, thread-safe, in-memory implementation of the storage.Storage interface.
// It is intended for example and testing purposes.	ype
type InMemoryStorage struct {
	mu    sync.RWMutex
	users map[string]models.User
}

// NewInMemoryStorage creates a new in-memory storage instance.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		users: make(map[string]models.User),
	}
}

// CreateUser saves a new user to the in-memory map.
func (s *InMemoryStorage) CreateUser(user models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.Username]; exists {
		return errors.New("user already exists")
	}
	s.users[user.Username] = user
	return nil
}

// GetUserByUsername retrieves a user from the in-memory map.
func (s *InMemoryStorage) GetUserByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, ok := s.users[username]
	if !ok {
		return nil, errors.New("user not found")
	}
	return &user, nil
}
