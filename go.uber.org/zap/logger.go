package zap

import (
	"context"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// A Logger provides fast, leveled, structured logging. All methods are safe
// for concurrent use.
//
// The Logger is designed for contexts in which every microsecond and every
// allocation matters, so its API intentionally favors performance and type
// safety over brevity. For most applications, the SugaredLogger strikes a
// better balance between performance and ergonomics.
type Logger struct {
	log *zap.Logger
}

// New constructs a new Logger from the provided zapcore.Core and Options. If
// the passed zapcore.Core is nil, it falls back to using a no-op
// implementation.
//
// This is the most flexible way to construct a Logger, but also the most
// verbose. For typical use cases, the highly-opinionated presets
// (NewProduction, NewDevelopment, and NewExample) or the Config struct are
// more convenient.
//
// For sample code, see the package-level AdvancedConfiguration example.
func New(core zapcore.Core, options ...Option) *Logger {
	log := zap.New(core, options...)
	return &Logger{log: log.WithOptions(AddCallerSkip(1))}
}

// NewNop returns a no-op Logger. It never writes out logs or internal errors,
// and it never runs user-defined hooks.
//
// Using WithOptions to replace the Core or error output of a no-op Logger can
// re-enable logging.
func NewNop() *Logger {
	return &Logger{log: zap.NewNop()}
}

// NewProduction builds a sensible production Logger that writes InfoLevel and
// above logs to standard error as JSON.
//
// It's a shortcut for NewProductionConfig().Build(...Option).
func NewProduction(options ...Option) (*Logger, error) {
	log, err := zap.NewProduction(options...)
	if err != nil {
		return nil, err
	}
	return &Logger{log: log.WithOptions(AddCallerSkip(1))}, nil
}

// NewDevelopment builds a development Logger that writes DebugLevel and above
// logs to standard error in a human-friendly format.
//
// It's a shortcut for NewDevelopmentConfig().Build(...Option).
func NewDevelopment(options ...Option) (*Logger, error) {
	log, err := zap.NewDevelopment(options...)
	if err != nil {
		return nil, err
	}
	return &Logger{log: log.WithOptions(AddCallerSkip(1))}, nil
}

// NewExample builds a Logger that's designed for use in zap's testable
// examples. It writes DebugLevel and above logs to standard out as JSON, but
// omits the timestamp and calling function to keep example output
// short and deterministic.
func NewExample(options ...Option) *Logger {
	return &Logger{log: zap.NewExample(options...)}
}

// Sugar wraps the Logger to provide a more ergonomic, but slightly slower,
// API. Sugaring a Logger is quite inexpensive, so it's reasonable for a
// single application to use both Loggers and SugaredLoggers, converting
// between them on the boundaries of performance-sensitive code.
func (log *Logger) Sugar() *SugaredLogger {
	return log.log.Sugar()
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (log *Logger) Named(s string) *Logger {
	return &Logger{log: log.log.Named(s)}
}

// WithOptions clones the current Logger, applies the supplied Options, and
// returns the resulting Logger. It's safe to use concurrently.
func (log *Logger) WithOptions(opts ...Option) *Logger {
	return &Logger{log: log.log.WithOptions(opts...)}
}

// With creates a child logger and adds structured context to it. Fields added
// to the child don't affect the parent, and vice versa.
func (log *Logger) With(fields ...Field) *Logger {
	return &Logger{log: log.log.With(fields...)}
}

// Check returns a CheckedEntry if logging a message at the specified level
// is enabled. It's a completely optional optimization; in high-performance
// applications, Check can help avoid allocating a slice to hold fields.
func (log *Logger) Check(lvl zapcore.Level, msg string) *zapcore.CheckedEntry {
	return log.log.Check(lvl, msg)
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) Debug(msg string, fields ...Field) {
	log.log.Debug(msg, fields...)
}

// DebugWithContext logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) DebugWithContext(ctx context.Context, msg string, fields ...Field) {
	log.log.Debug(msg, append(fieldsFromContext(ctx), fields...)...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) Info(msg string, fields ...Field) {
	log.log.Info(msg, fields...)
}

// InfoWithContext logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) InfoWithContext(ctx context.Context, msg string, fields ...Field) {
	log.log.Info(msg, append(fieldsFromContext(ctx), fields...)...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) Warn(msg string, fields ...Field) {
	log.log.Warn(msg, fields...)
}

// WarnWithContext logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) WarnWithContext(ctx context.Context, msg string, fields ...Field) {
	log.log.Warn(msg, append(fieldsFromContext(ctx), fields...)...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) Error(msg string, fields ...Field) {
	log.log.Error(msg, fields...)
}

// ErrorWithContext logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func (log *Logger) ErrorWithContext(ctx context.Context, msg string, fields ...Field) {
	log.log.Error(msg, append(fieldsFromContext(ctx), fields...)...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func (log *Logger) DPanic(msg string, fields ...Field) {
	log.log.DPanic(msg, fields...)
}

// DPanicWithContext logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func (log *Logger) DPanicWithContext(ctx context.Context, msg string, fields ...Field) {
	log.log.DPanic(msg, append(fieldsFromContext(ctx), fields...)...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (log *Logger) Panic(msg string, fields ...Field) {
	log.log.Panic(msg, fields...)
}

// PanicWithContext logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func (log *Logger) PanicWithContext(ctx context.Context, msg string, fields ...Field) {
	log.log.Panic(msg, append(fieldsFromContext(ctx), fields...)...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (log *Logger) Fatal(msg string, fields ...Field) {
	log.log.Fatal(msg, fields...)
}

// FatalWithContext logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func (log *Logger) FatalWithContext(ctx context.Context, msg string, fields ...Field) {
	log.log.Fatal(msg, append(fieldsFromContext(ctx), fields...)...)
}

// Sync calls the underlying Core's Sync method, flushing any buffered log
// entries. Applications should take care to call Sync before exiting.
func (log *Logger) Sync() error {
	return log.log.Sync()
}

// Core returns the Logger's underlying zapcore.Core.
func (log *Logger) Core() zapcore.Core {
	return log.log.Core()
}

func fieldsFromContext(ctx context.Context) []Field {
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return nil
	}
	sc := span.SpanContext()
	traceID := zap.String(logTraceID, sc.TraceID.String())
	spanID := zap.String(logSpanID, sc.SpanID.String())
	return []Field{traceID, spanID}
}
