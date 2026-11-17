package requestlog

import (
	"log/slog"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func New(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		requestID := requestid.Get(c)

		c.Next()

		log.InfoContext(c.Request.Context(), "HTTP request",
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(started)),
			slog.String("request_id", requestID),
		)

		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				log.Error("request error", slog.String("error", err.Error()))
			}
		}
	}
}
