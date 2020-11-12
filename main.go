package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/propagators"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
)

const (
	serviceName    = "hello-app"
	serviceVersion = "1.0"
)

var tracer = global.Tracer(serviceName)

func main() {

	collectorAddress := os.Getenv("COLLECTOR_ADDRESS")
	exporter, err := otlp.NewExporter(
		otlp.WithInsecure(),
		otlp.WithAddress(collectorAddress))

	if err != nil {
		log.Fatalf("Error creating the collector: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	defer bsp.Shutdown()

	res := resource.New(
		semconv.ServiceNameKey.String(serviceName),
		semconv.ServiceVersionKey.String(serviceVersion),
		semconv.TelemetrySDKNameKey.String("opentelemetry"),
		semconv.TelemetrySDKLanguageKey.String("go"),
		semconv.TelemetrySDKVersionKey.String("0.13.0"))

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(res))

	global.SetTracerProvider(tracerProvider)
	global.SetTextMapPropagator(otel.NewCompositeTextMapPropagator(
		propagators.TraceContext{}, propagators.Baggage{}))

	router := mux.NewRouter()
	router.Use(otelmux.Middleware(serviceName))
	router.HandleFunc("/hello", hello)
	http.ListenAndServe(":8888", router)

}

// Response struct
type Response struct {
	Message string `json:"message"`
}

func hello(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()

	_, customSpan := tracer.Start(ctx, "custom-span",
		trace.WithAttributes(
			label.String("custom-label", "Gopher")))
	customSpan.End()

	response := Response{"Hello World"}
	bytes, _ := json.Marshal(response)
	writer.Header().Add("Content-Type",
		"application/json")

	writer.Write(bytes)

}
