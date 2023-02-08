package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"fmt"
	"net/http"
	"math/rand"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/metric/instrument"
	"github.com/musobarlab/go-opentelemetry-example/helper/tracer"
)

const (
	Name = "consumer"
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

	// tracerProvider, err := tracer.InitZipkinProvider(zipkinURL, "consumer", "development", int64(2))
	tracerProvider, err := tracer.InitJaegerProvider(jaegerURL, "consumer", "development", int64(2))
	// tracerProvider, meterProvider, err := tracer.InitDatadogProvider(ctx, otelURL, "consumer", "development", int64(2))
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	meterProvider, err := tracer.InitMetricProvider(ctx, otelURL, "consumer", "development", int64(2))
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

	http.Handle("/consume", metricMiddleware(meterRequestCounter, consumeHandler()))

	log.Fatal(http.ListenAndServe(":9002", nil))
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

func consumeHandler() http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {

		if req.Method != "POST" {
			res.WriteHeader(http.StatusMethodNotAllowed)
			res.Write([]byte("method not allowed"))
			return
		}

		// start tracing
		// spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		// if err != nil {
		// 	log.Printf("error extract span %v\n", err)
		// }

		// span := tracer.StartSpan("consume", ext.RPCServerOption((spanCtx)))
		// defer span.Finish()
		// end tracing

		tr := otel.Tracer(Name)
		ctx := req.Context()

		_, span := tr.Start(ctx, "http_handler:consumeHandler")
		span.SetAttributes(attribute.Key("req.Method").String(req.Method))

		defer func() {
			span.End()
		}()

		search, ok := req.URL.Query()["search"]

		if !ok || len(search) < 1 {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("search param required"))
			return
		}

		var (
			products Products
			product  Product
		)
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&products)

		if err != nil {
			res.WriteHeader(http.StatusBadRequest)
			res.Write([]byte("error encode body"))
			return
		}

		// find product by search
		for _, v := range products {
			if strings.Contains(v.Name, search[0]) {
				product = v
				break
			}
		}

		productJSON, err := json.Marshal(product)
		if err != nil {
			log.Printf("error marshal product %v\n", err)
		}

		res.Header().Add("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		res.Write(productJSON)
	})
}

// Product data
type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Products data
type Products []Product