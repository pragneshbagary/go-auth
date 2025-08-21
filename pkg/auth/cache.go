package auth

import (
	"fmt"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Token validation caching
	SetTokenValidation(tokenID string, user *models.User, ttl time.Duration) error
	GetTokenValidation(tokenID string) (*models.User, bool, error)
	InvalidateToken(tokenID string) error
	
	// User data caching
	SetUser(userID string, user *models.User, ttl time.Duration) error
	GetUser(userID string) (*models.User, bool, error)
	InvalidateUser(userID string) error
	
	// Blacklist caching
	SetTokenBlacklist(tokenID string, ttl time.Duration) error
	IsTokenBlacklisted(tokenID string) (bool, bool, error) // value, found, error
	InvalidateTokenBlacklist(tokenID string) error
	
	// General operations
	Close() error
	Ping() error
}

// NoOpCache is a cache implementation that does nothing (disabled caching)
type NoOpCache struct{}

func NewNoOpCache() *NoOpCache {
	return &NoOpCache{}
}

func (c *NoOpCache) SetTokenValidation(tokenID string, user *models.User, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) GetTokenValidation(tokenID string) (*models.User, bool, error) {
	return nil, false, nil
}

func (c *NoOpCache) InvalidateToken(tokenID string) error {
	return nil
}

func (c *NoOpCache) SetUser(userID string, user *models.User, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) GetUser(userID string) (*models.User, bool, error) {
	return nil, false, nil
}

func (c *NoOpCache) InvalidateUser(userID string) error {
	return nil
}

func (c *NoOpCache) SetTokenBlacklist(tokenID string, ttl time.Duration) error {
	return nil
}

func (c *NoOpCache) IsTokenBlacklisted(tokenID string) (bool, bool, error) {
	return false, false, nil
}

func (c *NoOpCache) InvalidateTokenBlacklist(tokenID string) error {
	return nil
}

func (c *NoOpCache) Close() error {
	return nil
}

func (c *NoOpCache) Ping() error {
	return nil
}

// MemoryCache is an in-memory cache implementation for development/testing
type MemoryCache struct {
	tokenValidations map[string]*cacheEntry
	users           map[string]*cacheEntry
	blacklist       map[string]*cacheEntry
}

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

func NewMemoryCache() *MemoryCache {
	cache := &MemoryCache{
		tokenValidations: make(map[string]*cacheEntry),
		users:           make(map[string]*cacheEntry),
		blacklist:       make(map[string]*cacheEntry),
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

func (c *MemoryCache) SetTokenValidation(tokenID string, user *models.User, ttl time.Duration) error {
	c.tokenValidations[tokenID] = &cacheEntry{
		data:      user,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (c *MemoryCache) GetTokenValidation(tokenID string) (*models.User, bool, error) {
	entry, exists := c.tokenValidations[tokenID]
	if !exists || time.Now().After(entry.expiresAt) {
		if exists {
			delete(c.tokenValidations, tokenID)
		}
		return nil, false, nil
	}
	
	user, ok := entry.data.(*models.User)
	if !ok {
		return nil, false, fmt.Errorf("invalid cache entry type")
	}
	
	return user, true, nil
}

func (c *MemoryCache) InvalidateToken(tokenID string) error {
	delete(c.tokenValidations, tokenID)
	return nil
}

func (c *MemoryCache) SetUser(userID string, user *models.User, ttl time.Duration) error {
	c.users[userID] = &cacheEntry{
		data:      user,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (c *MemoryCache) GetUser(userID string) (*models.User, bool, error) {
	entry, exists := c.users[userID]
	if !exists || time.Now().After(entry.expiresAt) {
		if exists {
			delete(c.users, userID)
		}
		return nil, false, nil
	}
	
	user, ok := entry.data.(*models.User)
	if !ok {
		return nil, false, fmt.Errorf("invalid cache entry type")
	}
	
	return user, true, nil
}

func (c *MemoryCache) InvalidateUser(userID string) error {
	delete(c.users, userID)
	return nil
}

func (c *MemoryCache) SetTokenBlacklist(tokenID string, ttl time.Duration) error {
	c.blacklist[tokenID] = &cacheEntry{
		data:      true,
		expiresAt: time.Now().Add(ttl),
	}
	return nil
}

func (c *MemoryCache) IsTokenBlacklisted(tokenID string) (bool, bool, error) {
	entry, exists := c.blacklist[tokenID]
	if !exists || time.Now().After(entry.expiresAt) {
		if exists {
			delete(c.blacklist, tokenID)
		}
		return false, false, nil
	}
	
	return true, true, nil
}

func (c *MemoryCache) InvalidateTokenBlacklist(tokenID string) error {
	delete(c.blacklist, tokenID)
	return nil
}

func (c *MemoryCache) Close() error {
	return nil
}

func (c *MemoryCache) Ping() error {
	return nil
}

func (c *MemoryCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		
		// Clean up token validations
		for key, entry := range c.tokenValidations {
			if now.After(entry.expiresAt) {
				delete(c.tokenValidations, key)
			}
		}
		
		// Clean up users
		for key, entry := range c.users {
			if now.After(entry.expiresAt) {
				delete(c.users, key)
			}
		}
		
		// Clean up blacklist
		for key, entry := range c.blacklist {
			if now.After(entry.expiresAt) {
				delete(c.blacklist, key)
			}
		}
	}
}