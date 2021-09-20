module otel-with-golang

go 1.17

require go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.0.0

require (
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.0.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.23.0
	go.opentelemetry.io/proto/otlp v0.9.0 // indirect
	golang.org/x/net v0.0.0-20210917221730-978cfadd31cf // indirect
	golang.org/x/sys v0.0.0-20210917161153-d61c044b1678 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210920155426-26f343e4c215 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)

require (
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.23.0 // indirect
	go.opentelemetry.io/otel/sdk/export/metric v0.23.0 // indirect
)

require (
	github.com/gorilla/mux v1.8.0
	go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux v0.23.0
	go.opentelemetry.io/otel v1.0.0
	go.opentelemetry.io/otel/internal/metric v0.23.0 // indirect
	go.opentelemetry.io/otel/metric v0.23.0
	go.opentelemetry.io/otel/sdk v1.0.0
	go.opentelemetry.io/otel/trace v1.0.0
	google.golang.org/grpc v1.40.0
)

require (
	github.com/felixge/httpsnoop v1.0.2 // indirect
	go.opentelemetry.io/contrib v0.23.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.23.0
)
