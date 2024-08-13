package zin

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog"
	"gitlab.com/divikraf/lumos/ziconf"
	"gitlab.com/divikraf/lumos/zilog"
	"go.uber.org/fx"
)

type InitRouterParams struct {
	fx.In
	NrApp *newrelic.Application
}

func RegiterRouter(params InitRouterParams) *gin.Engine {
	router := gin.New()
	router.Use(nrgin.Middleware(params.NrApp))
	router.Use(zilog.HTTPLogMiddleware(zilog.WithLogHTTPRequest(), zilog.WithLogHTTPResponse()))
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
