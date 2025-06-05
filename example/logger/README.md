# Logger Example

This example demonstrates the comprehensive logging capabilities of FastApp, including different log levels, structured logging, context-aware logging, and error handling patterns.

## Features Demonstrated

### 1. Log Levels
- **Debug**: Detailed information for development and troubleshooting
- **Info**: General application flow and important events
- **Warn**: Unexpected situations that don't break functionality
- **Error**: Error conditions that need attention
- **Fatal**: Critical errors that cause application shutdown

### 2. Structured Logging
- Key-value pairs for machine-readable logs
- Support for different data types (strings, numbers, booleans, arrays, objects)
- Consistent field naming for better log analysis

### 3. Context-Aware Logging
- Adding fields to context that appear in all subsequent logs
- Nested contexts with additional fields
- Trace ID and session tracking

### 4. Error Logging Patterns
- Simple error logging
- Error with additional context
- Wrapped errors with original error preservation

## Running the Example

### Using Configuration File

```bash
cd example/logger
go run main.go
```

### Using Environment Variables

```bash
cd example/logger
cp .env.example .env
# Edit .env file as needed
export $(cat .env | xargs)
go run main.go
```

### Changing Log Level

Edit `config.yaml` and change the `Level` field:

```yaml
App:
  Logger:
    Level: "debug"  # Change to: debug, info, warn, error, fatal
```

Or set environment variable:
```bash
export APP_LOGGER_LEVEL=debug
```

### Development vs Production Logging

**Development Mode** (`DevMode: true`):
- Colored console output
- Human-readable format
- Great for local development

**Production Mode** (`DevMode: false`):
- JSON structured output
- Machine-readable format
- Better for log aggregation systems

## Configuration Options

### Logger Configuration

```yaml
App:
  Logger:
    AppName: "my-app"        # Application name in logs
    Level: "info"            # Minimum log level
    DevMode: true            # Console vs JSON output
    MessageKey: "message"    # JSON key for message
    LevelKey: "level"        # JSON key for level
    TimeKey: "timestamp"     # JSON key for timestamp
```

### Environment Variables

All configuration can be overridden with environment variables using the `APP_` prefix:

- `APP_LOGGER_APPNAME`
- `APP_LOGGER_LEVEL`
- `APP_LOGGER_DEVMODE`
- `APP_LOGGER_MESSAGEKEY`
- `APP_LOGGER_LEVELKEY`
- `APP_LOGGER_TIMEKEY`

## Best Practices

### 1. Use Appropriate Log Levels
```go
// Debug - detailed information for development
logger.Debug(ctx, "Processing user input", "input_length", len(input))

// Info - general application flow
logger.Info(ctx, "User logged in successfully", "user_id", userID)

// Warn - unexpected but handled situations
logger.Warn(ctx, "API rate limit approaching", "current_rate", rate)

// Error - errors that need attention
logger.Error(ctx, "Failed to save user data", "error", err, "user_id", userID)
```

### 2. Use Structured Logging
```go
// Good - structured with key-value pairs
logger.InfoKV(ctx, "Order processed",
    "order_id", orderID,
    "customer_id", customerID,
    "amount", amount,
    "processing_time_ms", duration.Milliseconds())

// Avoid - unstructured string formatting
logger.Info(ctx, fmt.Sprintf("Order %s processed for customer %d", orderID, customerID))
```

### 3. Use Context for Correlation
```go
// Add correlation fields to context
ctx = logger.WithFields(ctx,
    "request_id", requestID,
    "user_id", userID,
    "trace_id", traceID)

// All subsequent logs will include these fields
logger.Info(ctx, "Starting payment process")
logger.Info(ctx, "Payment completed")
```

### 4. Include Relevant Context in Errors
```go
logger.ErrorKV(ctx, "Database operation failed",
    "error", err,
    "operation", "user_update",
    "user_id", userID,
    "table", "users",
    "retry_count", retryCount)
```

## Observability Endpoints

When running the example, the following endpoints are available:

- **Metrics**: http://localhost:9090/metrics
- **Health Checks**: http://localhost:9090/health/checks
- **Liveness**: http://localhost:9090/health/live
- **Readiness**: http://localhost:9090/health/ready
- **Profiling**: http://localhost:9090/debug/pprof/

## Sample Output

### Development Mode (DevMode: true)
```
2024-01-15T10:30:45.123Z	INFO	logger-demo	🚀 Starting Logger Demo Service	{"service": "logger-demo"}
2024-01-15T10:30:45.124Z	DEBUG	logger-demo	This is a debug message - useful for development
2024-01-15T10:30:45.125Z	INFO	logger-demo	User action completed	{"user_id": 12345, "action": "profile_update", "duration_ms": 150}
```

### Production Mode (DevMode: false)
```json
{"timestamp":"2024-01-15T10:30:45.123Z","level":"info","application_name":"logger-demo","message":"🚀 Starting Logger Demo Service","service":"logger-demo"}
{"timestamp":"2024-01-15T10:30:45.124Z","level":"debug","application_name":"logger-demo","message":"This is a debug message - useful for development"}
{"timestamp":"2024-01-15T10:30:45.125Z","level":"info","application_name":"logger-demo","message":"User action completed","user_id":12345,"action":"profile_update","duration_ms":150}
```
