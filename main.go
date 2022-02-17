package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"log"
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
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
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
			semconv.TelemetrySDKVersionKey.String("v1.4.1"),
			semconv.TelemetrySDKLanguageGo,
		),
	)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create resource", err)
	}

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
		log.Print("The response is valid")
	}
	mySpan.End()

	// Update the metric
	numberOfExecutions.Add(ctx, 1,
		[]attribute.KeyValue{
			attribute.String(
				numberOfExecName,
				numberOfExecDesc)}...)

}

func buildResponse(writer http.ResponseWriter) response {

	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type",
		"application/json")

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

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			metricExporter,
		),
		controller.WithResource(res0urce),
		controller.WithExporter(metricExporter),
		controller.WithCollectPeriod(5*time.Second),
	)

	err = pusher.Start(ctx)
	if err != nil {
		log.Fatalf("%s: %v", "failed to start the pusher", err)
	}

	global.SetMeterProvider(pusher)
	meter = global.Meter("io.opentelemetry.metrics.hello")

}

func createMetrics() {

	// Metric to be updated manually
	numberOfExecutions = metric.Must(meter).
		NewInt64Counter(
			numberOfExecName,
			metric.WithDescription(numberOfExecDesc),
		)

	// Metric to be updated automatically
	_ = metric.Must(meter).
		NewInt64CounterObserver(
			heapMemoryName,
			func(_ context.Context, result metric.Int64ObserverResult) {
				var mem runtime.MemStats
				runtime.ReadMemStats(&mem)
				result.Observe(int64(mem.HeapAlloc),
					attribute.String(heapMemoryName,
						heapMemoryDesc))
			},
			metric.WithDescription(heapMemoryDesc))

}
