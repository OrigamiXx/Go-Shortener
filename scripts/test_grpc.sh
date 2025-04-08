#!/bin/bash

# Exit on error
set -e

echo "Generating protobuf code..."
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/urlshortener.proto

echo "Running gRPC tests..."
go test -v ./cmd/grpc/server/...

echo "Tests completed successfully!" 