package gorm

const (
	defaultInstrumentationName = "github.com/otel-contrib/instrumentation/gorm.io/gorm"
	defaultOperationName       = "gorm"
)

const (
	pluginName       = "gorm:otel"
	pluginNameBefore = "gorm:otel:before"
	pluginNameAfter  = "gorm:otel:after"
)

// Metrics semantic conventions
const (
	metricRedisClientDuration = "db.gorm.client.duration" // process time, milliseconds
)
