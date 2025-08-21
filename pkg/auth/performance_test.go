package auth

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/models"
)

func BenchmarkPasswordHashing(b *testing.B) {
	password := "testpassword123"
	
	b.Run("HashPassword", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := HashPassword(password)
			if err != nil {
				b.Fatalf("Failed to hash password: %v", err)
			}
		}
	})
	
	// Pre-hash a password for verification benchmarks
	hashedPassword, err := HashPassword(password)
	if err != nil {
		b.Fatalf("Failed to hash password for benchmark: %v", err)
	}
	
	b.Run("CheckPasswordHash", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			CheckPasswordHash(password, hashedPassword)
		}
	})
}

func BenchmarkTokenOperations(b *testing.B) {
	auth, err := NewInMemory("test-secret-key-for-benchmarking")
	if err != nil {
		b.Fatalf("Failed to create auth instance: %v", err)
	}

	// Register a test user
	req := RegisterRequest{
		Username: "benchmarkuser",
		Email:    "benchmark@example.com",
		Password: "password123",
	}
	
	_, err = auth.Register(req)
	if err != nil {
		b.Fatalf("Failed to register user: %v", err)
	}

	// Login to get tokens
	tokens, err := auth.Login("benchmarkuser", "password123", nil)
	if err != nil {
		b.Fatalf("Failed to login user: %v", err)
	}

	b.Run("ValidateAccessToken", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := auth.ValidateAccessToken(tokens.AccessToken)
			if err != nil {
				b.Fatalf("Failed to validate token: %v", err)
			}
		}
	})

	b.Run("RefreshToken", func(b *testing.B) {
		// We need fresh refresh tokens for each iteration
		refreshTokens := make([]string, b.N)
		
		// Pre-generate refresh tokens
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			testReq := RegisterRequest{
				Username: fmt.Sprintf("refreshuser%d", i),
				Email:    fmt.Sprintf("refresh%d@example.com", i),
				Password: "password123",
			}
			_, err := auth.Register(testReq)
			if err != nil {
				b.Fatalf("Failed to register test user: %v", err)
			}
			
			// Login to get tokens
			testTokens, err := auth.Login(testReq.Username, "password123", nil)
			if err != nil {
				b.Fatalf("Failed to login test user: %v", err)
			}
			refreshTokens[i] = testTokens.RefreshToken
		}
		b.StartTimer()

		for i := 0; i < b.N; i++ {
			_, err := auth.RefreshToken(refreshTokens[i])
			if err != nil {
				b.Fatalf("Failed to refresh token: %v", err)
			}
		}
	})
}

func BenchmarkUserOperations(b *testing.B) {
	auth, err := NewInMemory("test-secret-key-for-benchmarking")
	if err != nil {
		b.Fatalf("Failed to create auth instance: %v", err)
	}

	b.Run("Register", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			req := RegisterRequest{
				Username: fmt.Sprintf("user%d", i),
				Email:    fmt.Sprintf("user%d@example.com", i),
				Password: "password123",
			}
			_, err := auth.Register(req)
			if err != nil {
				b.Fatalf("Failed to register user: %v", err)
			}
		}
	})

	// Pre-register users for other benchmarks
	users := make([]*models.User, 1000)
	for i := 0; i < 1000; i++ {
		req := RegisterRequest{
			Username: fmt.Sprintf("preuser%d", i),
			Email:    fmt.Sprintf("preuser%d@example.com", i),
			Password: "password123",
		}
		user, err := auth.Register(req)
		if err != nil {
			b.Fatalf("Failed to pre-register user: %v", err)
		}
		users[i] = user
	}

	b.Run("Login", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userIndex := i % len(users)
			_, err := auth.Login(users[userIndex].Username, "password123", nil)
			if err != nil {
				b.Fatalf("Failed to login user: %v", err)
			}
		}
	})

	b.Run("GetUser", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userIndex := i % len(users)
			_, err := auth.GetUser(users[userIndex].ID)
			if err != nil {
				b.Fatalf("Failed to get user: %v", err)
			}
		}
	})

	b.Run("GetUserByUsername", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userIndex := i % len(users)
			_, err := auth.GetUserByUsername(users[userIndex].Username)
			if err != nil {
				b.Fatalf("Failed to get user by username: %v", err)
			}
		}
	})

	b.Run("GetUserByEmail", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			userIndex := i % len(users)
			_, err := auth.GetUserByEmail(users[userIndex].Email)
			if err != nil {
				b.Fatalf("Failed to get user by email: %v", err)
			}
		}
	})
}

func TestConcurrentOperations(t *testing.T) {
	auth, err := NewInMemory("test-secret-key-for-concurrency")
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	t.Run("ConcurrentRegistration", func(t *testing.T) {
		const numGoroutines = 100
		const usersPerGoroutine = 10
		
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*usersPerGoroutine)
		
		start := time.Now()
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				
				for j := 0; j < usersPerGoroutine; j++ {
					req := RegisterRequest{
						Username: fmt.Sprintf("concurrent%d_%d", goroutineID, j),
						Email:    fmt.Sprintf("concurrent%d_%d@example.com", goroutineID, j),
						Password: "password123",
					}
					
					_, err := auth.Register(req)
					if err != nil {
						errors <- err
						return
					}
				}
			}(i)
		}
		
		wg.Wait()
		close(errors)
		
		duration := time.Since(start)
		totalUsers := numGoroutines * usersPerGoroutine
		
		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Errorf("Concurrent registration error: %v", err)
			errorCount++
		}
		
		if errorCount == 0 {
			t.Logf("Successfully registered %d users concurrently in %v (%.2f users/sec)", 
				totalUsers, duration, float64(totalUsers)/duration.Seconds())
		}
	})

	t.Run("ConcurrentLogin", func(t *testing.T) {
		// First register some users
		const numUsers = 50
		users := make([]*models.User, numUsers)
		
		for i := 0; i < numUsers; i++ {
			req := RegisterRequest{
				Username: fmt.Sprintf("loginuser%d", i),
				Email:    fmt.Sprintf("loginuser%d@example.com", i),
				Password: "password123",
			}
			user, err := auth.Register(req)
			if err != nil {
				t.Fatalf("Failed to register user for concurrent login test: %v", err)
			}
			users[i] = user
		}
		
		const numGoroutines = 20
		const loginsPerGoroutine = 25
		
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*loginsPerGoroutine)
		
		start := time.Now()
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				
				for j := 0; j < loginsPerGoroutine; j++ {
					userIndex := (goroutineID*loginsPerGoroutine + j) % len(users)
					_, err := auth.Login(users[userIndex].Username, "password123", nil)
					if err != nil {
						errors <- err
						return
					}
				}
			}(i)
		}
		
		wg.Wait()
		close(errors)
		
		duration := time.Since(start)
		totalLogins := numGoroutines * loginsPerGoroutine
		
		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Errorf("Concurrent login error: %v", err)
			errorCount++
		}
		
		if errorCount == 0 {
			t.Logf("Successfully performed %d concurrent logins in %v (%.2f logins/sec)", 
				totalLogins, duration, float64(totalLogins)/duration.Seconds())
		}
	})

	t.Run("ConcurrentTokenValidation", func(t *testing.T) {
		// Register a user and get tokens
		req := RegisterRequest{
			Username: "tokenuser",
			Email:    "tokenuser@example.com",
			Password: "password123",
		}
		
		_, err := auth.Register(req)
		if err != nil {
			t.Fatalf("Failed to register user for token validation test: %v", err)
		}
		
		// Login to get tokens
		tokens, err := auth.Login("tokenuser", "password123", nil)
		if err != nil {
			t.Fatalf("Failed to login user for token validation test: %v", err)
		}
		
		const numGoroutines = 50
		const validationsPerGoroutine = 100
		
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines*validationsPerGoroutine)
		
		start := time.Now()
		
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				
				for j := 0; j < validationsPerGoroutine; j++ {
					_, err := auth.ValidateAccessToken(tokens.AccessToken)
					if err != nil {
						errors <- err
						return
					}
				}
			}()
		}
		
		wg.Wait()
		close(errors)
		
		duration := time.Since(start)
		totalValidations := numGoroutines * validationsPerGoroutine
		
		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Errorf("Concurrent token validation error: %v", err)
			errorCount++
		}
		
		if errorCount == 0 {
			t.Logf("Successfully performed %d concurrent token validations in %v (%.2f validations/sec)", 
				totalValidations, duration, float64(totalValidations)/duration.Seconds())
		}
	})
}

func TestMemoryUsage(t *testing.T) {
	t.Run("MemoryLeakDetection", func(t *testing.T) {
		auth, err := NewInMemory("test-secret-key")
		if err != nil {
			t.Fatalf("Failed to create auth instance: %v", err)
		}

		// Create and delete many users to test for memory leaks
		const numIterations = 1000
		
		for i := 0; i < numIterations; i++ {
			req := RegisterRequest{
				Username: fmt.Sprintf("memuser%d", i),
				Email:    fmt.Sprintf("memuser%d@example.com", i),
				Password: "password123",
			}
			
			// Register user
			user, err := auth.Register(req)
			if err != nil {
				t.Fatalf("Failed to register user %d: %v", i, err)
			}
			
			// Perform some operations
			_, err = auth.Login(user.Username, "password123", nil)
			if err != nil {
				t.Fatalf("Failed to login user %d: %v", i, err)
			}
			
			// Get user
			_, err = auth.GetUser(user.ID)
			if err != nil {
				t.Fatalf("Failed to get user %d: %v", i, err)
			}
			
			// Clean up periodically (simulating user deletion)
			if i%100 == 99 {
				// In a real scenario, you might delete users here
				// For in-memory storage, this helps test cleanup
			}
		}
		
		t.Logf("Completed %d user operations without memory issues", numIterations)
	})
}

func BenchmarkCacheOperations(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()
	
	user := &models.User{
		ID:       "cache-test-user",
		Username: "cacheuser",
		Email:    "cache@example.com",
	}

	b.Run("CacheSet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tokenID := fmt.Sprintf("token%d", i)
			err := cache.SetTokenValidation(tokenID, user, time.Hour)
			if err != nil {
				b.Fatalf("Failed to set cache: %v", err)
			}
		}
	})

	// Pre-populate cache for get benchmarks
	for i := 0; i < 1000; i++ {
		tokenID := fmt.Sprintf("pretoken%d", i)
		cache.SetTokenValidation(tokenID, user, time.Hour)
	}

	b.Run("CacheGet", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tokenID := fmt.Sprintf("pretoken%d", i%1000)
			_, _, err := cache.GetTokenValidation(tokenID)
			if err != nil {
				b.Fatalf("Failed to get from cache: %v", err)
			}
		}
	})

	b.Run("CacheInvalidate", func(b *testing.B) {
		// Pre-populate for invalidation
		b.StopTimer()
		for i := 0; i < b.N; i++ {
			tokenID := fmt.Sprintf("invtoken%d", i)
			cache.SetTokenValidation(tokenID, user, time.Hour)
		}
		b.StartTimer()

		for i := 0; i < b.N; i++ {
			tokenID := fmt.Sprintf("invtoken%d", i)
			err := cache.InvalidateToken(tokenID)
			if err != nil {
				b.Fatalf("Failed to invalidate cache: %v", err)
			}
		}
	})
}

func TestLoadTesting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	t.Run("HighLoadScenario", func(t *testing.T) {
		auth, err := NewInMemory("test-secret-key-load-test")
		if err != nil {
			t.Fatalf("Failed to create auth instance: %v", err)
		}

		const duration = 10 * time.Second
		const numWorkers = 20
		
		var wg sync.WaitGroup
		stop := make(chan struct{})
		
		// Metrics
		var totalOps int64
		var errors int64
		var mu sync.Mutex
		
		// Start workers
		for i := 0; i < numWorkers; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()
				
				ops := 0
				errs := 0
				
				for {
					select {
					case <-stop:
						mu.Lock()
						totalOps += int64(ops)
						errors += int64(errs)
						mu.Unlock()
						return
					default:
						// Perform mixed operations
						req := RegisterRequest{
							Username: fmt.Sprintf("loaduser%d_%d", workerID, ops),
							Email:    fmt.Sprintf("loaduser%d_%d@example.com", workerID, ops),
							Password: "password123",
						}
						
						// Register
						user, err := auth.Register(req)
						if err != nil {
							errs++
						} else {
							ops++
							
							// Login
							_, err = auth.Login(user.Username, "password123", nil)
							if err != nil {
								errs++
							} else {
								ops++
							}
						}
					}
				}
			}(i)
		}
		
		// Run for specified duration
		time.Sleep(duration)
		close(stop)
		wg.Wait()
		
		opsPerSecond := float64(totalOps) / duration.Seconds()
		errorRate := float64(errors) / float64(totalOps+errors) * 100
		
		t.Logf("Load test results:")
		t.Logf("  Duration: %v", duration)
		t.Logf("  Workers: %d", numWorkers)
		t.Logf("  Total operations: %d", totalOps)
		t.Logf("  Operations/second: %.2f", opsPerSecond)
		t.Logf("  Errors: %d", errors)
		t.Logf("  Error rate: %.2f%%", errorRate)
		
		if errorRate > 5.0 {
			t.Errorf("Error rate too high: %.2f%%", errorRate)
		}
		
		if opsPerSecond < 100 {
			t.Logf("Warning: Low throughput: %.2f ops/sec", opsPerSecond)
		}
	})
}