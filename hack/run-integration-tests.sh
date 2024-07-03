#!/bin/bash

set -e

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
echo "Running integration tests"
echo "**********************************"
make test-go-integration
