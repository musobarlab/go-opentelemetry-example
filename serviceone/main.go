package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"github.com/musobarlab/go-opentelemetry-example/helper/tracer"
)

const (
	Name = "producer"
)

func main() {
	// zipkinURL := "http://localhost:9411/api/v2/spans"
	// jaegerURL := "http://localhost:14268/api/traces"
	otelURL := "127.0.0.1:4317"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// tracerProvider, err := tracer.InitZipkinProvider(zipkinURL, "producer", "development", int64(1))
	// tracerProvider, err := tracer.InitJaegerProvider(jaegerURL, "producer", "development", int64(1))
	tracerProvider, err := tracer.InitDatadogProvider(ctx, otelURL, "producer", "development", int64(1))
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

	http.Handle("/produce", produceHandler())

	log.Fatal(http.ListenAndServe(":9001", nil))
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