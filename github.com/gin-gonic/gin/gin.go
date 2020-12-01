package gin

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc = gin.HandlerFunc

// Engine is the framework's instance, it contains the muxer, middleware and configuration settings.
// Create an instance of Engine, by using New() or Default()
type Engine = gin.Engine

// New returns a new blank Engine instance without any middleware attached.
// If opts is not nil, the OTel middleware will be attached.
func New(opts ...Option) *Engine {
	e := gin.New()
	if opts != nil {
		otel, err := OTel(opts...)
		if err != nil {
			panic(err)
		}
		e.Use(otel)
	}
	return e
}

// Default returns an Engine instance with the Logger and Recovery middleware already attached.
// If opts is not nil, the OTel middleware will be attached.
func Default(opts ...Option) *Engine {
	e := New(opts...)
	e.Use(Logger(), Recovery())
	return e
}

// OTel returns middleware that provides OpenTelemetry tracing and metrics to a gin web app.
func OTel(opts ...Option) (HandlerFunc, error) {
	cfg, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	return func(c *Context) {
		start := time.Now()

		ctx := cfg.propagator.Extract(c.Request.Context(), c.Request.Header)
		ctx, span := cfg.tracer.Start(ctx,
			cfg.spanNameFormatter(cfg.operationName, c),
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(semconv.NetAttributesFromHTTPRequest("tcp", c.Request)...),
			trace.WithAttributes(semconv.EndUserAttributesFromHTTPRequest(c.Request)...),
			trace.WithAttributes(
				semconv.HTTPServerAttributesFromHTTPRequest(cfg.serverName, c.FullPath(), c.Request)...,
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		statusCode := c.Writer.Status()
		spanCode, spanMsg := semconv.SpanStatusFromHTTPStatusCode(statusCode)
		span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(statusCode)...)
		span.SetStatus(spanCode, spanMsg)
		if len(c.Errors) > 0 {
			span.RecordError(fmt.Errorf(c.Errors.String()))
		}

		metricLabels := semconv.HTTPServerMetricAttributesFromHTTPRequest(cfg.serverName, c.Request)
		cfg.metricRequestCount.Add(ctx, 1, metricLabels...)
		elapsedTime := time.Since(start).Milliseconds()
		cfg.metricDuration.Record(ctx, elapsedTime, metricLabels...)
	}, nil
}
