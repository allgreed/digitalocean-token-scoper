PORT=8080
CLIENT_SECRET=aaaa
CONTAINER_NAME=do-token-scoper-companion
CONTAINER_PORT=5678

SOURCES=main.go rules.go utils.go config.go
TESTS=rules_test.go
# TODO: use magic functions to find all sources and tests
DOCKER_PROJECT := allgreed/digitalocean-token-scoper

LINTFLAGS := -e -s

# Porcelain
# ###############
.PHONY: container run build lint test env-up env-down test-watch secrets

run: secrets setup ## run the app
	# TODO: ensure that env is running?
	APP_LOG_LEVEL=debug APP_LOG_FORMAT=text APP_PORT=$(PORT) APP_TARGET_URL=http://localhost:$(CONTAINER_PORT) APP_USERTOKEN__allgreed=./secrets/users/allgreed/secret APP_TOKEN_PATH=./secrets/token/secret APP_USERTOKEN__dawid=./secrets/users/dawid/secret go run -mod=readonly $(SOURCES)

run-watch: setup ## run the app in dev mode, hot reloading
	ls $(SOURCES) Makefile | entr -cr make run

build: setup ## create artifact
	nix-build

lint: setup ## run static analysis
	gofmt $(LINTFLAGS) -w $(SOURCES)

lint-check: setup ## run static analysis - for CI
	test -z "$$(gofmt $(LINTFLAGS) -l .)"

test: setup ## run all tests
	go test -mod=readonly

test-watch: setup ## run tests in watch mode
	ls $(SOURCES) $(TESTS) | entr -c make test

env-up: ## set up dev environment
	docker run -d --name $(CONTAINER_NAME) --restart=unless-stopped -p $(CONTAINER_PORT):80 ealen/echo-server:0.5.0 --enable:environment false --enable:host
	sleep 2

env-down: ## tear down dev environment
	docker rm -f $(CONTAINER_NAME)

container: setup ## create container
	nix-build -A docker.image

interact: ## helper process to run predefined inputs
	# TODO: simple command runner with a few options that can be chosen at a keypress
	curl http://localhost:$(PORT)/v2/domains/olgierd.space/records --silent -H "Authorization: Bearer $(CLIENT_SECRET)" | jq

# Plumbing
# ###############
.PHONY: setup gitclean gitclean-with-libs secrets

secrets: secrets/token/secret secrets/users/allgreed/secret secrets/users/dawid/secret

main.out: $(SOURCES)
	go build -o $< $@

setup:

gitclean:

gitclean-with-libs:

secrets/token/secret:
	mkdir -p secrets/token
	echo "verymuchanexampletoken" > $@

secrets/users/allgreed/secret:
	mkdir -p secrets/users/allgreed
	echo "$(CLIENT_SECRET)" > $@

secrets/users/dawid/secret:
	mkdir -p secrets/users/dawid
	echo "very much whatever" > $@

# Helpers
# ###############
.PHONY: container-full prepare-release-image-tag

container-full: setup container ## create and load the image to the local repository
	docker load < result

ACQUIRE_VERSION_FORM_DEFAULT_NIX := grep 'version =' default.nix | cut -d' ' -f 5 | tr -d '\"\;'

check-version-uploaded:
	curl --silent -f -lSL https://index.docker.io/v1/repositories/$(DOCKER_PROJECT)/tags/$$($(ACQUIRE_VERSION_FORM_DEFAULT_NIX)) > /dev/null && false || true

prepare-release-image-tag:
	$(ACQUIRE_VERSION_FORM_DEFAULT_NIX) > .tags

# Utilities
# ###############
.PHONY: help todo clean really_clean init
init: ## one time setup
	direnv allow .

todo: ## list all TODOs in the project
	git grep -I --line-number TODO | grep -v 'list all TODOs in the project' | grep TODO

clean: gitclean ## remove artifacts
	rm main.out

really_clean: clean gitclean-with-libs ## remove EVERYTHING

help: ## print this message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
.DEFAULT_GOAL := help
