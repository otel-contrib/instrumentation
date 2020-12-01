package gin

const (
	defaultInstrumentationName = "github.com/otel-contrib/instrumentation/github.com/gin-gonic/gin"
	defaultServerName          = "gin"
	defaultOperationName       = "gin"
)

// Metrics semantic conventions
const (
	metricHTTPServerDuration     = "http.server.duration"      // Incoming end to end duration, milliseconds
	metricHTTPServerRequestCount = "http.server.request_count" // Incoming request count total
)
