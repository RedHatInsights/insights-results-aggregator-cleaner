SHELL := /bin/bash

.PHONY: default clean build build-cover shellcheck abcgo lint style run test cover coverage \
    license before_commit help function_list install_addlicense

SOURCES:=$(shell find . -name '*.go')
BINARY:=insights-results-aggregator-cleaner
DOCFILES:=$(addprefix docs/packages/, $(addsuffix .html, $(basename ${SOURCES})))

default: build

clean: ## Run go clean
	@go clean

build: ${BINARY} ## Build binary containing service executable

build-cover:	${SOURCES}  ## Build binary with code coverage detection support
	./build.sh -cover

${BINARY}: ${SOURCES}
	./build.sh

shellcheck: ## Run shellcheck checker
	pre-commit run --all-files shellcheck

abcgo: ## Run ABC metrics checker
	pre-commit run --all-files abcgo

lint: ## Run lint checker
	pre-commit run --all-files golangci-lint-full

style: shellcheck abcgo lint

run: ${BINARY} ## Build the project and executes the binary
	./$^

test: ${BINARY} ## Run the unit tests
	./unit-tests.sh

cover: test ## Display text coverage in Web browser
	@go tool cover -html=coverage.out

coverage: ## Display test coverage in terminal
	@go tool cover -func=coverage.out

license: install_addlicense
	addlicense -c "Red Hat, Inc" -l "apache" -v ./

before_commit: style test integration_tests openapi-check license ## Checks done before commit
	./check_coverage.sh

help: ## Show this help screen
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

function_list: ${BINARY} ## List all functions in generated binary file
	go tool objdump ${BINARY} | grep ^TEXT | sed "s/^TEXT\s//g"

docs/packages/%.html: %.go
	mkdir -p $(dir $@)
	docgo -outdir $(dir $@) $^
	addlicense -c "Red Hat, Inc" -l "apache" -v $@

install_addlicense:
	[[ `command -v addlicense` ]] || go install github.com/google/addlicense
