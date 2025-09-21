package zin

import (
	"context"
	"log"
	"net/http"

	"github.com/divikraf/lumos/ziconf"
	"github.com/divikraf/lumos/zilog"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/fx"
)

type InitRouterParams struct {
	fx.In
	Config         ziconf.Config
	TracerProvider trace.TracerProvider
	SkipPaths      []string `group:"http-metrics-skip-paths"`
}

func RegiterRouter(params InitRouterParams) *gin.Engine {
	router := gin.New()
	router.Use(otelgin.Middleware(params.Config.GetService().Name))
	router.Use(zilog.HTTPLogMiddleware(zilog.WithLogHTTPRequest(), zilog.WithLogHTTPResponse()))
	// Use skip paths from FX groups
	router.Use(httpMetricsMiddlewareWithSkipPaths(params.SkipPaths))
	router.Use(gin.Recovery())

	return router
}

type HttpServerParams struct {
	fx.In

	LC     fx.Lifecycle
	Logger *zerolog.Logger
	Config ziconf.Config
	Router *gin.Engine
}

func StartHttpServer(params HttpServerParams) {
	srv := &http.Server{
		Addr:    params.Config.GetHttpPort(),
		Handler: params.Router.Handler(),
	}

	params.LC.Append(fx.StartHook(func() error {
		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Could not listen on %s: %v\n", srv.Addr, err)
			}
		}()
		return nil
	}))

	params.LC.Append(fx.StopHook(func(ctx context.Context) {
		srv.Shutdown(ctx)
	}))
}
