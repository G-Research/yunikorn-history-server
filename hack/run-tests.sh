#!/bin/bash

set -e

TEST_TYPE=$1
if [ "$TEST_TYPE" != "integration" ] && [ "$TEST_TYPE" != "e2e" ]; then
    echo "Invalid test type: $TEST_TYPE. Please provide either 'integration' or 'e2e' as the first argument."
    exit 1
fi

# Cleanup the kind cluster when the script exits
cleanup () {
    echo "**********************************"
    echo "Deleting kind cluster"
    echo "**********************************"
    KIND_CLUSTER=yhs-integration-tests make kind-delete-cluster
}
trap cleanup EXIT

echo "**********************************"
echo "Creating kind cluster"
echo "**********************************"
KIND_CLUSTER=yhs-integration-tests make kind-create-cluster

echo "**********************************"
echo "Install and configure dependencies"
echo "**********************************"
make install-dependencies

echo "**********************************"
echo "Apply database migrations"
echo "**********************************"
make migrate-up

echo "**********************************"
echo "Running $TEST_TYPE tests"
echo "**********************************"
make "test-go-$TEST_TYPE"
