#!/bin/bash

# Variables
NAMESPACE="yunikorn" # Change this to your service's namespace if it's not in the default namespace
SERVICE_NAME="yunikorn-service" # Replace with your service's name

kubectl patch service $SERVICE_NAME --namespace $NAMESPACE -p '{
  "spec": {
    "type": "NodePort",
    "ports": [
      {
        "name": "yunikorn-core",
        "port": 9080,
        "protocol": "TCP",
        "targetPort": "http1",
        "nodePort": 30000
      },
      {
        "name": "yunikorn-service",
        "port": 9889,
        "protocol": "TCP",
        "targetPort": "http2",
        "nodePort": 30001
      }
    ]
  }
}'
