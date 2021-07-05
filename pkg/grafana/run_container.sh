#!/bin/sh

docker run --name grafana \
           -v $(pwd)/testdata:/etc/grafana \
           -e GF_PATHS_CONFIG=/etc/grafana/custom.ini \
           -e GF_PATHS_PROVISIONING=/etc/grafana/provisioning \
           --rm -p3000:3000 \
           grafana/grafana:8.0.4
