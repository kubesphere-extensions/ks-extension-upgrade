REPO ?= kubesphere
TAG ?= latest

IMAGE=${REPO}/ks-extension-upgrade:${TAG}

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

all: build

##@ Development

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...


##@ Build

build: ## Build binary.
	go build -o ks-extension-upgrade

docker-build: ## Build docker image.
	docker build -t $(IMAGE) -f Dockerfile .

docker-push:  ## Push docker image. 
	docker push $(IMAGE)

docker-buildx-multi-arch:
	docker buildx build --push --platform=linux/amd64,linux/arm64  -t $(IMAGE) -f Dockerfile .

