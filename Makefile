.PHONY: build install fmt vet test

build:
	go build -o bin/lgtm .

install:
	go install .

fmt:
	gofmt -l .

vet:
	go vet ./...

test:
	go test ./...
