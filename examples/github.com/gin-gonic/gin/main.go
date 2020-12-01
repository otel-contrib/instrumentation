package main

import (
	"log"
	"net/http"

	"github.com/otel-contrib/instrumentation/github.com/gin-gonic/gin"
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
	otel.SetTextMapPropagator(propagation.Baggage{})

	r := gin.Default(
		gin.WithServerName("hello"),
	)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Run()

	log.Println("end")
}
