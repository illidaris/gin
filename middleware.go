package gin

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/illidaris/logger"
)

type ParamMiddlewareOption func(opt *paramMiddlewareOptions)

type paramMiddlewareOptions struct {
	RequestContentLengthMax  uint64
	ResponseContentLengthMax uint64
}

func WithRequestContentLengthMax(max uint64) ParamMiddlewareOption {
	return func(opt *paramMiddlewareOptions) {
		opt.RequestContentLengthMax = max
	}
}

func WithResponseContentLengthMax(max uint64) ParamMiddlewareOption {
	return func(opt *paramMiddlewareOptions) {
		opt.ResponseContentLengthMax = max
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w bodyLogWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// ParamMiddleware 出参入参记录中间件
func ParamMiddleware(opts ...ParamMiddlewareOption) gin.HandlerFunc {
	opt := &paramMiddlewareOptions{}
	for _, f := range opts {
		f(opt)
	}
	return func(c *gin.Context) {
		// max > 0  enable log response
		if opt.RequestContentLengthMax > 0 {
			if c.Request.ContentLength < int64(opt.RequestContentLengthMax) && c.ContentType() != binding.MIMEMultipartPOSTForm {
				bodyBytes, _ := io.ReadAll(c.Request.Body)
				logger.InfoCtx(c.Request.Context(), "[Request]"+string(bodyBytes))
				_ = c.Request.Body.Close() //  must close
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			} else {
				logger.InfoCtx(c.Request.Context(), fmt.Sprintf("request %d is too long", c.Request.ContentLength))
			}
		}
		// max > 0  enable log response
		if opt.ResponseContentLengthMax > 0 {
			w := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = w
			c.Next()
			l := w.body.Len()
			if l < int(opt.ResponseContentLengthMax) {
				responseBody := w.body.String()
				logger.InfoCtx(c.Request.Context(), "[Response]"+responseBody)
			} else {
				logger.InfoCtx(c.Request.Context(), fmt.Sprintf("response %d is too long", l))
			}
		}
	}
}

// CORSMiddleware 防跨域
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", c.Request.Header.Get("Origin"))
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin,OsType,Accept-Language,X-Request-ID,Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Requested-With, Cache-Control")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
		} else {
			c.Next()
		}
	}
}
