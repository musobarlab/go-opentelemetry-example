package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"fmt"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"github.com/musobarlab/go-opentelemetry-example/helper/tracer"
)

const (
	Name = "consumer"
)

func main() {
	// zipkinURL := "http://localhost:9411/api/v2/spans"
	// jaegerURL := "http://localhost:14268/api/traces"
	otelURL := "127.0.0.1:4317"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// tracerProvider, err := tracer.InitZipkinProvider(zipkinURL, "consumer", "development", int64(2))
	// tracerProvider, err := tracer.InitJaegerProvider(jaegerURL, "consumer", "development", int64(2))
	tracerProvider, err := tracer.InitDatadogProvider(ctx, otelURL, "consumer", "development", int64(2))
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

	http.Handle("/consume", consumeHandler())

	log.Fatal(http.ListenAndServe(":9002", nil))
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