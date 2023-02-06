package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"github.com/musobarlab/go-opentelemetry-example/helper"
	"github.com/musobarlab/go-opentelemetry-example/helper/tracer"
)

const (
	Name = "client-service"
)

func main() {
	if len(os.Args) != 2 {
		panic("ERROR: Expecting one argument")
	}

	// zipkinURL := "http://localhost:9411/api/v2/spans"
	// jaegerURL := "http://localhost:14268/api/traces"
	otelURL := "127.0.0.1:4317"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// tracerProvider, err := tracer.InitZipkinProvider(zipkinURL, "client-service", "development", int64(3))
	// tracerProvider, err := tracer.InitJaegerProvider(jaegerURL, "client-service", "development", int64(3))
	tracerProvider, err := tracer.InitDatadogProvider(ctx, otelURL, "client-service", "development", int64(3))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("init provider succeed")

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tracerProvider)

	productName := os.Args[1]

	defer func() {
		tracerProvider.Shutdown(ctx)
	}()

	// span := tracer.StartSpan("search-product")
	// span.SetTag("search-to", productName)
	// defer span.Finish()

	// ctx := opentracing.ContextWithSpan(context.Background(), span)

	tr := otel.Tracer(Name)
	
	// Root span/span 1
	ctx, span := tr.Start(ctx, "client-service:main")
	span.SetAttributes(attribute.Key("search-to").String(productName))

	defer func() {
		span.End()
	}()

	// span 2
	data := produce(ctx)

	// span 3
	product := consume(ctx, data, productName)
	fmt.Println("product id : ", product.ID)
	fmt.Println("product name : ", product.Name)
	
	// span 4
	productAndID := formatString(ctx, fmt.Sprintf("%s:%s", product.ID, product.Name))
	fmt.Println(productAndID)

	// span 5
	printHello(ctx, "Opentelemetry")

}

func formatString(ctx context.Context, helloTo string) string {
	tr := otel.Tracer(Name)

	_, span := tr.Start(ctx, "client-service:formatString")
	span.SetAttributes(attribute.Key("helloTo").String(helloTo))

	defer func() {
		span.End()
	}()

	helloStr := fmt.Sprintf("Hello, %s!", helloTo)

	span.SetAttributes(
		attribute.Key("event").String("string-format"),
		attribute.Key("value").String(helloStr),
	)

	return helloStr
}

func printHello(ctx context.Context, helloStr string) {
	tr := otel.Tracer(Name)

	_, span := tr.Start(ctx, "client-service:printHello")
	span.SetAttributes(attribute.Key("helloStr").String(helloStr))

	defer func() {
		span.End()
	}()

	println(helloStr)
}

func produce(ctx context.Context) []byte {
	tr := otel.Tracer(Name)

	_, span := tr.Start(ctx, "client-service:produce")

	defer func() {
		span.End()
	}()

	url := "http://localhost:9001/produce"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		panic(err.Error())
	}

	resp, err := helper.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		panic(err.Error())
	}

	span.SetAttributes(
		attribute.Key("event").String("produce"),
		attribute.Key("url").String(url),
		attribute.Key("value").String(string(resp)),
	)

	return resp
}

func consume(ctx context.Context, data []byte, search string) Product {
	tr := otel.Tracer(Name)

	_, span := tr.Start(ctx, "client-service:consume")

	defer func() {
		span.End()
	}()

	url := fmt.Sprintf("http://localhost:9002/consume?search=%s", search)
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		panic(err.Error())
	}

	resp, err := helper.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		panic(err.Error())
	}

	span.SetAttributes(
		attribute.Key("event").String("consume"),
		attribute.Key("url").String(url),
		attribute.Key("search").String(search),
		attribute.Key("value").String(string(resp)),
	)

	var product Product
	err = json.Unmarshal(resp, &product)

	return product
}

// Product data
type Product struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}