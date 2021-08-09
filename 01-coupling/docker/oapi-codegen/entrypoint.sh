#!/bin/bash

for dir in /src/*; do
    echo "Running in $dir"

    oapi-codegen -generate types -o "$dir/internal/http_types.go" -package internal "$dir/openapi.yml"
    oapi-codegen -generate chi-server -o "$dir/internal/http_server.go" -package internal "$dir/openapi.yml"
done
