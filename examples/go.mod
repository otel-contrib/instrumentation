module github.com/otel-contrib/instrumentation/examples

go 1.14

require (
	github.com/otel-contrib/instrumentation v0.0.0
	go.opentelemetry.io/otel v0.14.0
	go.opentelemetry.io/otel/exporters/stdout v0.14.0
	go.opentelemetry.io/otel/sdk v0.14.0
)

replace github.com/otel-contrib/instrumentation => ../
