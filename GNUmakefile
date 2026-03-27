default: build

build:
	go build -o terraform-provider-mcs

install: build
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/pinkroccade/mcs/0.1.0/$$(go env GOOS)_$$(go env GOARCH)
	cp terraform-provider-mcs ~/.terraform.d/plugins/registry.terraform.io/pinkroccade/mcs/0.1.0/$$(go env GOOS)_$$(go env GOARCH)/

test:
	go test ./... -v -count=1

testacc:
	TF_ACC=1 go test ./... -v -count=1

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

generate:
	go generate ./...

.PHONY: build install test testacc lint fmt generate
