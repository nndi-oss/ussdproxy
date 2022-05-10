package telemetry

import (
	"fmt"
	"log"
	"net/http"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/sdk/export/metric/aggregation"
	"go.opentelemetry.io/otel/sdk/metric/aggregator/histogram"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

type Telemetry struct {
	exporter *prometheus.Exporter
	meter    metric.Meter
	measures map[string]*syncfloat64.Histogram
	counters map[string]*syncint64.Counter
	attrs    []attribute.KeyValue
}

func New(bindAddress string) *Telemetry {
	config := prometheus.Config{
		DefaultHistogramBoundaries: []float64{1, 2, 5, 10, 20, 50},
	}

	ctrl := controller.New(
		processor.NewFactory(
			selector.NewWithHistogramDistribution(
				histogram.WithExplicitBoundaries(config.DefaultHistogramBoundaries),
			),
			aggregation.CumulativeTemporalitySelector(),
			processor.WithMemory(true),
		),
	)

	exporter, err := prometheus.New(config, ctrl)
	if err != nil {
		log.Panicf("failed to initialize prometheus exporter %v", err)
	}
	global.SetMeterProvider(exporter.MeterProvider())

	if err != nil {
		log.Panicf("failed to initialize metric stdout exporter %v", err)
	}

	// batcher := defaultkeys.New(selector, sdkmetric.NewDefaultLabelEncoder(), false)
	// pusher := push.New(batcher, exporter, time.Second)
	// pusher.Start()

	go func() {
		_ = http.ListenAndServe(bindAddress, exporter)
	}()

	fmt.Println("Prometheus server running on ", bindAddress)

	meter := global.MeterProvider().Meter("ussdproxy")

	m := make(map[string]*syncfloat64.Histogram)
	c := make(map[string]*syncint64.Counter)

	commonAttrs := []attribute.KeyValue{
		attribute.String("server-name", "ussdproxy"),
	}

	return &Telemetry{exporter, meter, m, c, commonAttrs}
}

// AddMeasure to metrics
func (t *Telemetry) AddMeasure(key string) {
	if met, err := t.meter.SyncFloat64().Histogram(key); err == nil {
		t.measures[key] = &met
	}
}

// AddCounter to metrics collection
func (t *Telemetry) AddCounter(key string) {
	if met, err := t.meter.SyncInt64().Counter(key); err == nil {
		t.counters[key] = &met
	}
}

// // NewTiming creates a new timing metric and returns a done function
// func (t *Telemetry) NewTiming(key string) func() {
// 	// record the start time
// 	st := time.Now()

// 	return func() {
// 		dur := time.Now().Sub(st).Nanoseconds()
// 		handler := t.measures[key].AcquireHandle(nil)
// 		defer handler.Release()

// 		t.meter.RecordBatch(
// 			context.Background(),
// 			nil,
// 			t.measures[key].Measurement(float64(dur)),
// 		)
// 	}
// }
