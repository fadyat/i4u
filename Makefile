ifneq (,$(wildcard ./.env))
	include .env
	export
endif

cli:
	@go build -o bin/i4u cmd/i4u/main.go

i4u: cli
	@./bin/i4u $(filter-out $@,$(MAKECMDGOALS))

lint:
	@golangci-lint run --issues-exit-code 1 --print-issued-lines=true --config .golangci.yml ./...

test:
	@go clean -testcache
	@go test -v -cover -coverprofile=coverage.out ./...

mocks:
	@mockery --dir api --output mocks --filename summary.go --name Summarizer
	@mockery --dir api --output mocks --filename sender.go --name Sender
	@mockery --dir api --output mocks --filename mail.go --name Mail
	@mockery --dir api --output mocks --filename analyzer.go --name Analyzer

.PHONY: cli lint mocks test