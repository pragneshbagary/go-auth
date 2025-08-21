package auth

import (
	"testing"
	"time"
)

func TestMetricsCollector(t *testing.T) {
	collector := NewMetricsCollector()

	t.Run("InitialState", func(t *testing.T) {
		metrics := collector.GetMetrics()
		
		if metrics.RegistrationAttempts != 0 {
			t.Errorf("Expected initial registration attempts to be 0, got %d", metrics.RegistrationAttempts)
		}
		if metrics.LoginAttempts != 0 {
			t.Errorf("Expected initial login attempts to be 0, got %d", metrics.LoginAttempts)
		}
		if metrics.TokenValidations != 0 {
			t.Errorf("Expected initial token validations to be 0, got %d", metrics.TokenValidations)
		}
	})

	t.Run("RecordRegistrationAttempt", func(t *testing.T) {
		// Record successful registration
		collector.RecordRegistrationAttempt(true)
		metrics := collector.GetMetrics()
		
		if metrics.RegistrationAttempts != 1 {
			t.Errorf("Expected registration attempts to be 1, got %d", metrics.RegistrationAttempts)
		}
		if metrics.RegistrationSuccess != 1 {
			t.Errorf("Expected registration success to be 1, got %d", metrics.RegistrationSuccess)
		}
		if metrics.UsersCreated != 1 {
			t.Errorf("Expected users created to be 1, got %d", metrics.UsersCreated)
		}

		// Record failed registration
		collector.RecordRegistrationAttempt(false)
		metrics = collector.GetMetrics()
		
		if metrics.RegistrationAttempts != 2 {
			t.Errorf("Expected registration attempts to be 2, got %d", metrics.RegistrationAttempts)
		}
		if metrics.RegistrationFailures != 1 {
			t.Errorf("Expected registration failures to be 1, got %d", metrics.RegistrationFailures)
		}
	})

	t.Run("RecordLoginAttempt", func(t *testing.T) {
		duration := time.Millisecond * 500
		
		// Record successful login
		collector.RecordLoginAttempt(true, duration)
		metrics := collector.GetMetrics()
		
		if metrics.LoginAttempts != 1 {
			t.Errorf("Expected login attempts to be 1, got %d", metrics.LoginAttempts)
		}
		if metrics.LoginSuccess != 1 {
			t.Errorf("Expected login success to be 1, got %d", metrics.LoginSuccess)
		}
		if metrics.TokensGenerated != 2 { // Access + Refresh token
			t.Errorf("Expected tokens generated to be 2, got %d", metrics.TokensGenerated)
		}
		if metrics.AverageLoginDuration != duration {
			t.Errorf("Expected average login duration to be %v, got %v", duration, metrics.AverageLoginDuration)
		}

		// Record failed login
		collector.RecordLoginAttempt(false, duration)
		metrics = collector.GetMetrics()
		
		if metrics.LoginFailures != 1 {
			t.Errorf("Expected login failures to be 1, got %d", metrics.LoginFailures)
		}
		if metrics.AuthenticationErrors != 1 {
			t.Errorf("Expected authentication errors to be 1, got %d", metrics.AuthenticationErrors)
		}
	})

	t.Run("RecordTokenRefresh", func(t *testing.T) {
		duration := time.Millisecond * 100
		
		// Record successful token refresh
		collector.RecordTokenRefresh(true, duration)
		metrics := collector.GetMetrics()
		
		if metrics.TokenRefreshes != 1 {
			t.Errorf("Expected token refreshes to be 1, got %d", metrics.TokenRefreshes)
		}
		if metrics.AverageRefreshDuration != duration {
			t.Errorf("Expected average refresh duration to be %v, got %v", duration, metrics.AverageRefreshDuration)
		}

		// Record failed token refresh
		collector.RecordTokenRefresh(false, duration)
		metrics = collector.GetMetrics()
		
		if metrics.AuthenticationErrors < 1 {
			t.Errorf("Expected authentication errors to be incremented")
		}
	})

	t.Run("RecordTokenValidation", func(t *testing.T) {
		duration := time.Microsecond * 50
		
		// Record successful validation
		collector.RecordTokenValidation(true, duration)
		metrics := collector.GetMetrics()
		
		if metrics.TokenValidations != 1 {
			t.Errorf("Expected token validations to be 1, got %d", metrics.TokenValidations)
		}
		if metrics.AverageValidationDuration != duration {
			t.Errorf("Expected average validation duration to be %v, got %v", duration, metrics.AverageValidationDuration)
		}

		// Record failed validation
		collector.RecordTokenValidation(false, duration)
		metrics = collector.GetMetrics()
		
		if metrics.TokenValidationFail != 1 {
			t.Errorf("Expected token validation failures to be 1, got %d", metrics.TokenValidationFail)
		}
	})

	t.Run("RecordTokenRevocation", func(t *testing.T) {
		collector.RecordTokenRevocation(true)
		metrics := collector.GetMetrics()
		
		if metrics.TokenRevocations != 1 {
			t.Errorf("Expected token revocations to be 1, got %d", metrics.TokenRevocations)
		}
	})

	t.Run("RecordPasswordOperations", func(t *testing.T) {
		collector.RecordPasswordChange(true)
		collector.RecordPasswordReset(true)
		metrics := collector.GetMetrics()
		
		if metrics.PasswordChanges != 1 {
			t.Errorf("Expected password changes to be 1, got %d", metrics.PasswordChanges)
		}
		if metrics.PasswordResets != 1 {
			t.Errorf("Expected password resets to be 1, got %d", metrics.PasswordResets)
		}
	})

	t.Run("RecordUserOperations", func(t *testing.T) {
		collector.RecordUserUpdate(true)
		collector.RecordUserDeletion(true)
		metrics := collector.GetMetrics()
		
		if metrics.UsersUpdated != 1 {
			t.Errorf("Expected users updated to be 1, got %d", metrics.UsersUpdated)
		}
		if metrics.UsersDeleted != 1 {
			t.Errorf("Expected users deleted to be 1, got %d", metrics.UsersDeleted)
		}
	})

	t.Run("RecordErrors", func(t *testing.T) {
		collector.RecordDatabaseError()
		collector.RecordValidationError()
		metrics := collector.GetMetrics()
		
		if metrics.DatabaseErrors != 1 {
			t.Errorf("Expected database errors to be 1, got %d", metrics.DatabaseErrors)
		}
		if metrics.ValidationErrors != 1 {
			t.Errorf("Expected validation errors to be 1, got %d", metrics.ValidationErrors)
		}
	})

	t.Run("GetUptime", func(t *testing.T) {
		uptime := collector.GetUptime()
		if uptime <= 0 {
			t.Errorf("Expected uptime to be positive, got %v", uptime)
		}
	})

	t.Run("GetSuccessRates", func(t *testing.T) {
		// Reset collector for clean test
		collector.Reset()
		
		// Record some operations
		collector.RecordLoginAttempt(true, time.Millisecond)
		collector.RecordLoginAttempt(true, time.Millisecond)
		collector.RecordLoginAttempt(false, time.Millisecond)
		
		collector.RecordRegistrationAttempt(true)
		collector.RecordRegistrationAttempt(false)
		
		collector.RecordTokenValidation(true, time.Microsecond)
		collector.RecordTokenValidation(true, time.Microsecond)
		collector.RecordTokenValidation(false, time.Microsecond)

		loginSuccessRate := collector.GetLoginSuccessRate()
		expectedLoginRate := float64(2) / float64(3) * 100.0
		if loginSuccessRate != expectedLoginRate {
			t.Errorf("Expected login success rate to be %.2f%%, got %.2f%%", expectedLoginRate, loginSuccessRate)
		}

		registrationSuccessRate := collector.GetRegistrationSuccessRate()
		expectedRegistrationRate := float64(1) / float64(2) * 100.0
		if registrationSuccessRate != expectedRegistrationRate {
			t.Errorf("Expected registration success rate to be %.2f%%, got %.2f%%", expectedRegistrationRate, registrationSuccessRate)
		}

		tokenValidationSuccessRate := collector.GetTokenValidationSuccessRate()
		expectedTokenRate := float64(2) / float64(3) * 100.0
		if tokenValidationSuccessRate != expectedTokenRate {
			t.Errorf("Expected token validation success rate to be %.2f%%, got %.2f%%", expectedTokenRate, tokenValidationSuccessRate)
		}
	})

	t.Run("Reset", func(t *testing.T) {
		// Record some data
		collector.RecordLoginAttempt(true, time.Millisecond)
		collector.RecordRegistrationAttempt(true)
		
		// Reset
		collector.Reset()
		metrics := collector.GetMetrics()
		
		if metrics.LoginAttempts != 0 {
			t.Errorf("Expected login attempts to be 0 after reset, got %d", metrics.LoginAttempts)
		}
		if metrics.RegistrationAttempts != 0 {
			t.Errorf("Expected registration attempts to be 0 after reset, got %d", metrics.RegistrationAttempts)
		}
	})
}