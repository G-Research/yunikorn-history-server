# Developer Docs and Tools

## Folder Contents

This folder contains scripts and configuration files to assist in development
of the Yunikorn History Server (YHS).

- README.md - this file
- docker-compose.yml - a Docker Compose file for running infrastructure services for YHS
- postgres-init.sh - a script to create the basic entities for the YHS database
- yk-deploy.sh - a script to start a Kind (K8S) cluster with the Yunikorn Scheduler

## Running Yunikorn and YHS

- Start a K8S (Kind) cluster, with Yunikorn:
```sh
  $ ./yk-deploy.sh
``` 

- Run Kubernetes port-forwarding to allow external access to the K8S network:
```sh
  $ kubectl port-forward svc/yunikorn-service 9889 9080 -n yunikorn
``` 
(This will run continuously, so you will have to open a new terminal/shell for
subsequent commands.)

Verify that you can access the Yunikorn web UI by opening your browser to
http://localhost:9889/ .

- Start a Postgres container for the YHS database; this assumes you have Docker
installed and running on your system:
```sh
  $ docker compose up postgres
``` 
This will run continously - to detach the process and run it in the background, do:
```sh
  $ docker compose up -d postgres
``` 
- Go to the root of your yunikorn-history-server repository (i.e. one level up from
this directory) and build YHS:
```sh
  $ make clean; make
``` 
Then run YHS:
```sh
  $ build/event-collector
``` 

