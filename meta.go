package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/illidaris/core"
	"github.com/illidaris/logger"
	"go.uber.org/zap"
	"time"
)

type HTTPMetaData core.MetaData

const (
	HTTPStatusCode  HTTPMetaData = "statusCode"
	HTTPContentType HTTPMetaData = "contentType"
	HTTPMethod      HTTPMetaData = "httpMethod"
	HTTPPath        HTTPMetaData = "httpPath"
	HTTPQuery       HTTPMetaData = "httpQuery"
	HTTPClientIP    HTTPMetaData = "httpClientIp"
	HTTPUserAgent   HTTPMetaData = "httpUserAgent"
)

// WithTrace add trace log context
func WithTrace(c *gin.Context, birth time.Time) *gin.Context {
	// trace
	const rid = "X-Request-ID"
	traceID := c.GetHeader(rid)
	// session
	newUUID, _ := uuid.NewUUID()
	sID := newUUID.String()
	if traceID == "" {
		traceID = sID
	}
	sessionBirth := birth.UTC().UnixNano()
	// assembly trace & session
	ctx := c.Request.Context()
	ctx = logger.NewContext(ctx,
		zap.String(core.TraceID.String(), traceID),
		zap.String(core.SessionID.String(), sID),
		zap.Int64(core.SessionBirth.String(), sessionBirth))
	// instead of request
	c.Request = c.Request.WithContext(ctx)
	return c
}
