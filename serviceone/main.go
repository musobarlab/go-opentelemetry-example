package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"fmt"
	"net/http"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"github.com/musobarlab/go-opentelemetry-example/helper/tracer"
)

const (
	Name = "producer"
)

var (
	rng = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func main() {
	// zipkinURL := "http://localhost:9411/api/v2/spans"
	jaegerURL := "http://localhost:14268/api/traces"
	otelURL := "127.0.0.1:4317"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// tracerProvider, err := tracer.InitZipkinProvider(zipkinURL, "producer", "development", int64(1))
	tracerProvider, err := tracer.InitJaegerProvider(jaegerURL, "producer", "development", int64(1))
	// tracerProvider, meterProvider, err := tracer.InitDatadogProvider(ctx, otelURL, "producer", "development", int64(1))
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	meterProvider, err := tracer.InitMetricProvider(ctx, otelURL, "producer", "development", int64(1))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer func() {
		tracerProvider.Shutdown(ctx)
	}()

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tracerProvider)

	// set metric provider
	global.SetMeterProvider(meterProvider)

	meter := global.Meter(Name)
	meterRequestCounter, err := meter.Int64Counter(
		fmt.Sprintf("%s/request_counts", Name),
		instrument.WithDescription("Total number of requests received"),
	)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	http.Handle("/produce", metricMiddleware(meterRequestCounter, produceHandler()))

	log.Fatal(http.ListenAndServe(":9001", nil))
}

func metricMiddleware(requestCounter instrument.Int64Counter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		labels := []attribute.KeyValue {
			attribute.Key("req.Host").String(req.Host),
			attribute.Key("req.URL").String(req.URL.String()),
			attribute.Key("req.Method").String(req.Method),
		}
		//  random sleep to simulate latency
		var sleep int64

		switch modulus := time.Now().Unix() % 5; modulus {
		case 0:
			sleep = rng.Int63n(2000)
		case 1:
			sleep = rng.Int63n(15)
		case 2:
			sleep = rng.Int63n(917)
		case 3:
			sleep = rng.Int63n(87)
		case 4:
			sleep = rng.Int63n(1173)
		}
		time.Sleep(time.Duration(sleep) * time.Millisecond)
		ctx := req.Context()

		requestCounter.Add(ctx, 1, labels...)

		next.ServeHTTP(res, req)
	})
}

func produceHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			res.WriteHeader(http.StatusMethodNotAllowed)
			res.Write([]byte("method not allowed"))
			return
		}

		// start tracing
		// spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		// if err != nil {
		// 	log.Printf("error extract span %v\n", err)
		// }

		// span := tracer.StartSpan("produce", ext.RPCServerOption((spanCtx)))
		// defer span.Finish()
		// end tracing
		
		tr := otel.Tracer(Name)
		ctx := req.Context()

		_, span := tr.Start(ctx, "http_handler:produceHandler")
		span.SetAttributes(attribute.Key("req.Method").String(req.Method))

		defer func() {
			span.End()
		}()

		products := Products{
			Product{ID: "1", Name: "Samsung Galaxy s1"},
			Product{ID: "2", Name: "Samsung J1"},
			Product{ID: "3", Name: "Nokia 6"},
			Product{ID: "4", Name: "IPHONE 6"},
		}

		productsJSON, err := json.Marshal(products)
		if err != nil {
			log.Printf("error marshal product %v\n", err)
		}

		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(200)
		res.Write(productsJSON)
	})
}

// Product data
type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Products data
type Products []Product