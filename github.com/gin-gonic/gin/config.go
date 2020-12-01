package gin

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/unit"
)

type config struct {
	tracerProvider    trace.TracerProvider
	meterProvider     metric.MeterProvider
	propagator        propagation.TextMapPropagator
	serverName        string
	operationName     string
	spanNameFormatter SpanNameFormatter

	tracer             trace.Tracer
	meter              metric.Meter
	metricDuration     metric.Int64ValueRecorder
	metricRequestCount metric.Int64Counter
}

// Option applies a configuration to the given config.
type Option interface {
	Apply(*config)
}

// OptionFunc provides a convenience wrapper for simple Options that can be represented as functions.
type OptionFunc func(c *config)

// Apply will apply the option to the config.
func (o OptionFunc) Apply(c *config) {
	o(c)
}

// SpanNameFormatter creates a custom span name from the operation and context object.
type SpanNameFormatter func(operation string, c *gin.Context) string

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(tp trace.TracerProvider) Option {
	return OptionFunc(func(c *config) {
		c.tracerProvider = tp
	})
}

// WithMeterProvider specifies a meter provider to use for creating a meter.
// If none is specified, the global provider is used.
func WithMeterProvider(mp metric.MeterProvider) Option {
	return OptionFunc(func(cfg *config) {
		cfg.meterProvider = mp
	})
}

// WithPropagators specifies a propagators.
// If none is specified, the global propagator is used.
func WithPropagators(ps propagation.TextMapPropagator) Option {
	return OptionFunc(func(c *config) {
		c.propagator = ps
	})
}

// WithServerName specifies a server name.
// If none is specified, the default server name is used.
func WithServerName(name string) Option {
	return OptionFunc(func(c *config) {
		c.serverName = name
	})
}

// WithOperationName specifies a operation name.
// If none is specified, the default operation name is used.
func WithOperationName(name string) Option {
	return OptionFunc(func(c *config) {
		c.operationName = name
	})
}

// WithSpanNameFormatter specifies a formatter to used to format span names.
// If none is specified, the default SpanNameFormatter is used.
func WithSpanNameFormatter(f SpanNameFormatter) Option {
	return OptionFunc(func(c *config) {
		c.spanNameFormatter = f
	})
}

func newConfig(opts ...Option) (*config, error) {
	var err error
	c := &config{
		tracerProvider:    otel.GetTracerProvider(),
		meterProvider:     otel.GetMeterProvider(),
		propagator:        otel.GetTextMapPropagator(),
		serverName:        defaultServerName,
		operationName:     defaultOperationName,
		spanNameFormatter: defaultSpanNameFormatter,
	}
	for _, opt := range opts {
		opt.Apply(c)
	}

	c.tracer = c.tracerProvider.Tracer(
		defaultInstrumentationName,
		trace.WithInstrumentationVersion(contrib.SemVersion()),
	)
	c.meter = c.meterProvider.Meter(
		defaultInstrumentationName,
		metric.WithInstrumentationVersion(contrib.SemVersion()),
	)

	c.metricDuration, err = c.meter.NewInt64ValueRecorder(
		metricHTTPServerDuration,
		metric.WithDescription("request response time in microseconds"),
		metric.WithUnit(unit.Milliseconds),
	)
	if err != nil {
		return nil, err
	}
	c.metricRequestCount, err = c.meter.NewInt64Counter(
		metricHTTPServerRequestCount,
		metric.WithDescription("request count"),
		metric.WithUnit(unit.Dimensionless),
	)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func defaultSpanNameFormatter(operation string, c *gin.Context) string {
	return c.FullPath()
}
