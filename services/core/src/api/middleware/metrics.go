package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("http-middleware")

	requestCounter, _ = meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total number of HTTP requests"),
	)
	durationHistogram, _ = meter.Float64Histogram("http_request_duration_seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
	)
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()

		attrs := metric.WithAttributes(
			attribute.String("method", c.Request.Method),
			attribute.String("path", path),
			attribute.String("status", status),
		)

		requestCounter.Add(c.Request.Context(), 1, attrs)
		durationHistogram.Record(c.Request.Context(), duration, attrs)
	}
}
