test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

fmt:
	gofumpt -l -w .