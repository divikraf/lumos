package main

import (
	"fmt"
	"net/http"

	"github.com/divikraf/lumos/ziconf"
	"github.com/divikraf/lumos/zilog"
	"github.com/divikraf/lumos/zilong"
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"go.uber.org/fx"
)

var UserModule = fx.Module(
	"user",
	fx.Invoke(RegisterUserRoutes),
)

// RegisterUserRoutes registers the user-related routes.
func RegisterUserRoutes(router *gin.Engine) {
	userGroup := router.Group("/user")
	{
		userGroup.GET("/profile", func(c *gin.Context) {
			zlog := zilog.FromContext(c.Request.Context())
			txn := nrgin.Transaction(c)
			fmt.Println(txn.Name())
			txn.AddAttribute("test", "testing")
			sg := txn.StartSegment("testing-segment")
			sg.AddAttribute("hai", "hello")
			txn.Application().RecordCustomMetric("user_profile_access_count", 1)
			zlog.Info().Str("attribute", "testing-attribute").Msg("testing info")
			defer sg.End()
			c.JSON(http.StatusOK, gin.H{"message": "User Profile"})
		})
		userGroup.POST("/login", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "User Login"})
		})
	}
}

func main() {
	zilong.App[*Cfg](
		UserModule,
	).Run()
}

type Cfg struct {
	Service ziconf.ServiceConfig `json:"service"`
	Log     ziconf.LogConfig     `json:"log"`
	Http    HttpConfig           `json:"http"`
}

// GetHttpPort implements ziconf.Config.
func (c *Cfg) GetHttpPort() string {
	return c.Http.Port
}

// GetLog implements ziconf.Config.
func (c *Cfg) GetLog() ziconf.LogConfig {
	return c.Log
}

// GetService implements ziconf.Config.
func (c *Cfg) GetService() ziconf.ServiceConfig {
	return c.Service
}

type HttpConfig struct {
	Port string `json:"port"`
}
