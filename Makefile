.PHONY: lint test vendor clean

export GO111MODULE=on
EXEC=mon_executable
LOCAL_IMAGE=ohhfrancois/samlauthnrequester:latest


default: lint test

lint:
	golangci-lint run

run:
	go run SAMLAuthnRequester.go

test:
	go test -v -cover ./...

validate:	build-local
	openssl req -x509 -newkey rsa:2048 -keyout certificates/myservice.key -out certificates/myservice.cert -days 365 -nodes -subj "/CN=myservice.test.com"
	docker run -d -v $(PWD)/certificates:/app/certificates --env-file ./docker.envfile --rm $(LOCAL_IMAGE)
	./scripts/push-sp-metadata.sh
	curl -v http://localhost:8090/saml-requester

vendor:
	go mod vendor

build-local:
	./build.sh local

build-ecr:
	./build.sh ecr

clean:
	rm -rf ./vendor