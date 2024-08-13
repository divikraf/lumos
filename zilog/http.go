package zilog

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/rs/zerolog"
	"gitlab.com/divikraf/lumos/zilog/hook"
)

// HTTPLogMiddlewareOption is a functional option to customize
// the logging behavior of HTTPMuxMiddleware.
//
// WARNING: DO NOT CALL logEvent.Msg() HERE BECAUSE IT CAUSES PANIC. LET OUR
// UnaryServerInterceptor DO THE logEvent.Msg(). TO DISCARD THE LOG, YOU CAN USE
// logEvent.Discard() INSTEAD.
type HTTPLogMiddlewareOption interface {
	Pre(*HTTPLogMiddlewareCfg, *http.Request)
	Post(*HTTPLogMiddlewareCfg, *zerolog.Event, *http.Request, *bytes.Buffer, WrapResponseWriter)
}

type logHTTPRequest struct{}

func (logHTTPRequest) Pre(cfg *HTTPLogMiddlewareCfg, r *http.Request) {
	cfg.WithResponse = true
}

func (op logHTTPRequest) Post(cfg *HTTPLogMiddlewareCfg, logEvent *zerolog.Event, r *http.Request, response *bytes.Buffer, wrw WrapResponseWriter) { //nolint:revive // it's normal for this middleware
	body, _ := io.ReadAll(r.Body)
	logEvent.
		Interface("request.query", r.URL.Query()).
		Interface("request.body", body).
		Interface("request.form", r.Form)
}

// WithLogHTTPRequest append full request to log.
func WithLogHTTPRequest() HTTPLogMiddlewareOption {
	return logHTTPRequest{}
}

type logHTTPResponse struct{}

func (logHTTPResponse) Pre(cfg *HTTPLogMiddlewareCfg, r *http.Request) {
	cfg.WithResponse = true
}

func (op logHTTPResponse) Post(cfg *HTTPLogMiddlewareCfg, logEvent *zerolog.Event, r *http.Request, response *bytes.Buffer, wrw WrapResponseWriter) { //nolint:revive // it's normal for this middleware
	logEvent.
		Interface("response.header", wrw.Header())
}

// WithLogHTTPResponse append full response to log.
func WithLogHTTPResponse() HTTPLogMiddlewareOption {
	return logHTTPResponse{}
}

// HTTPLogMiddlewareCfg determines the behavior of HTTPMuxMiddleware.
type HTTPLogMiddlewareCfg struct {
	WithRequest  bool
	WithResponse bool
}

// HTTPLogMiddleware embeds zerolog.Logger into context.
func HTTPLogMiddleware(opts ...HTTPLogMiddlewareOption) func(c *gin.Context) {
	return func(c *gin.Context) {
		txn := nrgin.Transaction(c)
		r := c.Request
		w := c.Writer
		cfg := HTTPLogMiddlewareCfg{
			WithRequest:  false,
			WithResponse: false,
		}
		for _, o := range opts {
			o.Pre(&cfg, r)
		}
		rCtx := r.Context()
		newCtx, _ := NewContext(rCtx, hook.NewHTTPPath(r.URL.EscapedPath()), hook.NewRelicRecorderHook(txn))

		c.Request = c.Request.WithContext(newCtx)

		ww := newWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()

		buf := []byte("")
		resp := bytes.NewBuffer(buf)
		if cfg.WithResponse {
			ww.Tee(resp)
		}

		defer func() {
			logger := FromContext(newCtx)
			logEvent := logger.Info()
			if ww.Status() >= 500 {
				logEvent = logger.Error()
			} else if ww.Status() >= 400 {
				logEvent = logger.Warn()
			}

			logEvent.Dur("http.dur", time.Since(t1))

			for _, o := range opts {
				o.Post(&cfg, logEvent, r, resp, ww)
			}

			logEvent.
				Int("http.status", ww.Status()).
				Int("http.bytesw", ww.BytesWritten()).
				Msg(fmt.Sprintf("%s %s", r.Method, r.URL.Path))
		}()

		c.Next()
	}
}
