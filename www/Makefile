MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all

image_name := www.alexghr.me
IMAGE_TAG ?= $(shell git rev-parse --short HEAD)

REMOTE_REPO ?=
remote_image_name := $(REMOTE_REPO)/$(image_name)

DOCKER ?= docker

.PHONY: all
all: latest

.PHONY: build
build:
	$(DOCKER) build \
		--rm \
		--quiet \
		--tag $(image_name):$(IMAGE_TAG) \
		.

.PHONY: latest
latest: build
	$(DOCKER) tag $(image_name):$(IMAGE_TAG) $(image_name):latest


.PHONY: push_latest
push_latest: latest
	@$(shell if  [ -z "$(REMOTE_REPO)" ]; then echo "REMOTE_REPO env var not set"; exit 1; fi)
	$(DOCKER) tag $(image_name):latest $(remote_image_name):latest
	$(DOCKER) tag $(image_name):latest $(remote_image_name):$(IMAGE_TAG)
	$(DOCKER) push $(remote_image_name):latest
	$(DOCKER) push $(remote_image_name):$(IMAGE_TAG)
