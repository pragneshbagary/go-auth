package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/pragneshbagary/go-auth/pkg/storage"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusDegraded  HealthStatus = "degraded"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name      string       `json:"name"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	Details   interface{}  `json:"details,omitempty"`
	CheckedAt time.Time    `json:"checked_at"`
	Duration  string       `json:"duration"`
}

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status     HealthStatus      `json:"status"`
	Timestamp  time.Time         `json:"timestamp"`
	Uptime     string            `json:"uptime"`
	Version    string            `json:"version"`
	Components []ComponentHealth `json:"components"`
}

// SystemInfo provides detailed system information
type SystemInfo struct {
	// Application info
	Name      string    `json:"name"`
	Version   string    `json:"version"`
	StartTime time.Time `json:"start_time"`
	Uptime    string    `json:"uptime"`

	// Runtime info
	GoVersion    string `json:"go_version"`
	NumGoroutine int    `json:"num_goroutine"`
	NumCPU       int    `json:"num_cpu"`

	// Memory info
	MemoryAlloc      uint64 `json:"memory_alloc"`
	MemoryTotalAlloc uint64 `json:"memory_total_alloc"`
	MemorySys        uint64 `json:"memory_sys"`
	MemoryNumGC      uint32 `json:"memory_num_gc"`

	// Database info
	DatabaseType   string `json:"database_type"`
	DatabaseStatus string `json:"database_status"`

	// Metrics
	Metrics Metrics `json:"metrics"`
}

// Monitor provides health checking and monitoring capabilities
type Monitor struct {
	storage         storage.EnhancedStorage
	metricsCollector *MetricsCollector
	logger          *Logger
	startTime       time.Time
	appName         string
	version         string
}

// NewMonitor creates a new monitor instance
func NewMonitor(storage storage.EnhancedStorage, metricsCollector *MetricsCollector, logger *Logger, appName, version string) *Monitor {
	return &Monitor{
		storage:         storage,
		metricsCollector: metricsCollector,
		logger:          logger,
		startTime:       time.Now(),
		appName:         appName,
		version:         version,
	}
}

// CheckHealth performs a comprehensive health check
func (m *Monitor) CheckHealth() SystemHealth {
	components := []ComponentHealth{}

	// Check database health
	dbHealth := m.checkDatabaseHealth()
	components = append(components, dbHealth)

	// Check metrics collector health
	metricsHealth := m.checkMetricsHealth()
	components = append(components, metricsHealth)

	// Check logger health
	loggerHealth := m.checkLoggerHealth()
	components = append(components, loggerHealth)

	// Determine overall status
	overallStatus := HealthStatusHealthy
	for _, component := range components {
		if component.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
			break
		} else if component.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
		}
	}

	return SystemHealth{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Uptime:     time.Since(m.startTime).String(),
		Version:    m.version,
		Components: components,
	}
}

// checkDatabaseHealth checks the health of the database connection
func (m *Monitor) checkDatabaseHealth() ComponentHealth {
	start := time.Now()
	
	health := ComponentHealth{
		Name:      "database",
		CheckedAt: time.Now(),
	}

	if err := m.storage.Ping(); err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = "Database connection failed"
		health.Details = map[string]interface{}{
			"error": err.Error(),
		}
		m.logger.Error("Database health check failed", map[string]interface{}{
			"error": err,
			"duration": time.Since(start),
		})
	} else {
		health.Status = HealthStatusHealthy
		health.Message = "Database connection is healthy"
		
		// Additional database checks
		details := map[string]interface{}{}
		
		// Check schema version
		if version, err := m.storage.GetSchemaVersion(); err == nil {
			details["schema_version"] = version
		}
		
		health.Details = details
	}

	health.Duration = time.Since(start).String()
	return health
}

// checkMetricsHealth checks the health of the metrics collector
func (m *Monitor) checkMetricsHealth() ComponentHealth {
	start := time.Now()
	
	health := ComponentHealth{
		Name:      "metrics",
		Status:    HealthStatusHealthy,
		Message:   "Metrics collector is operational",
		CheckedAt: time.Now(),
	}

	if m.metricsCollector == nil {
		health.Status = HealthStatusUnhealthy
		health.Message = "Metrics collector is not initialized"
	} else {
		metrics := m.metricsCollector.GetMetrics()
		health.Details = map[string]interface{}{
			"uptime":                    m.metricsCollector.GetUptime().String(),
			"time_since_last_activity": m.metricsCollector.GetTimeSinceLastActivity().String(),
			"total_operations":         metrics.LoginAttempts + metrics.RegistrationAttempts + metrics.TokenValidations,
		}
	}

	health.Duration = time.Since(start).String()
	return health
}

// checkLoggerHealth checks the health of the logger
func (m *Monitor) checkLoggerHealth() ComponentHealth {
	start := time.Now()
	
	health := ComponentHealth{
		Name:      "logger",
		Status:    HealthStatusHealthy,
		Message:   "Logger is operational",
		CheckedAt: time.Now(),
	}

	if m.logger == nil {
		health.Status = HealthStatusDegraded
		health.Message = "Logger is not initialized"
	}

	health.Duration = time.Since(start).String()
	return health
}

// GetSystemInfo returns detailed system information
func (m *Monitor) GetSystemInfo() SystemInfo {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	info := SystemInfo{
		Name:      m.appName,
		Version:   m.version,
		StartTime: m.startTime,
		Uptime:    time.Since(m.startTime).String(),

		GoVersion:    runtime.Version(),
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),

		MemoryAlloc:      memStats.Alloc,
		MemoryTotalAlloc: memStats.TotalAlloc,
		MemorySys:        memStats.Sys,
		MemoryNumGC:      memStats.NumGC,

		DatabaseType: "unknown",
		DatabaseStatus: "unknown",
	}

	// Get database type information
	if m.storage != nil {
		if err := m.storage.Ping(); err == nil {
			info.DatabaseStatus = "connected"
		} else {
			info.DatabaseStatus = "disconnected"
		}
		
		// Try to determine database type from storage implementation
		switch m.storage.(type) {
		case interface{ IsSQLite() bool }:
			info.DatabaseType = "sqlite"
		case interface{ IsPostgres() bool }:
			info.DatabaseType = "postgres"
		default:
			info.DatabaseType = "memory"
		}
	}

	// Add metrics if available
	if m.metricsCollector != nil {
		info.Metrics = m.metricsCollector.GetMetrics()
	}

	return info
}

// HTTPHealthHandler returns an HTTP handler for health checks
func (m *Monitor) HTTPHealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := m.CheckHealth()
		
		// Set appropriate HTTP status code
		statusCode := http.StatusOK
		if health.Status == HealthStatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if health.Status == HealthStatusDegraded {
			statusCode = http.StatusPartialContent
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		
		if err := json.NewEncoder(w).Encode(health); err != nil {
			m.logger.Error("Failed to encode health response", map[string]interface{}{
				"error": err,
			})
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// HTTPMetricsHandler returns an HTTP handler for metrics
func (m *Monitor) HTTPMetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if m.metricsCollector == nil {
			http.Error(w, "Metrics collector not available", http.StatusServiceUnavailable)
			return
		}

		metrics := m.metricsCollector.GetMetrics()
		
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			m.logger.Error("Failed to encode metrics response", map[string]interface{}{
				"error": err,
			})
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// HTTPSystemInfoHandler returns an HTTP handler for system information
func (m *Monitor) HTTPSystemInfoHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		info := m.GetSystemInfo()
		
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(info); err != nil {
			m.logger.Error("Failed to encode system info response", map[string]interface{}{
				"error": err,
			})
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// HTTPReadinessHandler returns an HTTP handler for readiness checks
func (m *Monitor) HTTPReadinessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple readiness check - just verify database connectivity
		if err := m.storage.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "Service not ready: %v", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Ready")
	}
}

// HTTPLivenessHandler returns an HTTP handler for liveness checks
func (m *Monitor) HTTPLivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple liveness check - service is alive if it can respond
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Alive")
	}
}

// RegisterHTTPHandlers registers all monitoring HTTP handlers with a mux
func (m *Monitor) RegisterHTTPHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/health", m.HTTPHealthHandler())
	mux.HandleFunc("/health/ready", m.HTTPReadinessHandler())
	mux.HandleFunc("/health/live", m.HTTPLivenessHandler())
	mux.HandleFunc("/metrics", m.HTTPMetricsHandler())
	mux.HandleFunc("/info", m.HTTPSystemInfoHandler())
}