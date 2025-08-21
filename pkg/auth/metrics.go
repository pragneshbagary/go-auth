package auth

import (
	"sync"
	"time"
)

// Metrics holds authentication operation metrics
type Metrics struct {
	mu sync.RWMutex

	// Registration metrics
	RegistrationAttempts int64 `json:"registration_attempts"`
	RegistrationSuccess  int64 `json:"registration_success"`
	RegistrationFailures int64 `json:"registration_failures"`

	// Login metrics
	LoginAttempts int64 `json:"login_attempts"`
	LoginSuccess  int64 `json:"login_success"`
	LoginFailures int64 `json:"login_failures"`

	// Token metrics
	TokensGenerated     int64 `json:"tokens_generated"`
	TokenRefreshes      int64 `json:"token_refreshes"`
	TokenValidations    int64 `json:"token_validations"`
	TokenRevocations    int64 `json:"token_revocations"`
	TokenValidationFail int64 `json:"token_validation_failures"`

	// Password metrics
	PasswordChanges int64 `json:"password_changes"`
	PasswordResets  int64 `json:"password_resets"`

	// User management metrics
	UsersCreated int64 `json:"users_created"`
	UsersUpdated int64 `json:"users_updated"`
	UsersDeleted int64 `json:"users_deleted"`

	// Performance metrics
	AverageLoginDuration    time.Duration `json:"average_login_duration"`
	AverageRefreshDuration  time.Duration `json:"average_refresh_duration"`
	AverageValidationDuration time.Duration `json:"average_validation_duration"`

	// Timing accumulators (for calculating averages)
	totalLoginDuration      time.Duration
	totalRefreshDuration    time.Duration
	totalValidationDuration time.Duration

	// Error metrics
	DatabaseErrors    int64 `json:"database_errors"`
	ValidationErrors  int64 `json:"validation_errors"`
	AuthenticationErrors int64 `json:"authentication_errors"`

	// System metrics
	StartTime    time.Time `json:"start_time"`
	LastActivity time.Time `json:"last_activity"`
}

// MetricsCollector provides thread-safe metrics collection
type MetricsCollector struct {
	metrics *Metrics
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		metrics: &Metrics{
			StartTime:    time.Now(),
			LastActivity: time.Now(),
		},
	}
}

// GetMetrics returns a copy of the current metrics
func (mc *MetricsCollector) GetMetrics() Metrics {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	// Create a copy to avoid race conditions
	metricsCopy := *mc.metrics
	return metricsCopy
}

// RecordRegistrationAttempt records a registration attempt
func (mc *MetricsCollector) RecordRegistrationAttempt(success bool) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.RegistrationAttempts++
	mc.metrics.LastActivity = time.Now()

	if success {
		mc.metrics.RegistrationSuccess++
		mc.metrics.UsersCreated++
	} else {
		mc.metrics.RegistrationFailures++
	}
}

// RecordLoginAttempt records a login attempt with duration
func (mc *MetricsCollector) RecordLoginAttempt(success bool, duration time.Duration) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.LoginAttempts++
	mc.metrics.LastActivity = time.Now()

	if success {
		mc.metrics.LoginSuccess++
		mc.metrics.TokensGenerated += 2 // Access + Refresh token
		mc.metrics.totalLoginDuration += duration
		mc.metrics.AverageLoginDuration = mc.metrics.totalLoginDuration / time.Duration(mc.metrics.LoginSuccess)
	} else {
		mc.metrics.LoginFailures++
		mc.metrics.AuthenticationErrors++
	}
}

// RecordTokenRefresh records a token refresh operation
func (mc *MetricsCollector) RecordTokenRefresh(success bool, duration time.Duration) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.LastActivity = time.Now()

	if success {
		mc.metrics.TokenRefreshes++
		mc.metrics.TokensGenerated += 2 // New Access + Refresh token
		mc.metrics.totalRefreshDuration += duration
		mc.metrics.AverageRefreshDuration = mc.metrics.totalRefreshDuration / time.Duration(mc.metrics.TokenRefreshes)
	} else {
		mc.metrics.AuthenticationErrors++
	}
}

// RecordTokenValidation records a token validation operation
func (mc *MetricsCollector) RecordTokenValidation(success bool, duration time.Duration) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.TokenValidations++
	mc.metrics.LastActivity = time.Now()

	if success {
		mc.metrics.totalValidationDuration += duration
		mc.metrics.AverageValidationDuration = mc.metrics.totalValidationDuration / time.Duration(mc.metrics.TokenValidations)
	} else {
		mc.metrics.TokenValidationFail++
		mc.metrics.AuthenticationErrors++
	}
}

// RecordTokenRevocation records a token revocation
func (mc *MetricsCollector) RecordTokenRevocation(success bool) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.LastActivity = time.Now()

	if success {
		mc.metrics.TokenRevocations++
	} else {
		mc.metrics.AuthenticationErrors++
	}
}

// RecordPasswordChange records a password change operation
func (mc *MetricsCollector) RecordPasswordChange(success bool) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.LastActivity = time.Now()

	if success {
		mc.metrics.PasswordChanges++
	} else {
		mc.metrics.ValidationErrors++
	}
}

// RecordPasswordReset records a password reset operation
func (mc *MetricsCollector) RecordPasswordReset(success bool) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.LastActivity = time.Now()

	if success {
		mc.metrics.PasswordResets++
	} else {
		mc.metrics.ValidationErrors++
	}
}

// RecordUserUpdate records a user update operation
func (mc *MetricsCollector) RecordUserUpdate(success bool) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.LastActivity = time.Now()

	if success {
		mc.metrics.UsersUpdated++
	} else {
		mc.metrics.ValidationErrors++
	}
}

// RecordUserDeletion records a user deletion operation
func (mc *MetricsCollector) RecordUserDeletion(success bool) {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.LastActivity = time.Now()

	if success {
		mc.metrics.UsersDeleted++
	} else {
		mc.metrics.DatabaseErrors++
	}
}

// RecordDatabaseError records a database error
func (mc *MetricsCollector) RecordDatabaseError() {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.DatabaseErrors++
	mc.metrics.LastActivity = time.Now()
}

// RecordValidationError records a validation error
func (mc *MetricsCollector) RecordValidationError() {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	mc.metrics.ValidationErrors++
	mc.metrics.LastActivity = time.Now()
}

// Reset resets all metrics to zero
func (mc *MetricsCollector) Reset() {
	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()

	now := time.Now()
	mc.metrics = &Metrics{
		StartTime:    now,
		LastActivity: now,
	}
}

// GetUptime returns the uptime since metrics collection started
func (mc *MetricsCollector) GetUptime() time.Duration {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	return time.Since(mc.metrics.StartTime)
}

// GetTimeSinceLastActivity returns the time since last recorded activity
func (mc *MetricsCollector) GetTimeSinceLastActivity() time.Duration {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	return time.Since(mc.metrics.LastActivity)
}

// GetSuccessRate returns the overall success rate as a percentage
func (mc *MetricsCollector) GetSuccessRate() float64 {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	totalAttempts := mc.metrics.LoginAttempts + mc.metrics.RegistrationAttempts + mc.metrics.TokenRefreshes
	totalSuccess := mc.metrics.LoginSuccess + mc.metrics.RegistrationSuccess + mc.metrics.TokenRefreshes

	if totalAttempts == 0 {
		return 0.0
	}

	return float64(totalSuccess) / float64(totalAttempts) * 100.0
}

// GetLoginSuccessRate returns the login success rate as a percentage
func (mc *MetricsCollector) GetLoginSuccessRate() float64 {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	if mc.metrics.LoginAttempts == 0 {
		return 0.0
	}

	return float64(mc.metrics.LoginSuccess) / float64(mc.metrics.LoginAttempts) * 100.0
}

// GetRegistrationSuccessRate returns the registration success rate as a percentage
func (mc *MetricsCollector) GetRegistrationSuccessRate() float64 {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	if mc.metrics.RegistrationAttempts == 0 {
		return 0.0
	}

	return float64(mc.metrics.RegistrationSuccess) / float64(mc.metrics.RegistrationAttempts) * 100.0
}

// GetTokenValidationSuccessRate returns the token validation success rate as a percentage
func (mc *MetricsCollector) GetTokenValidationSuccessRate() float64 {
	mc.metrics.mu.RLock()
	defer mc.metrics.mu.RUnlock()

	if mc.metrics.TokenValidations == 0 {
		return 0.0
	}

	successfulValidations := mc.metrics.TokenValidations - mc.metrics.TokenValidationFail
	return float64(successfulValidations) / float64(mc.metrics.TokenValidations) * 100.0
}