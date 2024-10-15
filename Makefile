
.PHONY: default
default:
	$(MAKE) -s $(IMAGES)

.PHONY: all
all: default

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-45s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

############################################################################
# Variables
############################################################################

IMAGES ?= router example-target-application vpn-gateway stateless-load-balancer controller-manager network-daemon
VERSION ?= latest

# E2E tests
E2E_FOCUS ?= ""
E2E_SKIP ?= ""
E2E_ENVIRONMENT ?= $(shell pwd)/test/e2e/environment/kind-ipv4/config.yaml
E2E_SEED ?= $(shell shuf -i 1-2147483647 -n1)

UNIT_TEST_DOCKER_PARAMS ?= -it
UNIT_TEST_K8S_VERSION ?= 1.28.0

# Contrainer Registry
REGISTRY ?= localhost:5000/meridio-experiment

GO_MOD := github.com/lioneljouin/meridio-experiment

# Tools
export PATH := $(shell pwd)/bin:$(PATH)
GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GINKGO = $(shell pwd)/bin/ginkgo
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
CLIENT_GEN = $(shell pwd)/bin/client-gen
LISTER_GEN = $(shell pwd)/bin/lister-gen
INFORMER_GEN = $(shell pwd)/bin/informer-gen
GOFUMPT = $(shell pwd)/bin/gofumpt
ENVTEST = $(shell pwd)/bin/setup-envtest
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))

BUILD_DIR ?= build
BUILD_STEPS ?= build tag push
BUILD_CMD ?= build
BUILD_ARGS ?= 
BUILD_REGISTRY ?=

OUTPUT_DIR ?= _output

#############################################################################
# Container: Build, tag, push
#############################################################################

.PHONY: build
build:
	docker $(BUILD_CMD) \
	$(BUILD_ARGS) \
	-t $(BUILD_REGISTRY)$(IMAGE):$(VERSION) \
	--build-arg BUILD_VERSION=$(shell git describe --dirty --tags) \
	-f ./$(BUILD_DIR)/$(IMAGE)/Dockerfile .
.PHONY: tag
tag:
	docker tag $(BUILD_REGISTRY)$(IMAGE):$(VERSION) $(REGISTRY)/$(IMAGE):$(VERSION)
.PHONY: push
push:
	docker push $(REGISTRY)/$(IMAGE):$(VERSION)

#############################################################################
##@ Component (Build, tag, push): use VERSION to set the version. Use BUILD_STEPS to set the build steps (build, tag, push)
#############################################################################

.PHONY: router
router: ## Build the router.
	IMAGE=router $(MAKE) -s $(BUILD_STEPS)

.PHONY: example-target-application
example-target-application: ## Build the example target application.
	BUILD_DIR=examples/target-application/build IMAGE=target-application $(MAKE) $(BUILD_STEPS)

.PHONY: vpn-gateway
vpn-gateway: ## Build the vpn-gateway.
	BUILD_DIR=hack IMAGE=vpn-gateway $(MAKE) $(BUILD_STEPS)

.PHONY: stateless-load-balancer
stateless-load-balancer: ## Build the stateless-load-balancer.
	IMAGE=stateless-load-balancer $(MAKE) $(BUILD_STEPS)

.PHONY: controller-manager
controller-manager: ## Build the controller-manager.
	IMAGE=controller-manager $(MAKE) $(BUILD_STEPS)

.PHONY: network-daemon
network-daemon: ## Build the network-daemon.
	IMAGE=network-daemon $(MAKE) $(BUILD_STEPS)

#############################################################################
##@ Testing & Code check
#############################################################################

.PHONY: lint
lint: golangci-lint ## Run linter against golang code.
	$(GOLANGCI_LINT) run ./...
	
.PHONY: lint-dockerfiles
lint-dockerfiles: ## Run linter against dockerfiles.
	@for image in $(IMAGES); do \
		BUILD_STEPS=lint-dockerfile $(MAKE) -s $${image} ; \
	done

.PHONY: e2e
e2e: ginkgo output-dir ## Run the E2E tests.
	$(GINKGO) -v \
	--no-color --seed=$(E2E_SEED) \
	--repeat=0 --timeout=1h \
	--randomize-all \
	$(shell $(MAKE) -s print-e2e-skip-focus E2E_FOCUS=$(E2E_FOCUS) E2E_SKIP=$(E2E_SKIP)) \
	--json-report=e2e_report.json \
	--junit-report=e2e_report_junit.xml \
	--output-dir=$(OUTPUT_DIR) \
	./test/e2e/... -- --configuration="$(E2E_ENVIRONMENT)"

.PHONY: test
test: output-dir envtest setup-test ## Run the Unit tests (read coverage report: go tool cover -html=_output/cover_unit_test.out -o _output/cover_unit_test.html).
	go test -p 1 -race -cover -short -count=1 -coverprofile $(OUTPUT_DIR)/cover_unit_test.out ./...

.PHONY: setup-test
setup-test:
	$(ENVTEST) use $(UNIT_TEST_K8S_VERSION) --bin-dir $(shell pwd)/bin -p path

.PHONY: check
check: lint test ## Run the linter and the Unit tests.

#############################################################################
##@ Code generation
#############################################################################

.PHONY: generate
generate: gofmt manifests generate-controller generate-client generate-lister generate-informer## Generate all.

.PHONY: gofmt
gofmt: gofumpt ## Run gofumpt.
	$(GOFUMPT) -w .

.PHONY: manifests
manifests: controller-gen ## Generate CustomResourceDefinition objects.
	$(CONTROLLER_GEN) crd paths="./..." output:crd:artifacts:config=deployments/PoC/templates

.PHONY: generate-controller
generate-controller: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# https://github.com/kubernetes/sample-controller
# https://github.com/kubernetes/code-generator
# https://github.com/kubernete<s-sigs/gateway-api/blob/main/hack/update-codegen.sh
# https://github.com/kubernetes-sigs/network-policy-api/blob/main/hack/update-codegen.sh
.PHONY: generate-client
generate-client: output-dir client-gen ## Generate client code
	$(CLIENT_GEN) \
	--clientset-name "versioned" \
	--input-base "$(GO_MOD)" \
	--input "/apis/v1alpha1" \
	--output-dir "pkg/client/clientset" \
	--output-pkg "$(GO_MOD)/pkg/client/clientset" \
	--go-header-file hack/boilerplate.go.txt

.PHONY: generate-lister
generate-lister: lister-gen ## Generate lister code
	$(LISTER_GEN) \
	--output-dir "pkg/client/listers" \
	--output-pkg "$(GO_MOD)/pkg/client/listers" \
	--go-header-file hack/boilerplate.go.txt \
	$(GO_MOD)/apis/v1alpha1

.PHONY: generate-informer
generate-informer: informer-gen ## Generate informer code
	$(INFORMER_GEN) \
	--versioned-clientset-package "${GO_MOD}/pkg/client/clientset/versioned" \
	--listers-package "$(GO_MOD)/pkg/client/listers" \
	--output-dir "pkg/client/informers" \
	--output-pkg "$(GO_MOD)/pkg/client/informers" \
	--go-header-file hack/boilerplate.go.txt \
	$(GO_MOD)/apis/v1alpha1

.PHONY: generate-helm-chart
generate-helm-chart: output-dir ## Generate helm charts.
	helm package ./deployments/PoC --version $(shell $(MAKE) -s format-version VERSION=$(VERSION)) --destination ./_output/helm

#############################################################################
# Tools
#############################################################################

.PHONY: output-dir
output-dir:
	@mkdir -p $(OUTPUT_DIR)

# https://github.com/golangci/golangci-lint
.PHONY: golangci-lint
golangci-lint:
	$(call go-get-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@v1.61.0)

# https://github.com/onsi/ginkgo
.PHONY: ginkgo
ginkgo:
	$(call go-get-tool,$(GINKGO),github.com/onsi/ginkgo/v2/ginkgo@v2.20.2)

.PHONY: controller-gen
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.16.4)

.PHONY: client-gen
client-gen:
	$(call go-get-tool,$(CLIENT_GEN),k8s.io/code-generator/cmd/client-gen@v0.31.1)

.PHONY: lister-gen
lister-gen:
	$(call go-get-tool,$(LISTER_GEN),k8s.io/code-generator/cmd/lister-gen@v0.31.1)

.PHONY: informer-gen
informer-gen:
	$(call go-get-tool,$(INFORMER_GEN),k8s.io/code-generator/cmd/informer-gen@v0.31.1)

.PHONY: gofumpt
gofumpt:
	$(call go-get-tool,$(GOFUMPT),mvdan.cc/gofumpt@v0.7.0)

.PHONY: envtest
envtest:
	$(call go-get-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest@latest)

.PHONY: print-e2e-skip-focus
print-e2e-skip-focus:
	@focus="" ; \
	for f in $(call get_list,$(E2E_FOCUS)); do \
		focus="$${focus} --focus $${f}" ; \
	done ; \
	printf "$${focus}" ; \
	skip="" ; \
	for f in $(call get_list,$(E2E_SKIP)); do \
		skip="$${skip} --skip $${f}" ; \
	done ; \
	printf "$${skip}" 

define get_list
$$(echo "$(1)" | sed -r 's/ //g' | sed -r 's/,/ /g' )
endef

# go-get-tool will 'go get' any package $2 and install it to $1.
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# https://semver.org/#is-there-a-suggested-regular-expression-regex-to-check-a-semver-string
# https://github.com/semver/semver/pull/724
VERSION_REGEX = ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-((0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(\.(0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(\+([0-9a-zA-Z-]+(\.[0-9a-zA-Z-]+)*))?$
.PHONY: format-version
format-version:
	version="$(VERSION)" ; \
	if ! echo "$${version}" | grep -Eq "$(VERSION_REGEX)" ; then \
		version="v0.0.0-$${version}" ; \
	fi ; \
	printf "$${version}"