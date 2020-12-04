package gorm

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
	dialectmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config GORM config
type Config = gorm.Config

// DB GORM DB definition
type DB = gorm.DB

type otelHook struct {
	tracerProvider    trace.TracerProvider
	meterProvider     metric.MeterProvider
	operationName     string
	spanNameFormatter SpanNameFormatter
	attrs             []label.KeyValue

	tracer         trace.Tracer
	meter          metric.Meter
	metricDuration metric.Int64ValueRecorder
}

// OTelHook defines gorm OpenTelemetry hook interface.
type OTelHook interface {
	Before() func(*DB)
	After() func(*DB)
	Register(db *DB) error
}

type startTimeType struct{}

var (
	_ OTelHook = &otelHook{}

	startTimeContextKey = &startTimeType{}
)

// Open initialize db session based on dialector.
func Open(dialector Dialector, config *Config, opts ...Option) (*DB, error) {
	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}

	h, err := NewOTelHook(db, opts...)
	if err != nil {
		return nil, err
	}

	if err := h.Register(db); err != nil {
		return nil, err
	}
	return db, err
}

// NewOTelHook returns hook that provides OpenTelemetry tracing and metrics to gorm.
func NewOTelHook(db *DB, opts ...Option) (OTelHook, error) {
	c, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	h := &otelHook{
		tracerProvider:    c.tracerProvider,
		meterProvider:     c.meterProvider,
		operationName:     c.operationName,
		spanNameFormatter: c.spanNameFormatter,
		tracer:            c.tracer,
		meter:             c.meter,
		metricDuration:    c.metricDuration,
	}

	switch dial := db.Dialector.(type) {
	case *dialectmysql.Dialector:
		cfg, err := mysql.ParseDSN(dial.DSN)
		if err != nil {
			return nil, err
		}
		cfg.Passwd = ""

		netPeerIP, netPeerPort, err := parseAddr(cfg.Addr)
		if err != nil {
			return nil, err
		}
		h.attrs = append(h.attrs, semconv.DBSystemMySQL)
		h.attrs = append(h.attrs, semconv.DBConnectionStringKey.String(cfg.FormatDSN()))
		h.attrs = append(h.attrs, semconv.DBUserKey.String(cfg.User))
		h.attrs = append(h.attrs, semconv.NetPeerIPKey.String(netPeerIP))
		h.attrs = append(h.attrs, semconv.NetPeerPortKey.Int(netPeerPort))
		h.attrs = append(h.attrs, parseNetTransport(strings.ToLower(cfg.Net)))
		h.attrs = append(h.attrs, semconv.DBNameKey.String(cfg.DBName))
	default:
		return nil, fmt.Errorf("unsupported dialector type")
	}

	return h, nil
}

func (o *otelHook) Before() func(*DB) {
	return func(db *DB) {
		start := time.Now()
		ctx := context.WithValue(db.Statement.Context, startTimeContextKey, start)
		if !trace.SpanFromContext(ctx).IsRecording() {
			return
		}

		ctx, _ = o.tracer.Start(ctx, o.operationName, trace.WithSpanKind(trace.SpanKindClient),
			trace.WithAttributes(o.attrs...),
		)
		db.Statement.Context = ctx
	}
}

func (o *otelHook) After() func(*DB) {
	return func(db *DB) {
		ctx := db.Statement.Context
		span := trace.SpanFromContext(ctx)
		defer span.End()

		span.SetName(o.spanNameFormatter(o.operationName, db))
		sql := db.Statement.SQL.String()
		span.SetAttributes(
			semconv.DBStatementKey.String(db.Statement.Explain(sql, db.Statement.Vars...)),
			semconv.DBOperationKey.String(parseOperation(sql)),
		)
		if err := db.Error; err != nil {
			if err != ErrRecordNotFound {
				span.RecordError(err)
			}
		}
		span.SetStatus(spanStatusFromDB(db))

		start, ok := ctx.Value(startTimeContextKey).(time.Time)
		if !ok {
			start = time.Now()
		}
		elapsedTime := time.Since(start).Milliseconds()
		o.metricDuration.Record(ctx, elapsedTime)

		db.Statement.Context = ctx
	}
}

func (o *otelHook) Register(db *DB) error {
	if err := db.Callback().Create().Before("*").Register(hookNameBefore, o.Before()); err != nil {
		return err
	}
	if err := db.Callback().Create().After("*").Register(hookNameAfter, o.After()); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("*").Register(hookNameBefore, o.Before()); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("*").Register(hookNameAfter, o.After()); err != nil {
		return err
	}

	if err := db.Callback().Query().Before("*").Register(hookNameBefore, o.Before()); err != nil {
		return err
	}
	if err := db.Callback().Query().After("*").Register(hookNameAfter, o.After()); err != nil {
		return err
	}

	if err := db.Callback().Raw().Before("*").Register(hookNameBefore, o.Before()); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("*").Register(hookNameAfter, o.After()); err != nil {
		return err
	}

	if err := db.Callback().Row().Before("*").Register(hookNameBefore, o.Before()); err != nil {
		return err
	}
	if err := db.Callback().Row().After("*").Register(hookNameAfter, o.After()); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("*").Register(hookNameBefore, o.Before()); err != nil {
		return err
	}
	if err := db.Callback().Update().After("*").Register(hookNameAfter, o.After()); err != nil {
		return err
	}

	return nil
}

func parseAddr(addr string) (string, int, error) {
	i := strings.Index(addr, ":")
	if i < 0 {
		return addr, 0, nil
	}
	ip := addr[:i]
	port, err := strconv.Atoi(addr[i+1:])
	if err != nil {
		return "", 0, err
	}
	return ip, port, nil
}

func parseNetTransport(network string) label.KeyValue {
	switch network {
	case "tcp", "tcp4", "tcp6":
		return semconv.NetTransportTCP
	case "udp", "udp4", "udp6":
		return semconv.NetTransportUDP
	case "ip", "ip4", "ip6":
		return semconv.NetTransportIP
	case "unix", "unixgram", "unixpacket":
		return semconv.NetTransportUnix
	default:
		return semconv.NetTransportOther
	}
}

func parseOperation(sql string) string {
	i := strings.Index(sql, " ")
	if i < 0 {
		return sql
	}
	return sql[:i]
}

func spanStatusFromDB(db *DB) (codes.Code, string) {
	if err := db.Error; err != nil {
		if err != ErrRecordNotFound {
			return codes.Error, err.Error()
		}
		return codes.Unset, err.Error()
	}
	return codes.Unset, ""
}
