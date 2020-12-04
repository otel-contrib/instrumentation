package main

import (
	"context"
	"log"
	"time"

	"github.com/otel-contrib/instrumentation/gorm.io/gorm"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/trace"
	"gorm.io/driver/mysql"
)

type user struct {
	gorm.Model
	Name string
}

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

	dsn := "root:password@tcp(127.0.0.1:3306)/test?parseTime=true&loc=Asia%2FShanghai"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	trace := otel.Tracer("example/gorm")
	ctx, span := trace.Start(context.Background(), "example")

	if err := db.AutoMigrate(&user{}); err != nil {
		log.Fatal(err)
	}
	if err := db.WithContext(ctx).Create(&user{Name: "test"}).Error; err != nil {
		log.Fatal(err)
	}

	us := []user{}
	if err := db.WithContext(ctx).Find(&us).Error; err != nil {
		log.Fatal(err)
	}

	u1 := us[0]
	u1.Name = "my test"
	if err := db.WithContext(ctx).Save(&u1).Error; err != nil {
		log.Fatal(err)
	}

	if err := db.WithContext(ctx).Delete(&user{}, "name = ?", "test").Error; err != nil {
		log.Fatal(err)
	}

	span.End()
	time.Sleep(10 * time.Second)

	log.Println("end")
}
