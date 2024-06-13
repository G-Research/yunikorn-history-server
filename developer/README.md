# Developer Docs and Tools

## Folder Contents

This folder contains scripts and configuration files to assist in development
of the Yunikorn History Server (YHS).

- README.md - this file
- docker-compose.yml - a Docker Compose file for running infrastructure services for YHS
- postgres-init.sh - a script to create the basic entities for the YHS database
- yk-deploy.sh - a script to start a K8S cluster with the Yunikorn Scheduler

## Required Tools/Packages

You will need several software packages installed to develop and test YHS. The necessary
packages, and how to install them are:

- Docker Desktop - packages can be downloaded from https://www.docker.com/products/docker-desktop/
After installation, you should be able to run `docker ps` and it should show no containers running.

- The Go compiler and toolset - this can be installed from tar archives from https://go.dev - they
offer packages for Windows, macOS, and Linux. Note that if you are running macOS, you can install
Go using the Brew package manager (`brew install go`). On Linux systems, you may be able to
install using `apt` or `dnf` (e.g. `sudo apt install golang`), but note that some Linux distributions
ship fairly old versions of Go.

- Either `kind` ("Kubernetes on Docker") or `minikube` to run a K8S cluster on Docker; see documentation for
`kind` at https://kind.sigs.k8s.io and `minikube` at https://minikube.sigs.k8s.io.
Usually the easiest way to install `kind` is by using `go install` directly to download and build
the binary: `go install sigs.k8s.io/kind@v0.23.0`. To install `minikube`, follow the download
instructions at https://minikube.sigs.k8s.io.

- The `curl` and `jq` utilities are very useful for fetching and formatting response data from Yunikorn
and YHS servers, to do quick manual checks. They are often available via package managers on your system,
or can be downloaded directly from https://curl.se/ and https://github.com/jqlang/jq, respectively.


## Repository setup for developing and running Yunikorn and YHS

Developing the Yunikorn service assumes a layout of having each of the four
Yunikorn project repositories in a shared layout, and YHS uses this same layout.

YHS assumes you have the repos in a shared top-level directory at `$HOME/src/yunikorn` - if
you want it elsewhere, pass the absolute path of the top-level directory as shown in the following command.
```sh
  $ ./yk-deploy.sh /absolute/path/to/your/yunikorn-base-dir
``` 

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

- Start a K8S cluster, with Yunikorn. Change into the `yunikorn-history-server/developer` directory.
It will use `kind` to deploy a K8S cluster on Docker, but if you prefer `minikube`, edit the `yk-deploy.sh`
file, and change the `k8s_mgr` variable setting to `minikube`.

Then run:
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

