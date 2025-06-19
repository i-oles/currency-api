run:
	go run cmd/currencyapi/main.go

lint:
	golangci-lint run

test:
	go test -v ./...

build:
	go build -v -o bin/currency-api cmd/currencyapi/main.go