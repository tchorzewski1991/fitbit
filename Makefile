APP=fitbit

build:
	@go build -o $(APP) -ldflags '-X main.build=local' ./app/services/node/main.go

run-primary:
	@go run ./app/services/node/main.go

run-second-node:
	@go run ./app/services/node/main.go \
		--node-public-host 0.0.0.0:3001 \
		--node-private-host 0.0.0.0:4001 \
		--state-beneficiary rxtx \
		--state-data-path data/miner2

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
