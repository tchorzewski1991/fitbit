APP=fitbit

build:
	@go build -o $(APP) -ldflags '-X main.build=local' .

run:
	@go run main.go

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