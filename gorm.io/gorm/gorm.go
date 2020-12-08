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

// Plugin GORM plugin interface
type Plugin = gorm.Plugin

type otelPlugin struct {
	tracerProvider    trace.TracerProvider
	meterProvider     metric.MeterProvider
	operationName     string
	spanNameFormatter SpanNameFormatter
	attrs             []label.KeyValue

	tracer         trace.Tracer
	meter          metric.Meter
	metricDuration metric.Int64ValueRecorder
}

type startTimeType struct{}

var (
	_ Plugin = &otelPlugin{}

	startTimeContextKey = &startTimeType{}
)

// Open initialize db session based on dialector.
func Open(dialector Dialector, config *Config, opts ...Option) (*DB, error) {
	db, err := gorm.Open(dialector, config)
	if err != nil {
		return nil, err
	}
	if opts == nil {
		return db, nil
	}

	plugin, err := NewOTelPlugin(db, opts...)
	if err != nil {
		return nil, err
	}
	if err := db.Use(plugin); err != nil {
		return nil, err
	}

	return db, err
}

// NewOTelPlugin returns plugin that provides OpenTelemetry tracing and metrics to gorm.
func NewOTelPlugin(db *DB, opts ...Option) (Plugin, error) {
	c, err := newConfig(opts...)
	if err != nil {
		return nil, err
	}

	p := &otelPlugin{
		tracerProvider:    c.tracerProvider,
		meterProvider:     c.meterProvider,
		operationName:     c.operationName,
		spanNameFormatter: c.spanNameFormatter,
		tracer:            c.tracer,
		meter:             c.meter,
		metricDuration:    c.metricDuration,
	}

	switch dialector := db.Dialector.(type) {
	case *dialectmysql.Dialector:
		cfg, err := mysql.ParseDSN(dialector.DSN)
		if err != nil {
			return nil, err
		}
		cfg.Passwd = ""

		netPeerIP, netPeerPort, err := parseAddr(cfg.Addr)
		if err != nil {
			return nil, err
		}
		p.attrs = append(p.attrs, semconv.DBSystemMySQL)
		p.attrs = append(p.attrs, semconv.DBConnectionStringKey.String(cfg.FormatDSN()))
		p.attrs = append(p.attrs, semconv.DBUserKey.String(cfg.User))
		p.attrs = append(p.attrs, semconv.NetPeerIPKey.String(netPeerIP))
		p.attrs = append(p.attrs, semconv.NetPeerPortKey.Int(netPeerPort))
		p.attrs = append(p.attrs, parseNetTransport(cfg.Net))
		p.attrs = append(p.attrs, semconv.DBNameKey.String(cfg.DBName))
	default:
		return nil, fmt.Errorf("unsupported dialector type")
	}

	return p, nil
}

func (o *otelPlugin) Name() string {
	return pluginName
}

func (o *otelPlugin) Initialize(db *DB) error {
	before := o.before()
	after := o.after()

	if err := db.Callback().Create().Before("*").Register(pluginNameBefore, before); err != nil {
		return err
	}
	if err := db.Callback().Create().After("*").Register(pluginNameAfter, after); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("*").Register(pluginNameBefore, before); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("*").Register(pluginNameAfter, after); err != nil {
		return err
	}

	if err := db.Callback().Query().Before("*").Register(pluginNameBefore, before); err != nil {
		return err
	}
	if err := db.Callback().Query().After("*").Register(pluginNameAfter, after); err != nil {
		return err
	}

	if err := db.Callback().Raw().Before("*").Register(pluginNameBefore, before); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("*").Register(pluginNameAfter, after); err != nil {
		return err
	}

	if err := db.Callback().Row().Before("*").Register(pluginNameBefore, before); err != nil {
		return err
	}
	if err := db.Callback().Row().After("*").Register(pluginNameAfter, after); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("*").Register(pluginNameBefore, before); err != nil {
		return err
	}
	if err := db.Callback().Update().After("*").Register(pluginNameAfter, after); err != nil {
		return err
	}

	return nil
}

func (o *otelPlugin) before() func(*DB) {
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

func (o *otelPlugin) after() func(*DB) {
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
	switch strings.ToLower(network) {
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
