package main

import (
	"context"
	"log"
	"time"

	"github.com/otel-contrib/instrumentation/github.com/go-redis/redis"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	log.Println("start ...")

	exporter, err := stdout.NewExporter([]stdout.Option{stdout.WithPrettyPrint()}...)
	if err != nil {
		log.Fatal(err)
	}

	ssp := trace.NewSimpleSpanProcessor(exporter)
	tp := trace.NewTracerProvider(trace.WithSpanProcessor(ssp))
	pusher := push.New(basic.New(simple.NewWithHistogramDistribution([]float64{100, 200, 500}), exporter), exporter)
	pusher.Start()
	defer pusher.Stop()

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(pusher.MeterProvider())

	opt := &redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	c := redis.NewClient(opt, redis.WithOperationName("name string"))

	trace := otel.Tracer("example/redis")
	ctx, span := trace.Start(context.Background(), "example")

	if err := c.Set(ctx, "key", "value", time.Minute).Err(); err != nil {
		panic(err)
	}

	span.End()
	time.Sleep(10 * time.Second)

	log.Println("end")
}
