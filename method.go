package gin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"runtime/debug"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/illidaris/core"
	"github.com/illidaris/logger"
	"github.com/illidaris/signature"
	"go.uber.org/zap"
)

// LoggerHandler record log
func LoggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// content type
		contentType := c.GetHeader("Content-Type")
		sessionBirth := time.Now()
		// trace
		WithTrace(c, sessionBirth)
		// before
		c.Next()
		// after
		cost := time.Since(sessionBirth)
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		curCtx := c.Request.Context()
		curCtx = logger.NewContext(curCtx,
			zap.String(string(HTTPMethod), c.Request.Method),
			zap.String(string(HTTPContentType), contentType),
			zap.String(string(HTTPPath), path),
			zap.String(string(HTTPQuery), query),
			zap.String(string(HTTPClientIP), c.ClientIP()),
			zap.String(string(HTTPUserAgent), c.Request.UserAgent()),
			zap.Int(string(HTTPStatusCode), c.Writer.Status()),
			zap.Int64(core.Duration.String(), cost.Milliseconds()),
		)
		logger.InfoCtx(curCtx, c.Errors.ByType(gin.ErrorTypePrivate).String())
	}
}

func SignatureMiddleware(f func(ctx context.Context, key string) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		s := signature.DefaultSign(c.Request)
		sign := signature.NewSignature()
		secret := f(c.Request.Context(), s.AppID)
		if b, err := sign.Verify(c.Request.Context(), s, secret); !b {
			c.AbortWithError(http.StatusBadRequest, err)
		}
		c.Next()
	}
}

// RecoverHandler recover from panic
func RecoverHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				fmt.Println(err)
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					var se *os.SyscallError
					if ok := errors.Is(ne.Err, se); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					logger.WithContext(c.Request.Context()).Error(c.Request.URL.Path, zap.Any(core.Error.String(), err), zap.String("HttpRequest", string(httpRequest)))
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: error check
					c.Abort()
					return
				}

				logger.WithContext(c.Request.Context()).Error("recover from panic",
					zap.Any(core.Error.String(), err),
					zap.String("HttpRequest", string(httpRequest)),
					zap.String("stack", string(debug.Stack())))

				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

func GracefulRun(ctx context.Context, e http.Handler, addr string, timeout time.Duration) {
	// bind ip&port
	srv := &http.Server{
		Addr:    addr,
		Handler: e,
	}

	errCh := make(chan error, 1)
	defer close(errCh)
	// listen
	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	notifyCtx, stop := signal.NotifyContext(ctx, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case s := <-notifyCtx.Done():
		stop()
		logger.Info(fmt.Sprintf("Shutdown: Receive Sign(%s)", s))
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		if err := srv.Shutdown(timeoutCtx); err != nil {
			logger.Error(fmt.Sprintf("Shutdown: %s", err))
		}
		logger.Info("Shutdown: exit")
		break
	case err := <-errCh:
		logger.Error(fmt.Sprintf("Listen: Receive Error %s", err))
		break
	}
}
