run:
	go run cmd/currencyapi/main.go

lint:
	golangci-lint run

test:
	go test -v ./...