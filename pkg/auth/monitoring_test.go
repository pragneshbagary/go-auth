package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pragneshbagary/go-auth/internal/storage/memory"
)

func TestMonitor(t *testing.T) {
	storage := memory.NewInMemoryStorage()
	metricsCollector := NewMetricsCollector()
	logger := NewDefaultLogger()
	monitor := NewMonitor(storage, metricsCollector, logger, "test-app", "1.0.0")

	t.Run("CheckHealth", func(t *testing.T) {
		health := monitor.CheckHealth()
		
		if health.Status != HealthStatusHealthy {
			t.Errorf("Expected health status to be healthy, got %s", health.Status)
		}
		
		if len(health.Components) == 0 {
			t.Error("Expected health components to be present")
		}
		
		// Check that database component is present
		var dbComponent *ComponentHealth
		for _, component := range health.Components {
			if component.Name == "database" {
				dbComponent = &component
				break
			}
		}
		
		if dbComponent == nil {
			t.Error("Expected database component in health check")
		} else if dbComponent.Status != HealthStatusHealthy {
			t.Errorf("Expected database component to be healthy, got %s", dbComponent.Status)
		}
	})

	t.Run("GetSystemInfo", func(t *testing.T) {
		info := monitor.GetSystemInfo()
		
		if info.Name != "test-app" {
			t.Errorf("Expected app name to be 'test-app', got %s", info.Name)
		}
		
		if info.Version != "1.0.0" {
			t.Errorf("Expected version to be '1.0.0', got %s", info.Version)
		}
		
		if info.GoVersion == "" {
			t.Error("Expected Go version to be present")
		}
		
		if info.NumCPU <= 0 {
			t.Error("Expected number of CPUs to be positive")
		}
		
		if info.DatabaseType != "memory" {
			t.Errorf("Expected database type to be 'memory', got %s", info.DatabaseType)
		}
		
		if info.DatabaseStatus != "connected" {
			t.Errorf("Expected database status to be 'connected', got %s", info.DatabaseStatus)
		}
	})

	t.Run("HTTPHealthHandler", func(t *testing.T) {
		handler := monitor.HTTPHealthHandler()
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
		
		var health SystemHealth
		if err := json.Unmarshal(w.Body.Bytes(), &health); err != nil {
			t.Fatalf("Failed to parse health response: %v", err)
		}
		
		if health.Status != HealthStatusHealthy {
			t.Errorf("Expected health status to be healthy, got %s", health.Status)
		}
	})

	t.Run("HTTPMetricsHandler", func(t *testing.T) {
		// Record some metrics first
		metricsCollector.RecordLoginAttempt(true, 0)
		metricsCollector.RecordRegistrationAttempt(true)
		
		handler := monitor.HTTPMetricsHandler()
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
		
		var metrics Metrics
		if err := json.Unmarshal(w.Body.Bytes(), &metrics); err != nil {
			t.Fatalf("Failed to parse metrics response: %v", err)
		}
		
		if metrics.LoginAttempts != 1 {
			t.Errorf("Expected login attempts to be 1, got %d", metrics.LoginAttempts)
		}
		
		if metrics.RegistrationAttempts != 1 {
			t.Errorf("Expected registration attempts to be 1, got %d", metrics.RegistrationAttempts)
		}
	})

	t.Run("HTTPSystemInfoHandler", func(t *testing.T) {
		handler := monitor.HTTPSystemInfoHandler()
		req := httptest.NewRequest("GET", "/info", nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
		
		var info SystemInfo
		if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
			t.Fatalf("Failed to parse system info response: %v", err)
		}
		
		if info.Name != "test-app" {
			t.Errorf("Expected app name to be 'test-app', got %s", info.Name)
		}
	})

	t.Run("HTTPReadinessHandler", func(t *testing.T) {
		handler := monitor.HTTPReadinessHandler()
		req := httptest.NewRequest("GET", "/health/ready", nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
		
		if w.Body.String() != "Ready" {
			t.Errorf("Expected body to be 'Ready', got %s", w.Body.String())
		}
	})

	t.Run("HTTPLivenessHandler", func(t *testing.T) {
		handler := monitor.HTTPLivenessHandler()
		req := httptest.NewRequest("GET", "/health/live", nil)
		w := httptest.NewRecorder()
		
		handler(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}
		
		if w.Body.String() != "Alive" {
			t.Errorf("Expected body to be 'Alive', got %s", w.Body.String())
		}
	})

	t.Run("RegisterHTTPHandlers", func(t *testing.T) {
		mux := http.NewServeMux()
		monitor.RegisterHTTPHandlers(mux)
		
		// Test that handlers are registered by making requests
		testCases := []struct {
			path           string
			expectedStatus int
		}{
			{"/health", http.StatusOK},
			{"/health/ready", http.StatusOK},
			{"/health/live", http.StatusOK},
			{"/metrics", http.StatusOK},
			{"/info", http.StatusOK},
		}
		
		for _, tc := range testCases {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			
			mux.ServeHTTP(w, req)
			
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d for %s, got %d", tc.expectedStatus, tc.path, w.Code)
			}
		}
	})
}

func TestHealthStatus(t *testing.T) {
	tests := []struct {
		status   HealthStatus
		expected string
	}{
		{HealthStatusHealthy, "healthy"},
		{HealthStatusUnhealthy, "unhealthy"},
		{HealthStatusDegraded, "degraded"},
	}

	for _, test := range tests {
		if string(test.status) != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, string(test.status))
		}
	}
}