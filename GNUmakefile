default: build

build:
	go build ./...

test:
	go test ./... -v -count=1

testacc:
	TF_ACC=1 go test ./... -v -count=1 -timeout 120m

install:
	go install .

lint:
	golangci-lint run ./...

.PHONY: build test testacc install lint
