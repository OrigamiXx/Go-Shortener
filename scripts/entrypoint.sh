#!/bin/sh

# Start the REST API server in the background
./server &

# Start the gRPC server
./grpc-server

# Wait for any process to exit
wait -n

# Exit with status of process that exited first
exit $? 