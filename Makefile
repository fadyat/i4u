ifneq (,$(wildcard ./.env))
	include .env
	export
endif

cli:
	@go build -o bin/i4u cmd/i4u/main.go
	@chmod +x i4u.sh

lint:
	@golangci-lint run --issues-exit-code 1 --print-issued-lines=true --config .golangci.yml ./...

test:
	@go test -v -cover ./...

mocks:
	@mockery --dir api --output mocks --filename summary.go --name Summarizer

.PHONY: cli lint mocks test