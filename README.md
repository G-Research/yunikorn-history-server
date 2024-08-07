# Yunikorn History Server
[![GoReport Widget]][GoReport Status]
[![Latest Release](https://img.shields.io/github/v/release/G-Research/yunikorn-history-server?include_prereleases)](https://github.com/armadaproject/armada-operator/releases/latest)

[GoReport Widget]: https://goreportcard.com/badge/github.com/G-Research/yunikorn-history-server
[GoReport Status]: https://goreportcard.com/report/github.com/G-Research/yunikorn-history-server

Yunikorn History Server is an ancillary service for K8S Clusters using the Yunikorn Scheduler to
persist the state of a Yunikorn-managed cluster to a database, allowing for long-term
access to historical data of the cluster's operations (e.g. to view past Applications,
resource usage, etc.).

## Installation

### Prerequisites

Make sure you have the following dependencies installed:
* [PostgreSQL](https://www.postgresql.org/download/) - open-source relational database
* [Go v1.22+](https://golang.org/doc/install) (only needed for local development) - programming language used to build the Yunikorn History Server

### Quickstart

Use the following `make` commands to build and run the Yunikorn History Server locally:
```bash
# start all dependencies - if you are using kind as your K8S cluster manager:
make kind-all
# otherwise, if you want to use minikube for your cluster:
env CLUSTER_MGR=minikube make minikube-all

# build and run YHS
make run
```

## Contributing

Please feel free to contribute bug-reports or ideas for enhancements via
GitHub's issue system.

Code contributions are also welcome. When submitting a pull-request please
ensure it references a relevant issue as well as making sure all CI checks
pass.

## Testing

Please test contributions thoroughly before requesting reviews. At a minimum:
```bash
# Lint code
make lint
# Run tests using `kind` for cluster manager:
make tests
# Run tests using `minikube` for cluster manager:
env CLUSTER_MGR=minikube make tests
```
should all succeed without error.

Add and change appropriate unit and integration tests to ensure your changes
are covered by automated tests and appear to be correct.

## License

See the [LICENSE](LICENSE) file for licensing information.
