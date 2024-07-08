#!/bin/bash

cmd=$1
if [ "$cmd" != "up" ] && [ "$cmd" != "down" ]; then
  echo "Usage: $0 [up|down]"
  exit 1
fi

counter=0
threshold=3
interval=10
while [ $counter -lt $threshold ]; do
  make migrate ARGS=$cmd
  status=$?

  if [ $status -eq 0 ]; then
    echo "Migrations successfully applied!"
    break
  else
    echo "Error while applying migrations. Retrying in $interval seconds..."
  fi
    sleep $interval
    counter=$((counter + 1))
done
