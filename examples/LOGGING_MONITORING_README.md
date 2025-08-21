# Logging and Monitoring Features

This document describes the comprehensive logging and monitoring features available in the go-auth package.

## Overview

The go-auth package includes built-in structured logging, metrics collection, and monitoring capabilities that help you track authentication operations, monitor system health, and debug issues in production.

## Features

### 1. Structured Logging

- **JSON-formatted logs** with configurable levels (DEBUG, INFO, WARN, ERROR)
- **Authentication event logging** for all major operations
- **Contextual information** including user IDs, usernames, IP addresses, and operation durations
- **Field-based logging** for easy parsing and filtering

### 2. Metrics Collection

- **Operation counters** for registrations, logins, token operations, etc.
- **Success/failure rates** for all authentication operations
- **Performance metrics** including average operation durations
- **Error tracking** by category (database, validation, authentication)
- **System metrics** including uptime and activity tracking

### 3. Health Monitoring

- **Health check endpoints** for load balancers and monitoring systems
- **Component health status** (database, metrics, logger)
- **System information** including runtime stats and memory usage
- **Readiness and liveness probes** for Kubernetes deployments

### 4. HTTP Monitoring Endpoints

- `GET /health` - Comprehensive health check with component status
- `GET /health/ready` - Readiness probe for deployment orchestration
- `GET /health/live` - Liveness probe for container health
- `GET /metrics` - Raw metrics data in JSON format
- `GET /info` - Detailed system information

## Quick Start

### Basic Usage

```go
package main

import (
    "log"
    "github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
    // Create auth service (logging is enabled by default)
    authService, err := auth.NewSQLite("auth.db", "your-secret-key")
    if err != nil {
        log.Fatal(err)
    }

    // All authentication operations are automatically logged
    user, err := authService.Register(auth.RegisterRequest{
        Username: "john",
        Email:    "john@example.com",
        Password: "securepassword",
    })

    // Login operations include timing and success/failure tracking
    tokens, err := authService.Login("john", "securepassword", nil)

    // Access current metrics
    metrics := authService.GetMetrics()
    log.Printf("Login attempts: %d, Success rate: %.1f%%", 
        metrics.LoginAttempts, 
        authService.MetricsCollector().GetLoginSuccessRate())
}
```

### Custom Logging

```go
// Get the logger for custom logging
logger := authService.Logger()

// Log with structured fields
logger.Info("Custom operation completed", map[string]interface{}{
    "user_id": "123",
    "operation": "profile_update",
    "duration": time.Since(start),
})

// Create a field logger with default fields
userLogger := logger.WithFields(map[string]interface{}{
    "user_id": "123",
    "component": "user_service",
})

userLogger.Info("User operation started")
userLogger.Error("Operation failed", map[string]interface{}{
    "error": err,
})
```

### Setting Up Monitoring Server

```go
package main

import (
    "net/http"
    "github.com/pragneshbagary/go-auth/pkg/auth"
)

func main() {
    authService, _ := auth.NewSQLite("auth.db", "secret")
    
    // Create HTTP server with monitoring endpoints
    mux := http.NewServeMux()
    
    // Register all monitoring endpoints
    monitor := authService.Monitor()
    monitor.RegisterHTTPHandlers(mux)
    
    // Start server
    http.ListenAndServe(":8080", mux)
}
```

## Configuration

### Log Levels

Configure logging level through environment variables or configuration:

```bash
# Environment variable
export AUTH_LOG_LEVEL=debug

# Or through configuration
config := &auth.EnhancedConfig{
    LogLevel: "debug", // debug, info, warn, error
    // ... other config
}
```

### Log Level Behavior

- **DEBUG**: All operations including token validations
- **INFO**: Successful operations (registrations, logins, token refreshes)
- **WARN**: Failed operations (invalid credentials, expired tokens)
- **ERROR**: System errors (database failures, internal errors)

## Log Format

All logs are structured JSON with the following fields:

```json
{
  "timestamp": "2023-08-21T10:30:45Z",
  "level": "INFO",
  "message": "User logged in successfully",
  "component": "go-auth",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "username": "john",
  "event": "user_login",
  "ip": "192.168.1.100",
  "user_agent": "Mozilla/5.0...",
  "duration": "45ms",
  "fields": {
    "success": true
  }
}
```

## Metrics

### Available Metrics

```go
type Metrics struct {
    // Registration metrics
    RegistrationAttempts int64
    RegistrationSuccess  int64
    RegistrationFailures int64

    // Login metrics
    LoginAttempts int64
    LoginSuccess  int64
    LoginFailures int64

    // Token metrics
    TokensGenerated     int64
    TokenRefreshes      int64
    TokenValidations    int64
    TokenRevocations    int64
    TokenValidationFail int64

    // Performance metrics
    AverageLoginDuration    time.Duration
    AverageRefreshDuration  time.Duration
    AverageValidationDuration time.Duration

    // Error metrics
    DatabaseErrors    int64
    ValidationErrors  int64
    AuthenticationErrors int64

    // System metrics
    StartTime    time.Time
    LastActivity time.Time
}
```

### Accessing Metrics

```go
// Get current metrics
metrics := authService.GetMetrics()

// Get metrics collector for additional methods
collector := authService.MetricsCollector()

// Success rates
loginRate := collector.GetLoginSuccessRate()
registrationRate := collector.GetRegistrationSuccessRate()
tokenRate := collector.GetTokenValidationSuccessRate()

// System info
uptime := collector.GetUptime()
lastActivity := collector.GetTimeSinceLastActivity()
```

## Health Monitoring

### Health Check Response

```json
{
  "status": "healthy",
  "timestamp": "2023-08-21T10:30:45Z",
  "uptime": "2h30m15s",
  "version": "2.0.0",
  "components": [
    {
      "name": "database",
      "status": "healthy",
      "message": "Database connection is healthy",
      "checked_at": "2023-08-21T10:30:45Z",
      "duration": "2ms",
      "details": {
        "schema_version": 3
      }
    },
    {
      "name": "metrics",
      "status": "healthy",
      "message": "Metrics collector is operational",
      "checked_at": "2023-08-21T10:30:45Z",
      "duration": "0.1ms"
    }
  ]
}
```

### System Information

```go
// Get detailed system information
info := authService.GetSystemInfo()

// Includes:
// - Application name and version
// - Go version and runtime stats
// - Memory usage and GC stats
// - Database type and status
// - Current metrics
```

## Event Types

The following authentication events are automatically logged:

- **user_registration** - User account creation
- **user_login** - User authentication
- **token_refresh** - Access token renewal
- **token_revocation** - Token invalidation
- **token_validation** - Token verification
- **password_change** - Password updates
- **password_reset** - Password reset operations

## Integration Examples

### With Prometheus

```go
// Export metrics to Prometheus format
func prometheusHandler(authService *auth.Auth) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        metrics := authService.GetMetrics()
        
        fmt.Fprintf(w, "# HELP auth_login_attempts_total Total login attempts\n")
        fmt.Fprintf(w, "# TYPE auth_login_attempts_total counter\n")
        fmt.Fprintf(w, "auth_login_attempts_total %d\n", metrics.LoginAttempts)
        
        fmt.Fprintf(w, "# HELP auth_login_success_total Successful logins\n")
        fmt.Fprintf(w, "# TYPE auth_login_success_total counter\n")
        fmt.Fprintf(w, "auth_login_success_total %d\n", metrics.LoginSuccess)
        
        // ... more metrics
    }
}
```

### With ELK Stack

Configure your log shipper (Filebeat, Fluentd, etc.) to parse the JSON logs:

```yaml
# Filebeat configuration
filebeat.inputs:
- type: log
  paths:
    - /var/log/app/*.log
  json.keys_under_root: true
  json.add_error_key: true

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "auth-logs-%{+yyyy.MM.dd}"
```

### With Kubernetes

```yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
spec:
  ports:
  - name: http
    port: 8080
  - name: health
    port: 8080
  selector:
    app: auth-service

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  template:
    spec:
      containers:
      - name: auth
        image: your-auth-app:latest
        ports:
        - containerPort: 8080
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## Best Practices

### 1. Log Level Configuration

- Use **DEBUG** in development for detailed operation tracking
- Use **INFO** in production for successful operations
- Use **WARN** for failed operations that might indicate issues
- Use **ERROR** for system failures requiring attention

### 2. Metrics Monitoring

- Set up alerts for high failure rates (>5% login failures)
- Monitor average operation durations for performance issues
- Track database errors for infrastructure problems
- Monitor token validation failures for security issues

### 3. Health Checks

- Use `/health/ready` for load balancer health checks
- Use `/health/live` for container orchestration
- Use `/health` for comprehensive monitoring dashboards
- Set appropriate timeouts (2-5 seconds) for health checks

### 4. Security Considerations

- Logs never contain passwords or sensitive token data
- User IDs are logged instead of usernames when possible
- IP addresses and user agents are logged for security analysis
- Failed authentication attempts are logged for security monitoring

## Troubleshooting

### Common Issues

1. **High Memory Usage**: Check metrics collection frequency and consider implementing log rotation
2. **Slow Health Checks**: Database connectivity issues or high load
3. **Missing Logs**: Check log level configuration and output destination
4. **Incorrect Metrics**: Ensure all authentication operations go through the Auth service

### Debug Mode

Enable debug logging to see detailed operation traces:

```bash
export AUTH_LOG_LEVEL=debug
```

This will log all token validations, database queries, and internal operations.

## Example Application

See `examples/logging_monitoring_example.go` for a complete example that demonstrates:

- Setting up logging and monitoring
- Creating HTTP endpoints for metrics
- Performing authentication operations
- Displaying metrics and health information
- Running a monitoring server

Run the example:

```bash
go run examples/logging_monitoring_example.go
```

Then visit:
- http://localhost:8080/health - Health check
- http://localhost:8080/metrics - Raw metrics
- http://localhost:8080/info - System information