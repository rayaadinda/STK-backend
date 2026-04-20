APP_NAME=stk-backend

.PHONY: run build tidy test

run:
	go run ./cmd/server

build:
	go build -o bin/$(APP_NAME) ./cmd/server

tidy:
	go mod tidy

test:
	go test ./...
