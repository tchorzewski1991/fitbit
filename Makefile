APP=fitbit

build:
	@go build -o $(APP) -ldflags '-X main.build=local' ./app/services/node/main.go

run:
	@go run ./app/services/node/main.go

tidy:
	@go mod tidy
	@go mod verify
	@go mod vendor

test:
	@go test ./...

clean:
	@echo "  >  Cleaning build cache"
	@GOPATH=$(GOPATH) GOBIN=$(GOBIN) go clean
	@echo "  >  Removing $(APP) executable"
	@rm $(APP) 2> /dev/null | true

lint:
	@golangci-lint run -v -c golangci.yaml

wallet:
	@go run ./app/wallet/cli/main.go --help

wallet-version:
	@go run ./app/wallet/cli/main.go version

wallet-generate:
	@go run ./app/wallet/cli/main.go generate
