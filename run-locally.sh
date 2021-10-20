#!/bin/bash

export EXPORTER_ENDPOINT=https://clusterid.apm.region.provider.elastic-cloud.com:443
export EXPORTER_HEADERS="Authorization=Bearer APM_SECRET_TOKEN"

go run main.go
