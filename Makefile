ifneq (,$(wildcard ./.env))
	include .env
	export
endif

cli:
	@go build -o bin/i4u cmd/i4u/main.go
