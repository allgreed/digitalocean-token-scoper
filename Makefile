SOURCES=main.go auth.go
PORT=8080

CONTAINER_NAME=do-token-scoper-companion
CONTAINER_PORT=5678

# Porcelain
# ###############
.PHONY: container run build lint test env-up env-down

run: setup ## run the app
	# TODO: ensure that env is running?
	APP_PORT=$(PORT) APP_TARGET_URL=http://localhost:$(CONTAINER_PORT) go run $(SOURCES)

run-watch: setup ## run the app in dev mode
	ls $(SOURCES) Makefile | entr -cr make run

build: setup main.out ## create artifact

lint: setup ## run static analysis
	go fmt $(SOURCES)

test: setup ## run all tests
	@echo "Not implemented"; false

env-up: ## set up dev environment
	docker run -d --name $(CONTAINER_NAME) --restart=unless-stopped -p $(CONTAINER_PORT):5678 hashicorp/http-echo:0.2.3 -text="hello world"
	sleep 2

env-down: ## tear down dev environment
	docker rm -f $(CONTAINER_NAME)

container: build ## create container
	#docker build -t lmap .
	@echo "Not implemented"; false

interact: ## helper process to run predefined inputs
	# TODO: simple command runner with a few options that can be chosen at a keypress
	curl localhost:$(PORT) -H "Authorization: aaaa"

# Plumbing
# ###############
.PHONY: setup gitclean gitclean-with-libs

main.out: $(SOURCES)
	go build -o $< $@

setup:

gitclean:

gitclean-with-libs:

# Helpers
# ###############
.PHONY:

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
