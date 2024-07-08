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
IMAGE_NAME := yunikorn-history-server
# IMAGE_REPO defines the image repository and name where the operator image will be pushed.
IMAGE_REPO ?= $(IMAGE_REGISTRY)/$(IMAGE_NAME)
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

# Docker image config
BASE_IMAGE ?= alpine
BASE_IMAGE_TAG ?= 3.20

# Local Development
KIND_CLUSTER ?= yhs
NAMESPACE ?= yunikorn

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
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Database

YHS_CONFIG ?= config/yunikorn-history-server/local.yml

define yq_get
    $(YQ) e '$(1)' $(YHS_CONFIG)
endef

define yq_get_db
    $(shell $(call yq_get, .db.$(1)))
endef

define database_url
	postgres://$(strip $(call yq_get_db,user)):$(strip $(call yq_get_db,password))@$(strip $(call yq_get_db,host)):$(strip $(call yq_get_db,port))/$(strip $(call yq_get_db,dbname))?sslmode=disable
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
codegen: gomock ## generate code using go generate (mocks).
	go generate ./...

##@ Run

.PHONY: run
run: ## run the yunikorn-history-server binary.
	go run cmd/yunikorn-history-server/main.go --config config/yunikorn-history-server/local.yml

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

.PHONY: test
test: test-go-unit integration-tests ## run all tests.

.PHONY: integration-tests
integration-tests: ## start dependencies and run integration tests.
	hack/run-integration-tests.sh

.PHONY: test-go-unit
test-go-unit: gotestsum ## run go unit tests.
	$(GOTESTSUM) -- ./cmd/... ./internal/... -short -coverprofile operator.out

test-go-integration: gotestsum ## run go integration tests.
	$(GOTESTSUM) -- ./cmd/... ./internal/... -run Integration -coverprofile operator.out

test-go-e2e: gotestsum ## run go e2e tests.
	$(GOTESTSUM) -- ./test/e2e/... -coverprofile operator.out

##@ Build

.PHONY: build
build: bin/app ## build the yunikorn-history-server binary for current OS and architecture.
	echo "Building yunikorn-history-server binary for $(OS)/$(ARCH)"
	GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -o $(LOCALBIN_APP)/yunikorn-history-server ./cmd/yunikorn-history-server

.PHONY: build-linux-amd64
build-linux-amd64: ## build the yunikorn-history-server binary for linux/amd64.
	OS=linux ARCH=amd64 make build

.PHONY: clean
clean: ## remove generated build artifacts.
	rm -rf $(LOCALBIN_APP)

##@ Publish

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
  DOCKER_TAGS?=yunikorn-history-server:$(VERSION) yunikorn-history-server:latest
  DOCKER_TAGS:=$(addprefix --tag ,$(DOCKER_TAGS))
endif

.PHONY: docker-build
docker-build: OS=linux
docker-build: bin/docker clean build copy-build-files ## build docker image using buildx.
	echo "Building docker image for linux/$(ARCH)"
	docker buildx build    				     			 \
		--build-arg BASE_IMAGE=$(BASE_IMAGE) 			 \
		--build-arg BASE_IMAGE_TAG=$(BASE_IMAGE_TAG) 	 \
		--file build/yunikorn-history-server/Dockerfile  \
		--platform linux/$(ARCH) 		 				 \
		--output $(DOCKER_OUTPUT) 						 \
		$(DOCKER_TAGS) 		   						  	 \
		$(LOCALBIN_APP)

.PHONY: docker-build-tarball
docker-build-tarball: DOCKER_OUTPUT=type=oci,dest=$(LOCALBIN_DOCKER)/$(IMAGE_NAME)-oci-$(ARCH).tar
docker-build-tarball: docker-build ## build docker images and save them as tarballs.

.PHONY: docker-build-amd64
docker-build-amd64: ## build docker image for linux/amd64.
	OS=linux ARCH=amd64 make docker-build

.PHONY: copy-build-files
copy-build-files: ## copy required build files to local bin directory.
	cp config/yunikorn-history-server/config.yml $(LOCALBIN_APP)/config.yml
	cp -r migrations $(LOCALBIN_APP)/migrations

.PHONY: docker-push
docker-push: PUSH=--push
docker-push: docker-build-amd64 ## push linux/amd64 docker image to registry using buildx.

##@ External Dependencies

kind-all: kind-create-cluster install-dependencies migrate-up ## create kind cluster and install dependencies

.PHONY: kind-create-cluster
kind-create-cluster: kind ## create a kind cluster.
	$(KIND) create cluster --name $(KIND_CLUSTER) --config hack/kind-config.yml

.PHONY: kind-delete-cluster
kind-delete-cluster: kind ## delete the kind cluster.
	$(KIND) delete cluster --name $(KIND_CLUSTER)

.PHONY: install-dependencies
install-dependencies: helm-repos install-and-patch-yunikorn helm-install-postgres wait-for-dependencies ## install dependencies.

.PHONY: wait-for-dependencies
wait-for-dependencies: ## wait for dependencies to be ready.
	hack/wait-for-dependencies.sh

.PHONY: install-and-patch-yunikorn
install-and-patch-yunikorn: helm-install-yunikorn patch-yunikorn-service ## install yunikorn and patch Service to expose NodePorts.

.PHONY: helm-install-yunikorn
helm-install-yunikorn: ## install yunikorn using helm.
	helm upgrade --install yunikorn yunikorn/yunikorn --namespace $(NAMESPACE) --create-namespace

.PHONY: helm-uninstall-yunikorn
helm-uninstall-yunikorn: ## uninstall yunikorn using helm.
	helm uninstall yunikorn --namespace $(NAMESPACE)

.PHONY: helm-install-postgres
helm-install-postgres: ## install postgres using helm.
	helm upgrade --install postgresql bitnami/postgresql --values hack/postgres.values.yaml --namespace $(NAMESPACE) --create-namespace

.PHONY: helm-uninstall-postgres
helm-uninstall-postgres: ## uninstall postgres using helm.
	helm uninstall postgres --namespace $(NAMESPACE)

.PHONY: helm-repos
helm-repos:
	helm repo add yunikorn https://apache.github.io/yunikorn-release
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo update

##@ Utils

.PHONY: patch-yunikorn-service
patch-yunikorn-service: ## patch yunikorn service to expose it as NodePort (yunikorn-core@30000, yunikorn-service@30001).
	hack/patch-yunikorn-service.sh

##@ Build Dependencies

.PHONY: install-tools
install-tools: golangci-lint gotestsum kind yq ## install development tools.

GOTESTSUM ?= $(LOCALBIN_TOOLING)/gotestsum
GOTESTSUM_VERSION ?= v1.11.0
.PHONY: gotestsum
gotestsum: $(GOTESTSUM) ## Download gotestsum locally if necessary.
$(GOTESTSUM): bin/tooling
	test -s $(GOTESTSUM) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install gotest.tools/gotestsum@$(GOTESTSUM_VERSION)

GORELEASER ?= $(LOCALBIN_TOOLING)/goreleaser
GORELEASER_VERSION ?= v1.26.2
.PHONY: goreleaser
goreleaser: $(GORELEASER) ## Download GoReleaser locally if necessary.
$(GORELEASER): bin/tooling
	test -s $(GORELEASER) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)

GOLANGCI_LINT ?= $(LOCALBIN_TOOLING)/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.59.0
.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT) ## Download golangci-lint locally if necessary.
$(GOLANGCI_LINT): bin/tooling
	test -s $(GOLANGCI_LINT) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

GOMIGRATE ?= $(LOCALBIN_TOOLING)/migrate
GOMIGRATE_VERSION ?= v4.17.1
.PHONY: gomigrate
gomigrate: $(GOMIGRATE) ## Download gomigrate locally if necessary.
$(GOMIGRATE): bin/tooling
	test -s $(GOMIGRATE) || curl --silent -L https://github.com/golang-migrate/migrate/releases/download/$(GOMIGRATE_VERSION)/migrate.$(OS)-$(ARCH).tar.gz | tar xvz -C $(LOCALBIN_TOOLING)

YQ ?= $(LOCALBIN_TOOLING)/yq
YQ_VERSION ?= v4.44.2
.PHONY: yq
yq: $(YQ) ## Download gomigrate locally if necessary.
$(YQ): bin/tooling
	test -s $(YQ) || curl --silent -L https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$(OS)_$(ARCH) -o $(LOCALBIN_TOOLING)/yq
	chmod +x $(LOCALBIN_TOOLING)/yq

GOMOCK ?= $(LOCALBIN_TOOLING)/gomock
GOMOCK_VERSION ?= v0.4.0
.PHONY: gomock
gomock: $(GOMOCK) ## Download uber-go/gomock locally if necessary.
$(GOMOCK): bin/tooling
	test -s $(YQ) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install go.uber.org/mock/mockgen@$(GOMOCK_VERSION)

KIND ?= $(LOCALBIN_TOOLING)/kind
KIND_VERSION ?= v0.23.0
.PHONY: kind
kind: $(KIND) ## Download kind locally if necessary.
$(KIND): bin/tooling
	test -s $(KIND) || GOBIN=$(LOCALBIN_TOOLING) $(GO) install sigs.k8s.io/kind@$(KIND_VERSION)
