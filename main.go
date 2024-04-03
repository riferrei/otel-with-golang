package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"
)

const (
	serviceName        = "hello-app"
	serviceVersion     = "v1.0.0"
	metricPrefix       = "custom.metric."
	numberOfExecName   = metricPrefix + "number.of.exec"
	numberOfExecDesc   = "Count the number of executions."
	heapMemoryName     = metricPrefix + "heap.memory"
	heapMemoryDesc     = "Reports heap memory utilization."
	elasticCloudSuffix = "cloud.es.io"
	httpsPreffix       = "https://"
)

var (
	tracer             trace.Tracer
	meter              metric.Meter
	numberOfExecutions metric.Int64Counter
)

func main() {

	ctx := context.Background()

	// OpenTelemetry agent connectivity data
	endpoint := os.Getenv("EXPORTER_ENDPOINT")
	headers := os.Getenv("EXPORTER_HEADERS")
	headersMap := func(headers string) map[string]string {
		headersMap := make(map[string]string)
		if len(headers) > 0 {
			headerItems := strings.Split(headers, ",")
			for _, headerItem := range headerItems {
				parts := strings.Split(headerItem, "=")
				headersMap[parts[0]] = parts[1]
			}
		}
		return headersMap
	}(headers)

	// Resource to name traces/metrics
	res0urce, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.TelemetrySDKVersionKey.String(otel.Version()),
			semconv.TelemetrySDKLanguageGo,
		),
	)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create resource", err)
	}

	// Initialize the default logger
	initLogger()

	// Initialize the tracer provider
	initTracer(ctx, endpoint, headersMap, res0urce)

	// Initialize the meter provider
	initMeter(ctx, endpoint, headersMap, res0urce)

	// Create the metrics
	createMetrics()

	// Start the microservice
	router := mux.NewRouter()
	router.Use(otelmux.Middleware(serviceName))
	router.HandleFunc("/hello", hello)
	http.ListenAndServe(":8888", router)

}

func hello(writer http.ResponseWriter, request *http.Request) {

	ctx := request.Context()

	ctx, buildResp := tracer.Start(ctx, "buildResponse")
	response := buildResponse(writer)
	buildResp.End()

	// Create a custom span
	_, mySpan := tracer.Start(ctx, "mySpan")
	if response.isValid() {
		// Log records with context will include the trace id.
		slog.InfoContext(ctx, "The response is valid")
	}
	mySpan.End()

	// Update the metric
	numberOfExecutions.Add(ctx, 1,
		metric.WithAttributes(attribute.String(numberOfExecName, numberOfExecDesc)))
}

func buildResponse(writer http.ResponseWriter) response {

	writer.Header().Add("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	response := response{"Hello World"}
	bytes, _ := json.Marshal(response)
	writer.Write(bytes)
	return response

}

type response struct {
	Message string `json:"Message"`
}

func (r response) isValid() bool {
	return true
}

func initTracer(ctx context.Context, endpoint string,
	headersMap map[string]string, res0urce *resource.Resource) {

	traceOpts := []otlptracegrpc.Option{
		otlptracegrpc.WithTimeout(5 * time.Second),
	}
	if strings.Contains(endpoint, elasticCloudSuffix) {
		endpoint = strings.ReplaceAll(endpoint, httpsPreffix, "")
		traceOpts = append(traceOpts, otlptracegrpc.WithHeaders(headersMap))
		traceOpts = append(traceOpts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		traceOpts = append(traceOpts, otlptracegrpc.WithInsecure())
	}
	traceOpts = append(traceOpts, otlptracegrpc.WithEndpoint(endpoint))

	traceExporter, err := otlptracegrpc.New(ctx, traceOpts...)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create exporter", err)
	}

	otel.SetTracerProvider(sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res0urce),
		sdktrace.WithSpanProcessor(
			sdktrace.NewBatchSpanProcessor(traceExporter)),
	))

	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.Baggage{},
			propagation.TraceContext{},
		),
	)

	tracer = otel.Tracer("io.opentelemetry.traces.hello")

}

func initMeter(ctx context.Context, endpoint string,
	headersMap map[string]string, res0urce *resource.Resource) {

	metricOpts := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithTimeout(5 * time.Second),
	}
	if strings.Contains(endpoint, elasticCloudSuffix) {
		endpoint = strings.ReplaceAll(endpoint, httpsPreffix, "")
		metricOpts = append(metricOpts, otlpmetricgrpc.WithHeaders(headersMap))
		metricOpts = append(metricOpts, otlpmetricgrpc.WithTLSCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		metricOpts = append(metricOpts, otlpmetricgrpc.WithInsecure())
	}
	metricOpts = append(metricOpts, otlpmetricgrpc.WithEndpoint(endpoint))

	metricExporter, err := otlpmetricgrpc.New(ctx, metricOpts...)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create exporter", err)
	}

	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res0urce),
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				metricExporter,
				sdkmetric.WithInterval(5*time.Second), // Default is 1m
			),
		),
	)
	otel.SetMeterProvider(meterProvider)

	meter = meterProvider.Meter("io.opentelemetry.metrics.hello")
}

func createMetrics() {
	var err error

	// Metric to be updated manually
	numberOfExecutions, err = meter.Int64Counter(
		numberOfExecName,
		metric.WithDescription(numberOfExecDesc),
	)
	if err != nil {
		log.Fatalf("%s %q: %v", "failed to create metric", numberOfExecName, err)
	}

	// Metric to be updated automatically
	_, err = meter.Int64ObservableCounter(
		heapMemoryName,
		metric.WithDescription(heapMemoryDesc),
		metric.WithInt64Callback(
			func(_ context.Context, obs metric.Int64Observer) error {
				var mem runtime.MemStats
				runtime.ReadMemStats(&mem)
				obs.Observe(int64(mem.HeapAlloc),
					metric.WithAttributes(attribute.String(heapMemoryName, heapMemoryDesc)))
				return nil
			},
		),
	)
	if err != nil {
		log.Fatalf("%s %q: %v", "failed to create metric", heapMemoryName, err)
	}
}

// Embed slog.Handler so that we can wrap the Handle method
type slogHandler struct {
	slog.Handler
}

// Handle adds trace.id to the Record before calling the underlying handler.
func (h slogHandler) Handle(ctx context.Context, r slog.Record) error {
	if spanContext := trace.SpanContextFromContext(ctx); spanContext.IsValid() {
		r.AddAttrs(slog.String("trace.id", spanContext.TraceID().String()))
	}
	return h.Handler.Handle(ctx, r)
}

func initLogger() {
	logHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	slog.SetDefault(slog.New(&slogHandler{logHandler}))
}
