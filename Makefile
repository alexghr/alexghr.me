MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := images

ECS := ecs-cli
ECS_CLUSTER ?=
ECS_SERVICE ?=

.PHONY: images
images:
	$(MAKE) -C apex push_latest
	$(MAKE) -C www push_latest

.PHONY: ecs_deploy
ecs_deploy:
	$(ECS) compose --project-name $(ECS_SERVICE) service stop --cluster $(ECS_CLUSTER)
	$(ECS) compose --project-name $(ECS_SERVICE) service start --cluster $(ECS_CLUSTER)

.PHONY: ecs_up
ecs_up:
	$(ECS) compose --project-name $(ECS_SERVICE) service up --cluster $(ECS_CLUSTER)

.PHONY: ecs_down
ecs_down:
	$(ECS) compose --project-name $(ECS_SERVICE) service down --cluster $(ECS_CLUSTER)

