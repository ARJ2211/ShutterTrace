.PHONY: run build tidy test

run:
	go run ./cmd/prnu --help

build:
	go build -o ShutterTrace ./cmd/prnu

tidy:
	go mod tidy

test:
	go test ./... -v
