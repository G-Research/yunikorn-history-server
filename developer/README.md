# Developer Docs and Tools

## Folder Contents

This folder contains scripts and configuration files to assist in development
of the Yunikorn History Server (YHS).

- README.md - this file
- docker-compose.yml - a Docker Compose file for running infrastructure services for YHS
- postgres-init.sh - a script to create the basic entities for the YHS database
- yk-deploy.sh - a script to start a Kind (K8S) cluster with the Yunikorn Scheduler

## Repository setup for developing and running Yunikorn and YHS

Developing the Yunikorn service assumes a layout of having each of the four
Yunikorn project repositories in a shared layout, and YHS uses this same layout.

YHS assumes you have the repos in a shared top-level directory at `$HOME/src/yunikorn` - if
you want it elsewhere, adjust the value of `yk_repos_base` at the top of `yk-deploy.sh`.

The four Yunikorn repositories to be used are from:
```
  https://github.com/apache/yunikorn-core
  https://github.com/apache/yunikorn-k8shim
  https://github.com/apache/yunikorn-scheduler-interface
  https://github.com/apache/yunikorn-web
```

Clone (or fork) those four repos, as well as this repo (https://github.com/G-Research/yunikorn-history-server)
so your top-level directory (e.g. `$HOME/src/yunikorn`) looks like:

```
yunikorn-core/                yunikorn-k8shim/              yunikorn-web/
yunikorn-history-server/      yunikorn-scheduler-interface/
```

## Running Yunikorn and YHS

- Start a K8S (`kind`) cluster, with Yunikorn. Change into the `yunikorn-history-server/developer` directory
and run:
```sh
  $ ./yk-deploy.sh
``` 

NOTE: this script will edit some of the deployment YAML files in the `yunikorn-k8shim` directory to reflect
the detected architecture (e.g. `amd64`) and preferred Docker registry that are set at the start of 
`yk-deploy.sh`. If you desire to reverse those changes, simply to go the `yunikorn-k8shim` directory and
do a `git checkout deployments/scheduler`.


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
  $ build/event-collector yhs.yml
``` 
You can safely stop it by pressing `<Control>+C`.

## Shutting Down 
To stop the Postgres container, either type `<Control>+C` if it's running in attached mode, or run `docker compose down postgres` if it's running in detached mode.

To stop the Yunikorn cluster, run `kind delete cluster`. To also clean up stale Docker container storage on your disk, run `docker system prune --volumes -f`, although it's not strictly necessary every time.

