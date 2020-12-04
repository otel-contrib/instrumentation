package redis

import (
	"context"
	"time"

	"github.com/go-redis/redis/extra/rediscmd"
	"github.com/go-redis/redis/v8"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

// Client is a Redis client representing a pool of zero or more underlying connections.
// It's safe for concurrent use by multiple goroutines.
type Client = redis.Client

// Hook defines redis OpenTelemetry hook interface.
type Hook = redis.Hook

type otelHook struct {
	redisOptions *Options

	tracerProvider            trace.TracerProvider
	meterProvider             metric.MeterProvider
	operationName             string
	spanNameFormatter         SpanNameFormatter
	spanNameFormatterPipeline SpanNameFormatterPipeline

	tracer         trace.Tracer
	meter          metric.Meter
	metricDuration metric.Int64ValueRecorder
}

type startTimeType struct{}

const (
	// Nil reply returned by Redis when key does not exist.
	Nil = redis.Nil
)

var (
	_ Hook = &otelHook{}

	startTimeContextKey = &startTimeType{}
)

// NewClient returns a client to the Redis Server specified by Options.
func NewClient(opt *Options, opts ...Option) *Client {
	c := redis.NewClient(opt)
	if opts != nil {
		h, err := NewOTelHook(opt, opts...)
		if err != nil {
			panic(err)
		}
		c.AddHook(h)
	}
	return c
}

// NewOTelHook returns hook that provides OpenTelemetry tracing and metrics to redis.
func NewOTelHook(opt *Options, opts ...Option) (Hook, error) {
	c, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	return &otelHook{
		redisOptions:              opt,
		tracerProvider:            c.tracerProvider,
		meterProvider:             c.meterProvider,
		operationName:             c.operationName,
		spanNameFormatter:         c.spanNameFormatter,
		spanNameFormatterPipeline: c.spanNameFormatterPipeline,
		tracer:                    c.tracer,
		meter:                     c.meter,
		metricDuration:            c.metricDuration,
	}, nil
}

func (o *otelHook) BeforeProcess(ctx context.Context, cmd Cmder) (context.Context, error) {
	start := time.Now()
	ctx = context.WithValue(ctx, startTimeContextKey, start)

	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx, nil
	}

	ctx, _ = o.tracer.Start(ctx, o.spanNameFormatter(o.operationName, cmd),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemRedis,
			semconv.DBRedisDBIndexKey.Int(o.redisOptions.DB),
			semconv.DBStatementKey.String(rediscmd.CmdString(cmd)),
			semconv.DBOperationKey.String(cmd.FullName()),
			semconv.DBConnectionStringKey.String(o.redisOptions.Addr),
			semconv.DBUserKey.String(o.redisOptions.Username),
		),
	)

	return ctx, nil
}

func (o *otelHook) AfterProcess(ctx context.Context, cmd Cmder) error {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	if err := cmd.Err(); err != nil {
		if err != Nil {
			span.RecordError(err)
		}
	}
	span.SetStatus(spanStatusFromCmder(cmd))

	start, ok := ctx.Value(startTimeContextKey).(time.Time)
	if !ok {
		start = time.Now()
	}
	elapsedTime := time.Since(start).Milliseconds()
	o.metricDuration.Record(ctx, elapsedTime)

	return nil
}

func (o *otelHook) BeforeProcessPipeline(ctx context.Context, cmds []Cmder) (context.Context, error) {
	start := time.Now()
	ctx = context.WithValue(ctx, startTimeContextKey, start)

	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx, nil
	}

	summary, cmdsString := rediscmd.CmdsString(cmds)
	ctx, _ = o.tracer.Start(ctx, o.spanNameFormatterPipeline(o.operationName, cmds),
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(
			semconv.DBSystemRedis,
			semconv.DBRedisDBIndexKey.Int(o.redisOptions.DB),
			semconv.DBStatementKey.String(cmdsString),
			semconv.DBOperationKey.String(summary),
			semconv.DBConnectionStringKey.String(o.redisOptions.Addr),
			semconv.DBUserKey.String(o.redisOptions.Username),
			LabelKeyDBRedisNumCMD.Int(len(cmds)),
		),
	)

	return ctx, nil
}

func (o *otelHook) AfterProcessPipeline(ctx context.Context, cmds []Cmder) error {
	span := trace.SpanFromContext(ctx)
	defer span.End()

	if err := cmds[0].Err(); err != nil {
		if err != Nil {
			span.RecordError(err)
		}
	}
	span.SetStatus(spanStatusFromCmder(cmds[0]))

	start, ok := ctx.Value(startTimeContextKey).(time.Time)
	if !ok {
		start = time.Now()
	}
	elapsedTime := time.Since(start).Milliseconds()
	o.metricDuration.Record(ctx, elapsedTime)

	return nil
}

func spanStatusFromCmder(cmd Cmder) (codes.Code, string) {
	if err := cmd.Err(); err != nil {
		if err != Nil {
			return codes.Error, err.Error()
		}
		return codes.Unset, err.Error()
	}
	return codes.Unset, ""
}
