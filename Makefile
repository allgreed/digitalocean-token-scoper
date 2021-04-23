SOURCES=main.go
ENTRYPOINT=main.go

# Porcelain
# ###############
.PHONY: container run build lint test

run: setup ## run the app
	go run $(ENTRYPOINT)

run-watch: setup ## run the app in dev mode
	ls $(SOURCES) | entr -cr make run

build: setup main.out ## create artifact

lint: setup ## run static analysis
	@echo "Not implemented"; false

test: setup ## run all tests
	@echo "Not implemented"; false

container: build ## create container
	#docker build -t lmap .
	@echo "Not implemented"; false

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
