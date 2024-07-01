#!/bin/bash

STATEFULSETS=(
    "postgresql"
)

DEPLOYMENTS=(
    "yunikorn-admission-controller"
    "yunikorn-scheduler"
)

NAMESPACE=yunikorn

TIMEOUT=1200
INTERVAL=10

check_statefulsets() {
    local ss
    for ss in "${STATEFULSETS[@]}"; do
        echo "Checking StatefulSet $ss..."

        READY=$(kubectl get statefulset "$ss" -o jsonpath='{.status.readyReplicas}' --namespace $NAMESPACE)

        if [ -z "$READY" ] || [ "$READY" -le 0 ]; then
            echo "StatefulSet $ss does not have at least 1 ready replica: $READY"
            return 1
        else
            echo "StatefulSet $ss has $READY ready replica."
        fi
    done
    return 0
}

check_deployments() {
    local d
    for d in "${DEPLOYMENTS[@]}"; do
        echo "Checking Deployment $d..."

        READY=$(kubectl get deployment "$d" -o jsonpath='{.status.readyReplicas}' --namespace $NAMESPACE)

        if [ -z "$READY" ] || [ "$READY" -le 0 ]; then
            echo "Deployment $d does not have at least 1 ready replica: $READY"
            return 1
        else
            echo "Deployment $d has $READY at least 1 ready replica."
        fi
    done
    return 0
}

start_time=$(date +%s)

while true; do
    postgres_ready=false
    yunikorn_ready=false
    if check_statefulsets; then
        echo "All StatefulSets have at least 1 ready replicas."
        postgres_ready=true
    fi
    if check_deployments; then
        echo "All Deployments have at least 1 ready replicas."
        yunikorn_ready=true
    fi
    if [ "$postgres_ready" = true ] && [ "$yunikorn_ready" = true ]; then
        break
    fi

    elapsed_time=$(($(date +%s) - start_time))
    if [ "$elapsed_time" -ge "$TIMEOUT" ]; then
        echo "Timeout reached: Not all StatefulSets are ready within ${TIMEOUT} seconds."
        exit 1
    fi

    echo "Waiting for StatefulSets to be ready... (elapsed time: ${elapsed_time} seconds)"
    sleep $INTERVAL
done
