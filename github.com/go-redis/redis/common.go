package redis

import "go.opentelemetry.io/otel/label"

const (
	defaultInstrumentationName = "github.com/otel-contrib/instrumentation/github.com/go-redis/redis"
	defaultOperationName       = "redis"
)

// Semantic conventions for attribute keys for redis.
const (
	LabelKeyDBRedisNumCMD = label.Key("db.redis.num_cmd")
)

// Metrics semantic conventions
const (
	metricRedisClientDuration = "db.redis.client.duration" // process time, milliseconds
)
