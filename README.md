# OpenTelemetry in Go with Elastic APM

This project showcase how to use [Elastic APM](https://www.elastic.co/apm) with a service written in [Go](https://golang.org/) and instrumented using [OpenTelemetry](https://opentelemetry.io/). Everything is based on [Docker Compose](https://docs.docker.com/compose/) and you can test it with Elastic APM running locally or running on [Elastic Cloud](https://www.elastic.co/cloud/).

Once everything is running there will periodic requests being sent to the Go service, so you don't need to issue any requests by yourself. However, if you want to do it anyway just execute:

```bash
curl -X GET http://localhost:8888/hello
```

# License

This project is licensed under the [Apache 2.0 License](./LICENSE).