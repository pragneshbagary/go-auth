package auth

import (
	"testing"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
)

func TestNoOpCache(t *testing.T) {
	cache := NewNoOpCache()

	// Test token validation operations
	user := &models.User{ID: "test-user", Username: "testuser"}
	
	t.Run("SetTokenValidation", func(t *testing.T) {
		err := cache.SetTokenValidation("token1", user, time.Hour)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("GetTokenValidation", func(t *testing.T) {
		retrievedUser, found, err := cache.GetTokenValidation("token1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if found {
			t.Error("Expected not found for NoOpCache")
		}
		if retrievedUser != nil {
			t.Error("Expected nil user for NoOpCache")
		}
	})

	t.Run("InvalidateToken", func(t *testing.T) {
		err := cache.InvalidateToken("token1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Test user operations
	t.Run("SetUser", func(t *testing.T) {
		err := cache.SetUser("user1", user, time.Hour)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("GetUser", func(t *testing.T) {
		retrievedUser, found, err := cache.GetUser("user1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if found {
			t.Error("Expected not found for NoOpCache")
		}
		if retrievedUser != nil {
			t.Error("Expected nil user for NoOpCache")
		}
	})

	t.Run("InvalidateUser", func(t *testing.T) {
		err := cache.InvalidateUser("user1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Test blacklist operations
	t.Run("SetTokenBlacklist", func(t *testing.T) {
		err := cache.SetTokenBlacklist("token1", time.Hour)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("IsTokenBlacklisted", func(t *testing.T) {
		blacklisted, found, err := cache.IsTokenBlacklisted("token1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if found {
			t.Error("Expected not found for NoOpCache")
		}
		if blacklisted {
			t.Error("Expected not blacklisted for NoOpCache")
		}
	})

	t.Run("InvalidateTokenBlacklist", func(t *testing.T) {
		err := cache.InvalidateTokenBlacklist("token1")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	// Test connection operations
	t.Run("Close", func(t *testing.T) {
		err := cache.Close()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Ping", func(t *testing.T) {
		err := cache.Ping()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}

func TestMemoryCache(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	user := &models.User{
		ID:       "test-user",
		Username: "testuser",
		Email:    "test@example.com",
	}

	t.Run("TokenValidation", func(t *testing.T) {
		// Set token validation
		err := cache.SetTokenValidation("token1", user, time.Minute)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Get token validation
		retrievedUser, found, err := cache.GetTokenValidation("token1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !found {
			t.Fatal("Expected token to be found")
		}
		if retrievedUser.ID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, retrievedUser.ID)
		}

		// Invalidate token
		err = cache.InvalidateToken("token1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify token is invalidated
		_, found, err = cache.GetTokenValidation("token1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if found {
			t.Error("Expected token to be invalidated")
		}
	})

	t.Run("UserCaching", func(t *testing.T) {
		// Set user
		err := cache.SetUser("user1", user, time.Minute)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Get user
		retrievedUser, found, err := cache.GetUser("user1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !found {
			t.Fatal("Expected user to be found")
		}
		if retrievedUser.ID != user.ID {
			t.Errorf("Expected user ID %s, got %s", user.ID, retrievedUser.ID)
		}

		// Invalidate user
		err = cache.InvalidateUser("user1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify user is invalidated
		_, found, err = cache.GetUser("user1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if found {
			t.Error("Expected user to be invalidated")
		}
	})

	t.Run("TokenBlacklist", func(t *testing.T) {
		// Set token blacklist
		err := cache.SetTokenBlacklist("token1", time.Minute)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Check if token is blacklisted
		blacklisted, found, err := cache.IsTokenBlacklisted("token1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !found {
			t.Fatal("Expected blacklist entry to be found")
		}
		if !blacklisted {
			t.Error("Expected token to be blacklisted")
		}

		// Invalidate blacklist entry
		err = cache.InvalidateTokenBlacklist("token1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify blacklist entry is invalidated
		_, found, err = cache.IsTokenBlacklisted("token1")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if found {
			t.Error("Expected blacklist entry to be invalidated")
		}
	})

	t.Run("Expiration", func(t *testing.T) {
		// Set token with very short TTL
		err := cache.SetTokenValidation("short-token", user, time.Millisecond)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Verify token is expired
		_, found, err := cache.GetTokenValidation("short-token")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if found {
			t.Error("Expected token to be expired")
		}
	})

	t.Run("InvalidCacheEntry", func(t *testing.T) {
		// Manually set invalid cache entry
		cache.tokenValidations["invalid"] = &cacheEntry{
			data:      "invalid-data",
			expiresAt: time.Now().Add(time.Hour),
		}

		// Try to get invalid entry
		_, _, err := cache.GetTokenValidation("invalid")
		if err == nil {
			t.Error("Expected error for invalid cache entry")
		}
	})

	t.Run("Ping", func(t *testing.T) {
		err := cache.Ping()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})
}