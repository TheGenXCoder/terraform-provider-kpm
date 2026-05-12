default: build

build:
	GOWORK=off go build ./...

test:
	GOWORK=off go test ./... -v -count=1

testacc:
	GOWORK=off TF_ACC=1 go test ./... -v -count=1 -timeout 120m

install:
	GOWORK=off go install .

lint:
	GOWORK=off golangci-lint run ./...

.PHONY: build test testacc install lint
