package gorm

const (
	defaultInstrumentationName = "github.com/otel-contrib/instrumentation/gorm.io/gorm"
	defaultOperationName       = "gorm"
)

const (
	hookNameBefore = "otel:before"
	hookNameAfter  = "otel:after"
)

// Metrics semantic conventions
const (
	metricRedisClientDuration = "db.gorm.client.duration" // process time, milliseconds
)
