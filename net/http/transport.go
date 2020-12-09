package http

import (
	"net/http"
	"time"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

// DefaultTransport is the default implementation of Transport and is used by DefaultClient.
var DefaultTransport = http.DefaultTransport

// Transport is an implementation of RoundTripper that supports HTTP, HTTPS, and HTTP proxies.
type Transport = http.Transport

type otelTransport struct {
	rt http.RoundTripper

	tracerProvider    trace.TracerProvider
	meterProvider     metric.MeterProvider
	propagator        propagation.TextMapPropagator
	operationName     string
	spanNameFormatter SpanNameFormatter

	tracer                   trace.Tracer
	meter                    metric.Meter
	metricDuration           metric.Int64ValueRecorder
	metricRequestCount       metric.Int64Counter
	metricRequestFailedCount metric.Int64Counter
}

var _ RoundTripper = &otelTransport{}

// NewOTelTransport wraps the provided RoundTripper with one that starts a span and
// injects the span context into the outbound request headers.
// If none is specified, the DefaultTransport is used.
func NewOTelTransport(rt RoundTripper, opts ...Option) (RoundTripper, error) {
	c, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}
	if rt == nil {
		rt = DefaultTransport
	}

	o := &otelTransport{
		rt:                       rt,
		tracerProvider:           c.tracerProvider,
		meterProvider:            c.meterProvider,
		propagator:               c.propagator,
		operationName:            c.operationName,
		spanNameFormatter:        c.spanNameFormatter,
		tracer:                   c.tracer,
		meter:                    c.meter,
		metricDuration:           c.metricClientDuration,
		metricRequestCount:       c.metricClientRequestCount,
		metricRequestFailedCount: c.metricClientRequestFailedCount,
	}

	return o, nil
}

func (o *otelTransport) RoundTrip(req *Request) (*Response, error) {
	start := time.Now()

	netAttrs := semconv.NetAttributesFromHTTPRequest("tcp", req)
	httpClientAttrs := semconv.HTTPClientAttributesFromHTTPRequest(req)
	attrs := append(netAttrs, httpClientAttrs...)
	metricLabels := attrs
	ctx, span := o.tracer.Start(req.Context(), o.spanNameFormatter(o.operationName, req),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
	defer span.End()

	req = req.WithContext(ctx)
	o.propagator.Inject(ctx, req.Header)

	resp, err := o.rt.RoundTrip(req)
	if err != nil {
		span.RecordError(err)
		o.metricRequestFailedCount.Add(ctx, 1, metricLabels...)
	} else {
		httpAttributes := semconv.HTTPAttributesFromHTTPStatusCode(resp.StatusCode)
		span.SetAttributes(httpAttributes...)
		span.SetStatus(semconv.SpanStatusFromHTTPStatusCode(resp.StatusCode))
		metricLabels = append(attrs, httpAttributes...)
	}

	o.metricRequestCount.Add(ctx, 1, metricLabels...)
	elapsedTime := time.Since(start).Milliseconds()
	o.metricDuration.Record(ctx, elapsedTime, metricLabels...)

	return resp, err
}
