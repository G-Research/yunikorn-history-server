
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

# Go compiler selection
ifeq ($(GO),)
GO := go
endif

GO_VERSION := $(shell "$(GO)" version | awk '{print substr($$3, 3, 4)}')
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

# Make sure we are in the same directory as the Makefile
BASE_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))
TOOLS_DIR=tools

# Force Go modules even when checked out inside GOPATH
GO111MODULE := on
export GO111MODULE
GO_LDFLAGS=-s -w -X github.com/G-Research/yunikorn-history-server/pkg/version.Version=$(VERSION)

# Build the example binaries for dev and test
.PHONY: commands
commands: build/event-collector

build/event-collector: go.mod go.sum $(shell find cmd internal)
	@echo "building event-collector"
	@mkdir -p build
	"$(GO)" build $(RACE) -a -ldflags '-extldflags "-static"' -o build/event-collector ./cmd/event-collector

# Remove generated build artifacts
.PHONY: clean
clean:
	@echo "cleaning up caches and output"
	"$(GO)" clean -cache -testcache -r
	@echo "removing generated files"
	@rm -rf build

.PHONY: go-lint
go-lint: ## run go linters.
	@echo '>>> Running go linters.'
	@golangci-lint run -v --issues-exit-code 0 ## TODO: remove after fixing all lint issues

.PHONY: install-tools
install-tools: ## install tools.
	@echo '>>> Installing tools.'
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.0

DOCKER_OUTPUT?=type=docker
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
.PHONY: docker-dist
docker-dist: build/event-collector ## build docker image.
	@echo ">>> Building Docker image."
	@docker buildx build --provenance false --sbom false --platform linux/$(shell go env GOARCH) --output $(DOCKER_OUTPUT) $(DOCKER_TAGS) $(DOCKER_LABELS) .

.PHONY: dist
dist: build/event-collector docker-dist ## build the software archives.
