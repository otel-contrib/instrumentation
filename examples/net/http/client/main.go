package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/otel-contrib/instrumentation/net/http"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/propagation"
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
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	trace := otel.Tracer("example/client")
	ctx, span := trace.Start(context.Background(), "example")

	resp, err := http.GetWithContext(ctx, "http://baidu.com/")
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(body))

	span.End()
	time.Sleep(10 * time.Second)

	log.Println("end")
}
