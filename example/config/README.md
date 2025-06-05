# Configuration Example

This example demonstrates comprehensive configuration management in FastApp, including loading from files, environment variables, validation, and usage patterns.

## Features Demonstrated

### 1. Configuration Sources
- **YAML files** - Human-readable configuration
- **JSON files** - Structured configuration for tools
- **Environment variables** - Runtime configuration override
- **Default values** - Fallback configuration via struct tags

### 2. Configuration Structure
- **FastApp configuration** - Framework-level settings
- **Custom sections** - Application-specific configuration
- **Nested structures** - Complex configuration organization
- **Type safety** - Compile-time configuration validation

### 3. Configuration Patterns
- **Feature flags** - Runtime behavior control
- **Environment-specific settings** - Dev/staging/prod configurations
- **Sensitive data handling** - Secure configuration management
- **Validation** - Configuration correctness checks

## Running the Example

### Method 1: Using YAML Configuration

```bash
cd example/config
go run main.go
```

This loads `config.yaml` by default.

### Method 2: Using JSON Configuration

```bash
cd example/config
CONFIG_FILE=config.json go run main.go
```

### Method 3: Using Environment Variables

```bash
cd example/config
cp .env.example .env
# Edit .env file as needed
export $(cat .env | xargs)
go run main.go
```

### Method 4: Mixed Configuration

```bash
cd example/config
# Load base config from file, override with env vars
export DATABASE_MAXCONNECTIONS=50
export FEATURES_ENABLENEWUI=true
go run main.go
```

## Configuration Structure

### FastApp Configuration

```go
type AppConfig struct {
    App config.App  // FastApp framework configuration
    
    // Your custom configuration sections
    Database DatabaseConfig
    Redis    RedisConfig
    API      APIConfig
    Features FeatureFlags
}
```

### Custom Configuration Sections

```go
type DatabaseConfig struct {
    URL             string        `default:"postgres://..."`
    MaxConnections  int           `default:"25"`
    MaxIdleTime     time.Duration `default:"15m"`
    ConnectTimeout  time.Duration `default:"10s"`
    QueryTimeout    time.Duration `default:"30s"`
    EnableLogging   bool          `default:"false"`
    MigrationsPath  string        `default:"./migrations"`
}
```

## Configuration Loading Priority

Configuration values are loaded in the following order (later sources override earlier ones):

1. **Struct default tags** - `default:"value"`
2. **Configuration file** - YAML/JSON file
3. **Environment variables** - `APP_LOGGER_LEVEL`, `DATABASE_URL`, etc.

### Environment Variable Naming

Environment variables use uppercase with underscores:

- `App.Logger.Level` → `APP_LOGGER_LEVEL`
- `Database.MaxConnections` → `DATABASE_MAXCONNECTIONS`
- `Features.EnableNewUI` → `FEATURES_ENABLENEWUI`

## Configuration Files

### YAML Format (config.yaml)

```yaml
App:
  Logger:
    AppName: "my-app"
    Level: "info"
    DevMode: true

Database:
  URL: "postgres://user:password@localhost:5432/mydb"
  MaxConnections: 25
  MaxIdleTime: "15m"

Features:
  EnableNewUI: false
  EnableCaching: true
```

### JSON Format (config.json)

```json
{
  "App": {
    "Logger": {
      "AppName": "my-app",
      "Level": "info",
      "DevMode": true
    }
  },
  "Database": {
    "URL": "postgres://user:password@localhost:5432/mydb",
    "MaxConnections": 25,
    "MaxIdleTime": "15m"
  }
}
```

## Environment-Specific Configuration

### Development

```yaml
App:
  Logger:
    Level: "debug"
    DevMode: true
Database:
  EnableLogging: true
Features:
  EnableBetaAPI: true
```

### Production

```yaml
App:
  Logger:
    Level: "info"
    DevMode: false
Database:
  MaxConnections: 100
  EnableLogging: false
Features:
  EnableBetaAPI: false
```

## Best Practices

### 1. Use Struct Tags for Defaults

```go
type Config struct {
    Port    int           `default:"8080"`
    Timeout time.Duration `default:"30s"`
    Enabled bool          `default:"true"`
}
```

### 2. Group Related Configuration

```go
type AppConfig struct {
    App      config.App
    Database DatabaseConfig
    Cache    CacheConfig
    External ExternalServicesConfig
}
```

### 3. Validate Configuration

```go
func (c *DatabaseConfig) Validate() error {
    if c.MaxConnections <= 0 {
        return errors.New("max connections must be positive")
    }
    if c.URL == "" {
        return errors.New("database URL is required")
    }
    return nil
}
```

### 4. Use Feature Flags

```go
type FeatureFlags struct {
    EnableNewUI     bool `default:"false"`
    EnableBetaAPI   bool `default:"false"`
    MaintenanceMode bool `default:"false"`
}

// Usage
if cfg.Features.EnableNewUI {
    // Use new UI
} else {
    // Use legacy UI
}
```

### 5. Handle Sensitive Data

```go
// Don't log sensitive configuration
func (c *DatabaseConfig) String() string {
    return fmt.Sprintf("Database{MaxConnections: %d, URL: %s}", 
        c.MaxConnections, maskURL(c.URL))
}
```

## Configuration Loading Examples

### Basic Loading

```go
cfg, err := configloader.New[AppConfig]()
if err != nil {
    log.Fatal("Failed to load config:", err)
}
```

### With File Override

```go
cfg, err := configloader.New[AppConfig](
    configloader.WithFile("config.yaml", "config.json"),
)
```

### With Environment File Path

```go
cfg, err := configloader.New[AppConfig](
    configloader.WithFileFromEnv("config.yaml"),
)
```

### With Custom Environment Prefix

```go
cfg, err := configloader.New[AppConfig](
    configloader.WithEnv("MYAPP_"),
)
```

## Observability Endpoints

When running the example, the following endpoints are available:

- **Metrics**: http://localhost:9090/metrics
- **Health Checks**: http://localhost:9090/health/checks
- **Liveness**: http://localhost:9090/health/live
- **Readiness**: http://localhost:9090/health/ready
- **Profiling**: http://localhost:9090/debug/pprof/

## Troubleshooting

### Configuration Not Loading

1. Check file path and permissions
2. Verify YAML/JSON syntax
3. Check environment variable names
4. Enable debug logging: `APP_LOGGER_LEVEL=debug`

### Environment Variables Not Working

1. Verify variable names match struct fields
2. Check for typos in variable names
3. Ensure variables are exported: `export VAR=value`
4. Use correct data types (strings for durations: `"30s"`)

### Default Values Not Applied

1. Ensure struct tags are correct: `default:"value"`
2. Check that configloader processes defaults
3. Verify field types match default values
