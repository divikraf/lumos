# Lumos 🌟

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/divikraf/lumos)](https://goreportcard.com/report/github.com/divikraf/lumos)

**Lumos** is a comprehensive Go microservices framework that provides essential building blocks for modern cloud-native applications. Built with simplicity and performance in mind, Lumos offers a cohesive set of tools for configuration management, logging, telemetry, database operations, validation, and more.

## ✨ Features

### 🎯 **Core Components**
- **Configuration Management** (`ziconf`) - Flexible, type-safe configuration loading with YAML support
- **Structured Logging** (`zilog`) - High-performance logging with zerolog and slog integration
- **Observability** (`zitelemetry`) - OpenTelemetry-based tracing and metrics with no-op fallbacks
- **Database Operations** (`zisqlx`) - SQLx wrapper with built-in metrics and tracing
- **Input Validation** (`zivalidator`) - Request validation with internationalized error messages
- **Internationalization** (`i18n`) - Context-based language handling
- **HTTP Routing** (`zin`) - Gin-based HTTP server with middleware support
- **Dependency Injection** (`zilong`) - FX-based DI container with opinionated defaults

### 🚀 **Key Benefits**
- **Zero Configuration** - Sensible defaults for rapid development
- **Production Ready** - Built-in observability, metrics, and error handling
- **Type Safe** - Full compile-time type checking
- **Context Aware** - Proper context propagation throughout the stack
- **Cloud Native** - Designed for containerized and distributed environments
- **Performance First** - Optimized for high-throughput applications

## 📦 Installation

```bash
go get github.com/divikraf/lumos
```

## 🏃‍♂️ Quick Start

### Basic HTTP Server

```go
package main

import (
    "time"
    
    "github.com/divikraf/lumos/ziconf"
    "github.com/divikraf/lumos/zilog"
    "github.com/divikraf/lumos/zitelemetry/observe"
    "github.com/gin-gonic/gin"
    "go.uber.org/fx"
)

// AppConfig represents your application configuration
type AppConfig struct {
    Service   ziconf.ServiceConfig `json:"service" yaml:"service"`
    Log       ziconf.LogConfig     `json:"log" yaml:"log"`
    Telemetry observe.Config       `json:"telemetry" yaml:"telemetry"`
    HttpPort  string               `json:"http_port" yaml:"http_port"`
}

// Implement ziconf.Config interface
func (c AppConfig) GetService() ziconf.ServiceConfig { return c.Service }
func (c AppConfig) GetEnvironment() string          { return "development" }
func (c AppConfig) GetLog() ziconf.LogConfig        { return c.Log }
func (c AppConfig) GetHttpPort() string             { return c.HttpPort }
func (c AppConfig) GetTelemetry() observe.Config    { return c.Telemetry }

// Handler
type Handler struct{}

func (h *Handler) GetUsers(c *gin.Context) {
    logger := zilog.FromContext(c.Request.Context())
    
    logger.Info().
        Str("endpoint", "GET /users").
        Msg("Fetching users")
    
    users := []map[string]interface{}{
        {"id": 1, "name": "John Doe"},
        {"id": 2, "name": "Jane Smith"},
    }
    
    c.JSON(200, gin.H{"data": users})
}

func main() {
    app := fx.New(
        // Configuration
        fx.Provide(func() AppConfig {
            return *ziconf.ReadConfig[AppConfig]()
        }),
        
        // Services
        fx.Provide(func() *Handler { return &Handler{} }),
        fx.Provide(func() *gin.Engine { return gin.Default() }),
        
        // Routes
        fx.Invoke(func(r *gin.Engine, h *Handler) {
            r.GET("/users", h.GetUsers)
        }),
    )
    
    app.Run()
}
```

### Configuration File (`config.yaml`)

```yaml
service:
  name: "my-service"
  code: "MY_SERVICE"

http_port: ":8080"

log:
  level: "info"

telemetry:
  enabled: true
  tracing:
    enabled: true
    exporter:
      type: "console"
      endpoint: ""
  metrics:
    enabled: true
    exporter:
      type: "console"
      endpoint: ""
```

## 🔧 Components Overview

### Configuration Management (`ziconf`)

Type-safe configuration loading with YAML support:

```go
type Config struct {
    Service   ziconf.ServiceConfig `json:"service" yaml:"service"`
    Database  DatabaseConfig       `json:"database" yaml:"database"`
    HttpPort  string               `json:"http_port" yaml:"http_port"`
}

// Implement the interface
func (c Config) GetService() ziconf.ServiceConfig { return c.Service }
func (c Config) GetEnvironment() string          { return "development" }
func (c Config) GetLog() ziconf.LogConfig        { return ziconf.LogConfig{Level: "info"} }
func (c Config) GetHttpPort() string             { return c.HttpPort }
func (c Config) GetTelemetry() observe.Config    { return observe.Config{} }

// Load configuration
config := ziconf.ReadConfig[Config]()
```

### Structured Logging (`zilog`)

High-performance logging with context support:

```go
// Get logger from context
logger := zilog.FromContext(ctx)

logger.Info().
    Str("user_id", "123").
    Str("action", "login").
    Msg("User logged in successfully")

logger.Error().
    Err(err).
    Str("operation", "database_query").
    Msg("Database operation failed")

// Create context with logger
ctx, logger := zilog.NewContext(ctx, hooks...)
```

### Observability (`zitelemetry`)

OpenTelemetry integration with no-op fallbacks:

```go
// Get tracer from context
tracer := observe.EnhancedFromContext(ctx)

// Start span
spanCtx, span := tracer.Start(ctx, "operation-name")
defer span.End()

span.SetAttributes(
    attribute.String("user.id", "123"),
    attribute.String("operation.type", "query"),
)

// Create tracer from config
tracer := observe.CreateTracerFromConfig(config, "service-name")
```

### Database Operations (`zisqlx`)

SQLx wrapper with metrics and tracing:

```go
// Create database wrapper
db := zisqlx.New(sqlxDB)

// Query operations
var users []User
err := db.SelectContext(ctx, "get_all_users", &users, 
    "SELECT * FROM users WHERE active = $1", true)

var user User
err := db.GetContext(ctx, "get_user_by_id", &user,
    "SELECT * FROM users WHERE id = $1", userID)

// Transaction support
tx, err := db.BeginTx(ctx, "create_user", nil)
if err != nil {
    return err
}
defer tx.Rollback()

// Use transaction
err = tx.ExecContext(ctx, "insert_user", 
    "INSERT INTO users (name, email) VALUES ($1, $2)", name, email)
if err != nil {
    return err
}

return tx.Commit()
```

### Input Validation (`zivalidator`)

Request validation with internationalized errors:

```go
type CreateUserRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=100"`
    Email string `json:"email" validate:"required,email"`
}

validator := zivalidator.New()

func (h *Handler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "Invalid JSON"})
        return
    }
    
    // Validate request
    result := validator.ValidateStruct(c.Request.Context(), req)
    if result != nil && len(result.FieldErrors) > 0 {
        c.JSON(400, gin.H{
            "error": "Validation failed",
            "field_errors": result.FieldErrors,
        })
        return
    }
    
    // Process request...
}
```

### Internationalization (`i18n`)

Context-based language handling:

```go
// Set language in context
ctx := i18n.WithContext(ctx, language.English)

// Get language from context
lang := i18n.FromContext(ctx) // Returns language.English

// Use in handlers
func (h *Handler) GetMessage(c *gin.Context) {
    lang := i18n.FromContext(c.Request.Context())
    
    var message string
    switch lang.String() {
    case "en":
        message = "Hello, World!"
    case "id":
        message = "Halo, Dunia!"
    default:
        message = "Hello, World!"
    }
    
    c.JSON(200, gin.H{"message": message})
}
```

## 🏗️ Architecture

Lumos follows a modular architecture where each component is independently usable:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   ziconf        │    │   zilog         │    │  zitelemetry    │
│   Configuration │    │   Logging       │    │  Observability  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   zisqlx        │
                    │   Database      │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │   zivalidator   │
                    │   Validation    │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │   i18n          │
                    │   I18n          │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │   zin           │
                    │   HTTP Router   │
                    └─────────────────┘
```

## 📊 Observability

Lumos provides comprehensive observability out of the box:

### Metrics
- Database operation duration and error rates
- HTTP request metrics
- Custom business metrics

### Tracing
- Distributed tracing with OpenTelemetry
- Automatic span creation for database operations
- Custom span creation for business logic

### Logging
- Structured JSON logging
- Context-aware logging
- Performance optimized

## 🔧 Configuration

### Telemetry Configuration

```yaml
telemetry:
  enabled: true
  tracing:
    enabled: true
    exporter:
      type: "otlp"  # or "console", "jaeger"
      endpoint: "http://localhost:4317"
      protocol: "grpc"
    sampler:
      type: "traceidratio"
      fraction: 0.1
  metrics:
    enabled: true
    exporter:
      type: "otlp"
      endpoint: "http://localhost:4317"
    reader:
      interval: "10s"
      timeout: "5s"
```

### Database Configuration

```yaml
database:
  postgresql:
    enabled: true
    host: "localhost"
    port: 5432
    database: "myapp"
    username: "user"
    password: "password"
    ssl_mode: "disable"
```

## 🚀 Production Deployment

### Docker Example

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
CMD ["./main"]
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: lumos-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: lumos-app
  template:
    metadata:
      labels:
        app: lumos-app
    spec:
      containers:
      - name: lumos-app
        image: lumos-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENVIRONMENT
          value: "production"
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "200m"
```

## 🧪 Testing

Lumos components are designed to be easily testable:

```go
func TestService(t *testing.T) {
    // Create test context
    ctx := context.Background()
    
    // Create service
    service := NewService()
    
    // Test with mock data
    result, err := service.ProcessData(ctx, "test")
    assert.NoError(t, err)
    assert.Equal(t, "processed_test", result)
}

func TestHandler(t *testing.T) {
    gin.SetMode(gin.TestMode)
    
    // Create test dependencies
    service := NewService()
    handler := NewHandler(service)
    
    // Create test router
    router := gin.New()
    router.POST("/process", handler.ProcessData)
    
    // Create test request
    reqBody := gin.H{"data": "test"}
    reqJSON, _ := json.Marshal(reqBody)
    
    req := httptest.NewRequest("POST", "/process", bytes.NewBuffer(reqJSON))
    req.Header.Set("Content-Type", "application/json")
    
    // Perform request
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
    
    // Assert response
    assert.Equal(t, 200, w.Code)
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/divikraf/lumos.git
cd lumos

# Install dependencies
go mod download

# Run tests
go test ./...

# Run linter
golangci-lint run

# Build examples
go build ./examples/...
```

## 📚 Examples

Check out the [examples](./examples/) directory for comprehensive usage examples:

- **Basic** - Simple HTTP server
- **Database** - PostgreSQL integration
- **Telemetry** - Observability features
- **Middleware** - Custom middleware
- **Testing** - Testing strategies
- **Full-Featured** - Complete application

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [SQLx](https://github.com/jmoiron/sqlx) - SQL toolkit
- [Zerolog](https://github.com/rs/zerolog) - Fast logging
- [OpenTelemetry](https://opentelemetry.io/) - Observability framework
- [FX](https://github.com/uber-go/fx) - Dependency injection
- [Validator](https://github.com/go-playground/validator) - Struct validation

## 📞 Support

- 📖 [Documentation](https://github.com/divikraf/lumos/wiki)
- 🐛 [Issue Tracker](https://github.com/divikraf/lumos/issues)
- 💬 [Discussions](https://github.com/divikraf/lumos/discussions)

---

**Made with ❤️ for the Go community**
