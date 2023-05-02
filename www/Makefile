MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all

image_name := www.alexghr.me
IMAGE_TAG ?= $(shell git rev-parse --short HEAD)

DOCKER_REPO ?=
remote_image_name := $(DOCKER_REPO)/$(image_name)

DOCKER ?= docker

.PHONY: all
all: latest

.PHONY: build
build:
	$(DOCKER) build \
		--rm \
		--quiet \
		--build-arg GOOGLE_SITE_VERIFICATION="$(GOOGLE_SITE_VERIFICATION)" \
		--tag $(image_name):$(IMAGE_TAG) \
		.

.PHONY: latest
latest: build
	$(DOCKER) tag $(image_name):$(IMAGE_TAG) $(image_name):latest


.PHONY: push_latest
push_latest: latest
	@$(shell if  [ -z "$(DOCKER_REPO)" ]; then echo "DOCKER_REPO env var not set"; exit 1; fi)
	$(DOCKER) tag $(image_name):latest $(remote_image_name):latest
	$(DOCKER) tag $(image_name):latest $(remote_image_name):$(IMAGE_TAG)
	$(DOCKER) push $(remote_image_name):latest
	$(DOCKER) push $(remote_image_name):$(IMAGE_TAG)
