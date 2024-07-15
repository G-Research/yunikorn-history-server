#!/bin/bash

set -e

TEST_TYPE=$1
if [ "$TEST_TYPE" != "integration" ] && [ "$TEST_TYPE" != "e2e" ] && [ "$TEST_TYPE" != "performance" ]; then
    echo "Invalid test type: $TEST_TYPE. Please provide either 'integration', 'e2e' or 'performance' as the first argument."
    exit 1
fi

YHS_SERVER=$(YHS_SERVER:-http://localhost:8989}

# Cleanup the kind cluster and yunikorn-history-server process when the script exits
cleanup () {
    echo "**********************************"
    echo "Deleting kind cluster"
    echo "**********************************"
    KIND_CLUSTER=yhs-test make kind-delete-cluster

    if [ "$TEST_TYPE" == "performance" ]; then
        echo "**********************************"
        echo "Terminating yunikorn-history-server"
        echo "**********************************"
        pkill -f yunikorn-history-server
    fi
}
trap cleanup EXIT

wait_for_yhs() {
    echo "**********************************"
    echo "Waiting for yunikorn history server to start"
    echo "**********************************"
    while true; do
        echo "Sending request to yunikorn history server..."
        url="$YHS_SERVER/ws/v1/health/readiness"
        status_code=$(curl --write-out %{http_code} --silent --output /dev/null $url || true)

        if [ "$status_code" -eq 200 ] ; then
            echo "Yunikorn history server is up and running."
            break
        else
            echo "Waiting for yunikorn history server to start..."
            sleep 10
        fi
    done
}

echo "**********************************"
echo "Creating kind cluster"
echo "**********************************"
KIND_CLUSTER=yhs-test make kind-create-cluster

echo "**********************************"
echo "Install and configure dependencies"
echo "**********************************"
make install-dependencies migrate-up

if [ "$TEST_TYPE" == "performance" ]; then
    echo "**********************************"
    echo "Run yunikorn history server"
    make clean build
    make run &

    # Wait for yunikorn history server to start
    wait_for_yhs
fi

echo "**********************************"
echo "Running $TEST_TYPE tests"
echo "**********************************"
if [ "$TEST_TYPE" == "performance" ]; then
    make "test-k6-$TEST_TYPE"
else
    make "test-go-$TEST_TYPE"
fi
