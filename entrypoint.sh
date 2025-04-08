#!/bin/bash

# Start the gRPC server in the background
/app/grpc-server &

# Start the REST API server in the foreground
exec /app/rest-server 