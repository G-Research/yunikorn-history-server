# Variables
SCRIPT ?= event_handler_test.js
NAMESPACE ?= default
YHS_HOST ?= localhost
YHS_PORT ?= 8989
YHS_URL = http://$(YHS_HOST):$(YHS_PORT)
K6 = ./k6
LABEL = app=load-test-by-k6

# Phony Targets
.PHONY: setup-env run cleanup

# Install xk6 and build k6 with kubernetes support
setup-env:
	go install go.k6.io/xk6/cmd/xk6@latest
	xk6 build --with github.com/grafana/xk6-kubernetes

# Run the k6 test script
run:
	$(K6) run -e YHS_URL=$(YHS_URL) -e NAMESPACE=$(NAMESPACE) $(SCRIPT)

# Cleanup the created pods
cleanup:
	kubectl delete pods -l $(LABEL) -n $(NAMESPACE)