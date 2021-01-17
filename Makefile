.PHONY: lint test vendor clean

export GO111MODULE=on

default: lint test

lint:
	golangci-lint run

run:
	go run SAMLAuthnRequester.go

test:
	go test -v -cover ./...

vendor:
	go mod vendor

build-local:
	./build.sh local

build-ecr:
	./build.sh ecr

clean:
	rm -rf ./vendor