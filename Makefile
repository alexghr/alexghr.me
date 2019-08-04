MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := images


.PHONY: images
images:
	$(MAKE) -C apex latest
	$(MAKE) -C www latest

.PHONY: ecs_images
ecs_images:
	$(MAKE) -C apex push_latest
	$(MAKE) -C www push_latest


# Local Docker Compose setup
DOCKER_COMPOSE ?= docker-compose
DOCKER_COMPOSE_COMMAND := $(DOCKER_COMPOSE) --file docker-compose.yml --file docker-compose.local.yml

.PHONY: up
up:
	$(DOCKER_COMPOSE_COMMAND) up

.PHONY: down
down:
	$(DOCKER_COMPOSE_COMMAND) down


# ECS setup
ECS_CLUSTER ?=
ECS_SERVICE ?=

ECS ?= ecs-cli
ECS_COMMAND := $(ECS) compose --project-name $(ECS_SERVICE) --cluster $(ECS_CLUSTER) --file docker-compose.yml --file docker-compose.aws.yml

.PHONY: ecs_deploy
ecs_deploy:
	$(ECS_COMMAND) service stop
	$(ECS_COMMAND) service up --force-deployment
