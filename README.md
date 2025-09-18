# Lumos

[![Go Version](https://img.shields.io/badge/go-1.22.2-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/divikraf/lumos)](https://goreportcard.com/report/github.com/divikraf/lumos)

A production-ready Go backend foundation framework built with opinionated defaults for rapid development of robust microservices and APIs.

## ✨ Features

- 🏗️ **Dependency Injection** - Built on Uber FX for clean architecture
- 🌐 **HTTP Server** - Gin-based router with middleware stack
- 📊 **Observability** - New Relic APM integration across all components
- 🗄️ **Database Support** - MySQL, PostgreSQL, and Redis connectors
- 📝 **Structured Logging** - Zerolog with context propagation
- ⚙️ **Configuration** - YAML-based config with type safety
- ✅ **Validation** - Multi-language validation with structured errors
- 🌍 **Internationalization** - Built-in i18n support
- 🚀 **Performance** - Optimized for high-throughput scenarios

## 🚀 Quick Start

### Installation

```bash
go get github.com/divikraf/lumos
```

### Basic Usage

1. **Create your configuration struct:**

```go
package main

import (
    "github.com/divikraf/lumos/ziconf"
    "github.com/divikraf/lumos/zilong"
    "github.com/gin-gonic/gin"
    "net/http"
    "go.uber.org/fx"
)

type Config struct {
    Service ziconf.ServiceConfig `json:"service"`
    Log     ziconf.LogConfig     `json:"log"`
    Http    HttpConfig           `json:"http"`
}

type HttpConfig struct {
    Port string `json:"port"`
}

// Implement required interfaces
func (c *Config) GetHttpPort() string { return c.Http.Port }
func (c *Config) GetLog() ziconf.LogConfig { return c.Log }
func (c *Config) GetService() ziconf.ServiceConfig { return c.Service }
```

2. **Create your routes module:**

```go
var UserModule = fx.Module(
    "user",
    fx.Invoke(RegisterUserRoutes),
)

func RegisterUserRoutes(router *gin.Engine) {
    userGroup := router.Group("/user")
    {
        userGroup.GET("/profile", func(c *gin.Context) {
            c.JSON(http.StatusOK, gin.H{"message": "User Profile"})
        })
    }
}
```

3. **Start your application:**

```go
func main() {
    zilong.App[*Config](UserModule).Run()
}
```

4. **Create your configuration file (`config.yaml`):**

```yaml
service:
  name: "my-service"
  code: "my-service-code"
newrelic:
  license_key: "your-newrelic-license-key"
log:
  level: "DEBUG"
http:
  port: ":8080"
```

## 🏗️ Architecture

### Core Components

- **`zilong`** - Main application bootstrap and dependency injection
- **`ziconf`** - Configuration management with YAML support
- **`zilog`** - Structured logging with Zerolog
- **`zin`** - HTTP server and router management
- **`zivalidator`** - Validation system with multi-language support
- **`i18n`** - Internationalization utilities

### Database Connectors

- **`zimysql`** - MySQL connector with connection pooling
- **`zipg`** - PostgreSQL connector with connection pooling
- **`ziredis`** - Redis connector (single and cluster support)
- **`zimemo`** - SQL prepared statement memoization

## 📚 Detailed Usage

### Configuration

Lumos uses a type-safe configuration system. Your config struct must implement the `ziconf.Config` interface:

```go
type Config interface {
    GetService() ServiceConfig
    GetLog() LogConfig
    GetHttpPort() string
}
```

### Logging

Lumos provides structured logging with context propagation:

```go
import "github.com/divikraf/lumos/zilog"

// In your handler
func MyHandler(c *gin.Context) {
    logger := zilog.FromContext(c.Request.Context())
    logger.Info().
        Str("user_id", "123").
        Msg("User accessed endpoint")
}
```

### Database Connections

#### MySQL

```go
import "github.com/divikraf/lumos/db/zimysql"

// In your service
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    db := s.mysqlConnector.Connect(ctx, zimysql.Input{
        HostPort: zimysql.HostPort{Host: "localhost", Port: "3306"},
        Username: "user",
        Password: "password",
        DatabaseName: "mydb",
        ConnConfig: zimysql.ConnectionConfig{
            MaxOpen: 10,
            MaxIdle: 5,
            ConnMaxIdleTime: 30 * time.Minute,
        },
    })
    
    var user User
    err := db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = ?", id)
    return &user, err
}
```

#### Redis

```go
import "github.com/divikraf/lumos/db/ziredis"

// Single Redis
client := redisConnector.ConnectSingle(ctx, ziredis.InputSingle{
    HostPort: ziredis.HostPort{Host: "localhost", Port: "6379"},
    Password: "password",
    ConnConfig: ziredis.ConnectionConfig{
        PoolSize: 10,
        MaxLifeTime: 30 * time.Minute,
    },
})

// Redis Cluster
clusterClient := redisConnector.ConnectCluster(ctx, ziredis.InputCluster{
    ClientName: "my-app",
    HostPorts: []ziredis.HostPort{
        {Host: "localhost", Port: "7000"},
        {Host: "localhost", Port: "7001"},
    },
    Password: "password",
    ConnConfig: ziredis.ConnectionConfig{
        PoolSize: 10,
        MaxLifeTime: 30 * time.Minute,
    },
})
```

### Validation

Lumos provides structured validation with multi-language support:

```go
import "github.com/divikraf/lumos/zivalidator"

type UserRequest struct {
    Name  string `json:"name" validate:"required,min=3"`
    Email string `json:"email" validate:"required,email"`
}

func (s *UserService) CreateUser(ctx context.Context, req UserRequest) error {
    result := s.validator.ValidateStruct(ctx, req)
    if result != nil {
        return errors.New(result.Message)
    }
    // Process user creation...
}
```

### Internationalization

```go
import "github.com/divikraf/lumos/i18n"

func MyHandler(c *gin.Context) {
    // Get language from context
    lang := i18n.FromContext(c.Request.Context())
    
    // Use language-specific logic
    if lang == language.English {
        // English logic
    } else {
        // Default (Indonesian) logic
    }
}
```

## 🔧 Configuration

### Service Configuration

```yaml
service:
  name: "my-service"        # Service name
  code: "my-service-code"   # Service code
```

### Logging Configuration

```yaml
log:
  level: "DEBUG"  # DEBUG, INFO, WARN, ERROR
```

### New Relic Configuration

```yaml
newrelic:
  license_key: "your-license-key"
  app_name: "my-app"  # Optional, defaults to service name
```

### HTTP Configuration

```yaml
http:
  port: ":8080"  # Server port
```

## 🧪 Example Application

Check out the `example.go` file for a complete working example that demonstrates:

- Basic HTTP server setup
- Route registration
- New Relic integration
- Structured logging
- Configuration management

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

If you have any questions or need help, please:

1. Check the [documentation](docs/)
2. Search existing [issues](https://github.com/divikraf/lumos/issues)
3. Create a new issue if needed

## 🏷️ Versioning

This project uses [Semantic Versioning](https://semver.org/). For the versions available, see the [tags on this repository](https://github.com/divikraf/lumos/tags).

---

**Lumos** - Illuminate your Go backend development! ✨
