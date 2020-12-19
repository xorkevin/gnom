## PROLOG

.PHONY: help all

CMDNAME=gnom
CMDDESC=a simple lexer and parser

help: ## Print this help
	@./help.sh '$(CMDNAME)' '$(CMDDESC)'

all: test ## Default

## TESTS

TEST_ARGS=
COVERAGE=cover.out
COVERAGE_HTML=coverage_report.html
COVERAGE_ARGS=-covermode count -coverprofile $(COVERAGE)
BENCHMARK_ARGS=-benchtime 5s -benchmem

.PHONY: test coverage cover bench

test: ## Run tests
	go test $(TEST_ARGS) -cover $(COVERAGE_ARGS) ./...

coverage: ## View test coverage
	go tool cover -html $(COVERAGE) -o $(COVERAGE_HTML)

cover: test coverage ## Create coverage report

bench: ## Run benchmarks
	go test -bench . $(BENCHMARK_ARGS)

## FMT

.PHONY: fmt vet prepare

fmt: ## Run go fmt
	go fmt ./...

vet: ## Lint code
	go vet ./...

prepare: fmt vet ## Prepare code for PR
