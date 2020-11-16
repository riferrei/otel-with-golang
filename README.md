# OpenTelemetry in Go with Elastic APM

This project showcase how to use [Elastic APM](https://www.elastic.co/apm) with a microservice written in [Go](https://golang.org/) and instrumented using [OpenTelemetry](https://opentelemetry.io/). Everything is based on [Docker Compose](https://docs.docker.com/compose/) and you can test it with Elastic APM running locally or running on [Elasticsearch Service](https://www.elastic.co/elasticsearch/service).

## Elastic APM running in your local machine

Just execute:

```bash
docker-compose -f docker-compose-local.yaml up -d
```

## Elastic APM running on Elasticsearch Service

You will need to edit the file [collector-config-cloud.yaml](collector-config-cloud.yaml) and provide the following information:

```bash
exporters:
  elastic:
    apm_server_url: "<APM_SERVER_URL>"
    secret_token: "<SECRET_TOKEN>"
```

Then you can execute:

```bash
docker-compose -f docker-compose-cloud.yaml up -d
```

Once everything is running there will periodic requests being sent to the microservice so you don't need to issue any requests by yourself. However, if you want to do it anyway just execute:

```bash
curl -X GET http://localhost:8888/hello
```

# License

This project is licensed under the [Apache 2.0 License](./LICENSE).
