package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/propagation"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

const (
	metricPrefix     = "custom.metric."
	numberOfExecName = metricPrefix + "number.of.exec"
	numberOfExecDesc = "Count the number of executions."
	heapMemoryName   = metricPrefix + "heap.memory"
	heapMemoryDesc   = "Reports heap memory utilization."
	serviceName      = "hello-app"
	serviceVersion   = "1.0"
)

var (
	tracer             trace.Tracer
	meter              metric.Meter
	numberOfExecutions metric.BoundInt64Counter
)

func main() {

	ctx := context.Background()

	// Create an gRPC-based OTLP exporter that
	// will receive the created telemetry data
	endpoint := os.Getenv("EXPORTER_ENDPOINT")
	driver := otlpgrpc.NewDriver(
		otlpgrpc.WithInsecure(),
		otlpgrpc.WithEndpoint(endpoint),
	)
	exporter, err := otlp.NewExporter(ctx, driver)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create exporter", err)
	}

	// Create a resource to decorate the app
	// with common attributes from OTel spec
	res0urce, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		),
	)
	if err != nil {
		log.Fatalf("%s: %v", "failed to create resource", err)
	}

	// Create a tracer provider that processes
	// spans using a batch-span-processor. This
	// tracer provider will create a sample for
	// every trace created, which is great for
	// demos but horrible for production –– as
	// volume of data generated will be intense
	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res0urce),
		sdktrace.WithSpanProcessor(bsp),
	)

	// Creates a pusher for the metrics that runs
	// in the background and push data every 5sec
	pusher := controller.New(
		processor.New(
			simple.NewWithExactDistribution(),
			exporter,
		),
		controller.WithExporter(exporter),
		controller.WithCollectPeriod(5*time.Second),
	)
	err = pusher.Start(ctx)
	if err != nil {
		log.Fatalf("%s: %v", "failed to start the controller", err)
	}
	defer func() { _ = pusher.Stop(ctx) }()

	// Register the tracer provider and propagator
	// so libraries and frameworks used in the app
	// can reuse it to generate traces and metrics
	otel.SetTracerProvider(tracerProvider)
	global.SetMeterProvider(pusher.MeterProvider())
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.Baggage{},
			propagation.TraceContext{},
		),
	)

	// Instances to support custom traces/metrics
	tracer = otel.Tracer(serviceName)
	meter = global.Meter(serviceName)

	// Creating a custom metric that is updated
	// manually each time the API is executed
	numberOfExecutions = metric.Must(meter).
		NewInt64Counter(
			numberOfExecName,
			metric.WithDescription(numberOfExecDesc),
		).Bind(
		[]attribute.KeyValue{
			semconv.ServiceNameKey.String(serviceName),
			attribute.String(
				numberOfExecName,
				numberOfExecDesc)}...)

	// Creating a custom metric that is updated
	// automatically using an int64 observer
	_ = metric.Must(meter).
		NewInt64ValueObserver(
			heapMemoryName,
			func(_ context.Context, result metric.Int64ObserverResult) {
				var mem runtime.MemStats
				runtime.ReadMemStats(&mem)
				result.Observe(int64(mem.HeapAlloc),
					semconv.ServiceNameKey.String(serviceName),
					attribute.String(heapMemoryName,
						heapMemoryDesc))
			},
			metric.WithDescription(heapMemoryDesc))

	// Register the API handler and starts the app
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

	// Creating a custom span just for fun...
	_, mySpan := tracer.Start(ctx, "mySpan")
	if response.isValid() {
		log.Print("The response is valid")
	}
	mySpan.End()

	// Updating the number of executions metric...
	numberOfExecutions.Add(ctx, 1)

}

func buildResponse(writer http.ResponseWriter) Response {

	writer.WriteHeader(http.StatusOK)
	writer.Header().Add("Content-Type",
		"application/json")

	bytes, _ := json.Marshal("Hello World")
	writer.Write(bytes)
	return Response{}

}

// Response struct
type Response struct {
}

func (r Response) isValid() bool {
	return true
}
