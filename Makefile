#
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Check Make version (we need at least GNU Make 3.82).
ifeq ($(filter undefine,$(value .FEATURES)),)
$(error Unsupported Make version. \
    The build system does not work properly with GNU Make $(MAKE_VERSION), \
    please use GNU Make 3.82 or above, version 4.3 or higher works best)
endif

# Check if this GO tools version used is at least the version of go specified in
# the go.mod file. The version in go.mod should be in sync with other repos.

# Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin

# LOCALBIN_DOCKER refers to the directory where docker tarballs are created.
LOCALBIN_DOCKER ?= $(LOCALBIN)/docker
bin/docker: ## Create local bin directory for docker artifacts if necessary.
	mkdir -p $(LOCALBIN_DOCKER)

# LOCALBIN_TOOLING refers to the directory where tooling binaries are installed.
LOCALBIN_TOOLING ?= $(LOCALBIN)/tooling
bin/tooling: ## Create local bin directory for tooling if necessary.
	mkdir -p $(LOCALBIN_TOOLING)

# LOCALBIN_APP refers to the directory where application binaries are installed.
LOCALBIN_APP ?= $(LOCALBIN)/app
bin/app: ## Create local bin directory for app binary if necessary.
	mkdir -p $(LOCALBIN_APP)

# PLATFORMS defines the target platforms for the operator image.
PLATFORMS ?= linux/amd64,linux/arm64
# IMAGE_REGISTRY defines the registry where the operator image will be pushed.
IMAGE_REGISTRY ?= gresearch
# IMAGE_NAME defines the name of the operator image.
IMAGE_NAME := unicorn-history-server
# IMAGE_REPO defines the image repository and name where the operator image will be pushed.
IMAGE_REPO ?= $(IMAGE_REGISTRY)/$(IMAGE_NAME)
# BUILD_TIME defines the build time of the operator image.
BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
# GIT_COMMIT defines the git commit of the operator image.
GIT_COMMIT ?= $(shell git rev-parse HEAD)
# GIT_TAG defines the git tag of the operator image.
GIT_TAG ?= $(shell git describe --tags --dirty --always)
# IMAGE_TAG defines the name and tag of the operator image.
IMAGE_TAG ?= $(IMAGE_REPO):$(GIT_TAG)

# Go compiler selection
GO := go
GO_VERSION := $(shell $(GO) version | awk '{print substr($$3, 3, 4)}')
MOD_VERSION := $(shell cat .go_version)

GM := $(word 1,$(subst ., ,$(GO_VERSION)))
MM := $(word 1,$(subst ., ,$(MOD_VERSION)))
FAIL := $(shell if [ $(GM) -lt $(MM) ]; then echo MAJOR; fi)
ifdef FAIL
$(error Build should be run with at least go $(MOD_VERSION) or later, found $(GO_VERSION))
endif
GM := $(word 2,$(subst ., ,$(GO_VERSION)))
MM := $(word 2,$(subst ., ,$(MOD_VERSION)))
FAIL := $(shell if [ $(GM) -lt $(MM) ]; then echo MINOR; fi)
ifdef FAIL
$(error Build should be run with at least go $(MOD_VERSION) or later, found $(GO_VERSION))
endif

# Force Go modules even when checked out inside GOPATH
GO111MODULE := on
export GO111MODULE

# Machine info
OS ?= $(shell $(GO) env GOOS)
ARCH ?= $(shell $(GO) env GOARCH)

# Local Development
CLUSTER_MGR ?= kind     # either 'kind' or 'minikube'
CLUSTER_NAME ?= uhs
NAMESPACE ?= yunikorn

KIND ?= $(LOCALBIN_TOOLING)/kind
KIND_VERSION ?= latest

MINIKUBE ?= $(LOCALBIN_TOOLING)/minikube
MINIKUBE_VERSION ?= latest

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Database

UHS_CONFIG ?= config/unicorn-history-server/local.yml

define yq_get
    $(YQ) e '$(1)' $(UHS_CONFIG)
endef

define yq_get_db
    $(shell $(call yq_get, .db.$(1)))
endef

define url_escape
    $(shell printf "%s" $(1) | go run hack/url/main.go)
endef

define yq_get_jk
    $(shell $(call yq_get, .yunikorn.$(1)))
endef

define yq_get_uhs
	$(shell $(call yq_get, .uhs.$(1)))
endef

DB_USER ?= $(strip $(call url_escape,$(strip $(call yq_get_db,user))))
DB_PASSWORD ?= $(strip $(call url_escape,$(strip $(call yq_get_db,password))))
DB_HOST ?= $(strip $(call url_escape,$(strip $(call yq_get_db,host))))
DB_PORT ?= $(strip $(call yq_get_db,port))
DB_NAME ?= $(strip $(call url_escape,$(strip $(call yq_get_db,dbname))))
YUNIKORN_HOST ?= $(strip $(call yq_get_jk,host))
YUNIKORN_PORT ?= $(strip $(call yq_get_jk,port))
UHS_PORT ?= $(strip $(call yq_get_uhs,port))

define database_url
	postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable
endef

define yunikorn_api_url
	http://$(YUNIKORN_HOST):$(YUNIKORN_PORT)
endef

define uhs_api_url
	http://$(YUNIKORN_HOST):$(UHS_PORT)
endef

.PHONY: migrate
migrate: yq gomigrate ## run migrations.
	$(GOMIGRATE) -path migrations -database $(call database_url) $(ARGS)

.PHONY: migrate-up
migrate-up: ## migrate up using gomigrate.
	hack/migrate.sh up

.PHONY: migrate-down
migrate-down: ## migrate down using gomigrate.
	hack/migrate.sh down

##@ Codegen

.PHONY: codegen
codegen: mockgen ## generate code using go generate (mocks).
	PATH=$(LOCALBIN_TOOLING):$$PATH go generate ./...

##@ Run

.PHONY: run
run: ## run the unicorn-history-server binary.
	go run cmd/unicorn-history-server/main.go --config config/unicorn-history-server/local.yml

##@ Json Server

.PHONY: json-server
json-server: ## start the mock server using json-server.
	cd web && npm run start:json-server

##@ Lint

.PHONY: lint
lint: go-lint ## lint code.

.PHONY: go-lint
go-lint: golangci-lint ## lint Golang code using golangci-lint.
	$(GOLANGCI_LINT) run

.PHONY: go-lint
go-lint-fix: golangci-lint ## lint Golang code using golangci-lint.
	$(GOLANGCI_LINT) run --fix

##@ Test

define start-cluster
	@echo "**********************************"
	@echo "Creating cluster"
	@echo "**********************************"
	@CLUSTER_NAME=uhs-test $(MAKE) create-cluster

	@echo "**********************************"
	@echo "Install and configure dependencies"
	@echo "**********************************"
	$(MAKE) install-dependencies migrate-up
endef

define cleanup-cluster
	cleanup() {
	    echo "**********************************"
	    echo "Deleting cluster"
	    echo "**********************************"
	    CLUSTER_NAME=uhs-test $(MAKE) delete-cluster
    }
endef

.PHONY: test
test: test-go-unit integration-tests ## run all tests.

.PHONY: integration-tests
.ONESHELL:
integration-tests: ## start dependencies and run integration tests.
	@$(cleanup-cluster); trap cleanup EXIT
	@$(start-cluster)
	UHS_SERVER=${UHS_SERVER:-http://localhost:8989} $(MAKE) test-go-integration

.PHONY: e2e-tests
.ONESHELL:
e2e-tests: ## start dependencies and run e2e tests.
	@$(cleanup-cluster); trap cleanup EXIT
	@$(start-cluster)
	CLUSTER_NAME=uhs-test UHS_SERVER=${UHS_SERVER:-http://localhost:8989} $(MAKE) test-go-e2e

.PHONY: performance-tests
.ONESHELL:
performance-tests: k6 ## start dependencies and run performance tests.
	@$(cleanup-cluster)
	@stop_perf_cluster() {
	    uhs_pid=`ps ax | grep 'unicorn-history-server' | grep -v grep | awk '{print $$1}'`
		if [ "$${uhs_pid}" != "" ] ; then
		    echo "**********************************"
		    echo "Terminating unicorn-history-server"
		    echo "**********************************"
		    kill -TERM $${uhs_pid}
		fi
		cleanup
	}; trap stop_perf_cluster EXIT
	@$(start-cluster)
	@echo "**********************************"
	@echo "Run unicorn history server"
	@mkdir -p test-reports/performance
	$(MAKE) clean build
	bin/app/unicorn-history-server \
		--config config/unicorn-history-server/local.yml > test-reports/performance/uhs.log & disown
	UHS_SERVER=$${UHS_SERVER:-http://localhost:8989}
	@echo "UHS_SERVER is $${UHS_SERVER}"
	@echo "**********************************"
	@echo "Waiting for unicorn history server to start"
	@echo "**********************************"
	while true; do
		echo "Sending request to unicorn history server..."
		URL="$${UHS_SERVER}/ws/v1/health/readiness"
		http_status=`curl --write-out %{http_code} --silent --output /dev/null $${URL} || true`
		if [ $$http_status -eq 200 ] ; then
			echo "Unicorn history server is up and running."
			break
		else
			echo "Waiting for unicorn history server to start..."
			sleep 10
		fi
	done
	echo "**********************************"
	echo "Running performance tests"
	echo "**********************************"
	$(MAKE) test-k6-performance

TEST_ARGS ?= --junitfile=test-reports/junit.xml --jsonfile=test-reports/report.json -- -coverprofile=test-reports/coverage.out -covermode=atomic

.PHONY: test-go-unit
test-go-unit: gotestsum ## run go unit tests.
	$(GOTESTSUM) $(TEST_ARGS) ./cmd/... ./internal/... -short

test-go-integration: gotestsum ## run go integration tests.
	$(GOTESTSUM) $(TEST_ARGS) ./cmd/... ./internal/... -run Integration

test-go-e2e: gotestsum ## run go e2e tests.
	$(GOTESTSUM) $(TEST_ARGS) ./test/e2e/... -run E2E

test-k6-performance: ## run k6 performance tests.
	K6_WEB_DASHBOARD=true K6_WEB_DASHBOARD_EXPORT=test-reports/performance/report.html $(K6) run -e NAMESPACE=$(NAMESPACE) --out json=test-reports/performance/report.json test/performance/*_test.js

##@ Build

.PHONY: web-build
web-build: ng ## build the web components.
	npm install --prefix web
	uhsApiURL=$(strip $(call uhs_api_url)) yunikornApiURL=$(strip $(call yunikorn_api_url)) npm run setenv --prefix web
	npm run build --prefix web

.PHONY: build
build: bin/app ## build the unicorn-history-server binary for current OS and architecture.
	echo "Building unicorn-history-server binary for $(OS)/$(ARCH)"
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -o $(LOCALBIN_APP)/unicorn-history-server 										\
		-ldflags "-X github.com/G-Research/unicorn-history-server/cmd/unicorn-history-server/info.Version=$(GIT_TAG) 		\
				  -X github.com/G-Research/unicorn-history-server/cmd/unicorn-history-server/info.Commit=$(GIT_COMMIT) 	\
				  -X github.com/G-Research/unicorn-history-server/cmd/unicorn-history-server/info.BuildTime=$(BUILD_TIME)" \
	  	./cmd/unicorn-history-server

.PHONY: build-linux-amd64
build-linux-amd64: ## build the unicorn-history-server binary for linux/amd64.
	OS=linux ARCH=amd64 $(MAKE) build

.PHONY: clean
clean: ## remove generated build artifacts.
	rm -rf $(LOCALBIN_APP)

##@ Publish
NODE_VERSION ?= 20
ALPINE_VERSION ?= 3.20
DOCKER_OUTPUT ?= type=docker
DOCKER_TAGS ?= $(IMAGE_TAG)
ifneq ($(origin DOCKER_METADATA), undefined)
  # If DOCKER_METADATA is defined, use it to set the tags and labels.
  # DOCKER_METADATA should be a JSON object with the following structure:
  # {
  #   "tags": ["image:tag1", "image:tag2"],
  #   "labels": {
  #     "label1": "value1",
  #     "label2": "value2"
  #   }
  # }
  DOCKER_TAGS=$(shell echo $$DOCKER_METADATA | jq -r '.tags | map("--tag \(.)") | join(" ")')
  DOCKER_LABELS=$(shell echo $$DOCKER_METADATA | jq -r '.labels | to_entries | map("--label \(.key)=\"\(.value)\"") | join(" ")')
else
  # Otherwise, use DOCKER_TAGS if defined, otherwise use the default.
  # DOCKER_TAGS should be a space-separated list of tags.
  # e.g. DOCKER_TAGS="image:tag1 image:tag2"
  # We do not set DOCKER_LABELS because of the way make handles spaces
  # in variable values. Use DOCKER_METADATA if you need to set labels.
  DOCKER_TAGS?=unicorn-history-server:$(VERSION) unicorn-history-server:latest
  DOCKER_TAGS:=$(addprefix --tag ,$(DOCKER_TAGS))
endif

.PHONY: docker-build
docker-build: OS=linux
docker-build: bin/docker clean build ## build docker image using buildx.
	echo "Building docker image for linux/$(ARCH)"
	docker buildx build    				     			 \
		--file build/unicorn-history-server/Dockerfile  \
		--platform linux/$(ARCH) 		 				 \
		--output $(DOCKER_OUTPUT) 						 \
		--build-arg NODE_VERSION=$(NODE_VERSION) 		 \
        --build-arg ALPINE_VERSION=$(ALPINE_VERSION)     \
		$(DOCKER_TAGS) 		   						  	 \
		.

.PHONY: docker-build-tarball
docker-build-tarball: DOCKER_OUTPUT=type=oci,dest=$(LOCALBIN_DOCKER)/$(IMAGE_NAME)-oci-$(ARCH).tar
docker-build-tarball: docker-build ## build docker images and save them as tarballs.

.PHONY: docker-build-amd64
docker-build-amd64: ## build docker image for linux/amd64.
	OS=linux ARCH=amd64 $(MAKE) docker-build

.PHONY: docker-push
docker-push: PUSH=--push
docker-push: docker-build-amd64 ## push linux/amd64 docker image to registry using buildx.

##@ External Dependencies

.PHONY: kind-all-local
kind-all-local: kind-all helm-install-uhs-local ## create kind cluster, install dependencies locally and build & install unicorn-history-server.

.PHONY: kind-all
kind-all minikube-all: create-cluster install-dependencies migrate-up ## create cluster and install dependencies.

.PHONY: create-cluster
create-cluster: $(KIND) $(MINIKUBE) ## create a cluster.
ifeq ($(strip $(CLUSTER_MGR)),kind)
	$(KIND) create cluster --name $(CLUSTER_NAME) --config hack/kind-config.yml
else
	$(MINIKUBE) start --ports=30000:30000 --ports=30001:30001 --ports=30002:30002 --ports=30003:30003
endif

.PHONY: delete-cluster
delete-cluster: $(KIND) $(MINIKUBE) ## delete the cluster.
ifeq ($(strip $(CLUSTER_MGR)),kind)
	$(KIND) delete cluster --name $(CLUSTER_NAME)
else
	$(MINIKUBE) delete
endif

.PHONY: kind-load-image
kind-load-image: docker-build-amd64 ## inject the local docker image into the kind cluster.
	kind load docker-image $(IMAGE_TAG) --name $(CLUSTER_NAME)

.PHONY: install-dependencies
install-dependencies: helm-repos install-and-patch-yunikorn helm-install-postgres wait-for-dependencies ## install dependencies.

.PHONY: wait-for-dependencies
wait-for-dependencies: ## wait for dependencies to be ready.
	hack/wait-for-dependencies.sh

.PHONY: install-and-patch-yunikorn
install-and-patch-yunikorn: helm-install-yunikorn patch-yunikorn-service ## install yunikorn and patch Service to expose NodePorts.

.PHONY: helm-install-yunikorn
helm-install-yunikorn: ## install yunikorn using helm.
	$(HELM) upgrade --install yunikorn yunikorn/yunikorn --namespace $(NAMESPACE) --create-namespace

.PHONY: helm-uninstall-yunikorn
helm-uninstall-yunikorn: ## uninstall yunikorn using helm.
	$(HELM) uninstall yunikorn --namespace $(NAMESPACE)

.PHONY: helm-install-postgres
helm-install-postgres: ## install postgres using helm.
	$(HELM) upgrade --install postgresql bitnami/postgresql --values hack/postgres.values.yaml \
		--namespace $(NAMESPACE) --create-namespace

.PHONY: helm-uninstall-postgres
helm-uninstall-postgres: ## uninstall postgres using helm.
	$(HELM) uninstall postgres --namespace $(NAMESPACE)

.PHONY: helm-install-uhs-local
helm-install-uhs-local: kind-load-image ## build & install unicorn-history-server using helm.
	helm upgrade --install unicorn-history-server charts/unicorn-history-server \
		--set image.registry=""   			 \
		--set image.repository=$(IMAGE_REPO) \
		--set image.tag=$(GIT_TAG) 			 \
		--set service.type=NodePort 		 \
		--namespace $(NAMESPACE)  			 \
		--create-namespace

helm-uninstall-uhs-local:
	helm uninstall unicorn-history-server --namespace $(NAMESPACE)

.PHONY: helm-repos
helm-repos: helm
	$(HELM) repo add gresearch https://g-research.github.io/charts
	helm repo add yunikorn https://apache.github.io/yunikorn-release
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami
	$(HELM) repo update

##@ Utils

.PHONY: patch-yunikorn-service
patch-yunikorn-service: ## patch yunikorn service to expose it as NodePort (yunikorn-core@30000, yunikorn-service@30001).
	hack/patch-yunikorn-service.sh

##@ Build Dependencies

.PHONY: install-tools
install-tools: golangci-lint gotestsum $(CLUSTER_MGR) helm yq ## install development tools.

GOTESTSUM ?= $(LOCALBIN_TOOLING)/gotestsum
GOTESTSUM_VERSION ?= v1.11.0
.PHONY: gotestsum
gotestsum: $(GOTESTSUM) ## download gotestsum locally if necessary.
$(GOTESTSUM): bin/tooling
	test -s $(GOTESTSUM) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)

GORELEASER ?= $(LOCALBIN_TOOLING)/goreleaser
GORELEASER_VERSION ?= v1.26.2
.PHONY: goreleaser
goreleaser: $(GORELEASER) ## download GoReleaser locally if necessary.
$(GORELEASER): bin/tooling
	test -s $(GORELEASER) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)

GOLANGCI_LINT ?= $(LOCALBIN_TOOLING)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.60.2
.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## download golangci-lint locally if necessary.
$(GOLANGCI_LINT): bin/tooling
	test -s $(GOLANGCI_LINT) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

GOMIGRATE ?= $(LOCALBIN_TOOLING)/migrate
GOMIGRATE_VERSION ?= v4.17.1
.PHONY: gomigrate
gomigrate: $(GOMIGRATE) ## download gomigrate locally if necessary.
$(GOMIGRATE): bin/tooling
	test -s $(GOMIGRATE) || curl --silent -L https://github.com/golang-migrate/migrate/releases/download/$(GOMIGRATE_VERSION)/migrate.$(OS)-$(ARCH).tar.gz | tar xvz -C $(LOCALBIN_TOOLING)

HELM ?= $(LOCALBIN_TOOLING)/helm
HELM_VERSION ?= v3.15.3
.PHONY: helm
.ONESHELL:
helm: $(HELM) ## Download helm locally if necessary.
$(HELM): bin/tooling
	if [ ! -s $(HELM) ]; then \
		curl --silent -L https://get.helm.sh/helm-$(HELM_VERSION)-$(OS)-$(ARCH).tar.gz | tar xvzf - ; \
		mv $(OS)-$(ARCH)/helm $(LOCALBIN_TOOLING) ; \
		rm -r $(OS)-$(ARCH) ; \
	fi

YQ ?= $(LOCALBIN_TOOLING)/yq
YQ_VERSION ?= v4.44.2
.PHONY: yq
yq: $(YQ) ## download gomigrate locally if necessary.
$(YQ): bin/tooling
	test -s $(YQ) || curl --silent -L https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$(OS)_$(ARCH) -o $(LOCALBIN_TOOLING)/yq
	chmod +x $(LOCALBIN_TOOLING)/yq

MOCKGEN ?= $(LOCALBIN_TOOLING)/mockgen
MOCKGEN_VERSION ?= v0.4.0
.PHONY: mockgen
mockgen: $(MOCKGEN) ## Download mockgen locally if necessary.
$(MOCKGEN): bin/tooling
	test -s $(MOCKGEN) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install go.uber.org/mock/mockgen@$(MOCKGEN_VERSION)

.PHONY: kind
kind: $(KIND) ## download kind locally if necessary.
$(KIND): bin/tooling
	test -s $(KIND) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install sigs.k8s.io/kind@$(KIND_VERSION)

.PHONY: minikube
minikube: $(MINIKUBE) ## Download minikube locally if necessary.
$(MINIKUBE): bin/tooling
	test -s $(MINIKUBE) || \
	curl --silent -L https://storage.googleapis.com/minikube/releases/$(MINIKUBE_VERSION)/minikube-$(OS)-$(ARCH) \
		-o $(LOCALBIN_TOOLING)/minikube
	chmod +x $(LOCALBIN_TOOLING)/minikube

XK6 ?= $(LOCALBIN_TOOLING)/xk6
K6 ?= $(LOCALBIN_TOOLING)/k6
K6_VERSION ?= v0.52.0

.PHONY: xk6
xk6: $(XK6) ## download xk6 locally if necessary.
$(XK6): bin/tooling
	test -s $(XK6) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install go.k6.io/xk6/cmd/xk6@latest

.PHONY: k6
k6: xk6 $(K6) ## download k6 locally if necessary.
$(K6): bin/tooling
	test -s $(K6) || $(XK6) build $(K6_VERSION) --with github.com/grafana/xk6-kubernetes --output $(K6)

.PHONY: ng
ng: ## install Angular CLI.
	npm install -g @angular/cli@18
