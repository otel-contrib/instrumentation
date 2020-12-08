package http

const (
	defaultInstrumentationName = "github.com/otel-contrib/instrumentation/net/http"
	defaultOperationName       = "http"
)

// Metrics semantic conventions
const (
	metricHTTPClientDuration           = "http.client.duration"             // process time, milliseconds
	metricHTTPClientRequestCount       = "http.client.request_count"        // incoming request count total
	metricHTTPClientRequestFailedCount = "http.client.request_failed_count" // incoming request failed count total
)
