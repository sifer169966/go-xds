#! /bin/sh

# ./healtz.sh {Monitor port} {gRPC Healthcheck port} {gRPC Healtcheck connection timeout} {gRPC Healtcheck rpc timeout}

response=$(curl -s -o /dev/null -w "%{http_code}" localhost:$1/healthz 2>&1)
exit_code=$?
if [ $exit_code -ne 0 ] || [ "$response" != "200" ]; then
    echo "Error: Monitor server response with: $response"
    exit 1
fi

response=$(grpc-health-probe -addr=localhost:$2 -connect-timeout $3 -rpc-timeout $4)
exit_code=$?
if [ $exit_code -ne 0 ]; then
    echo "Error: gRPC server response with: $response"
    exit 1
fi
