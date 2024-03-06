# OpenTelemetry in Go with Elastic Observability

This project showcase how to instrument a microservice written in [Go](https://golang.org/) using [OpenTelemetry](https://opentelemetry.io/), to produce telemetry data (traces and metrics) to [Elastic Observability](https://www.elastic.co/observability).

## Run with the collector

The Go microservice sends the traces and metrics to a collector that forwards them to Elastic Observability.

```bash
docker compose -f run-with-collector.yaml up -d
```

## Run without the collector

The Go microservice sends the traces and metrics directly to Elastic Observability.

```bash
docker compose -f run-without-collector.yaml up -d
```

## Accessing Elastic Observability

After executing the services you can reach the Elastic Observability application in the following URL:

```bash
http://localhost:5601/app/apm/services
```

Use the following credentials:

```bash
User: admin
Pass: changeme
```

## Invoking the microservice API

Once everything is running, periodic requests will be sent to the microservice, so you don't need to issue any requests by yourself. However, if you want to do it anyway, just execute:

```bash
curl -X GET http://localhost:8888/hello
```

# SELinux Permissions

If your host is running SELinux, the containers may be unable to access files that
are mounted from volumes, such as `/usr/share/elasticsearch/config/roles.yml`.
You can fix this by relabeling the files and directories that are to be mounted:

```bash
    chcon -R -t container_file_t collector-config.yaml environment fleet-server
```

# License

This project is licensed under the [Apache 2.0 License](./LICENSE).