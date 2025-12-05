run:
	go run main.go

test:
	go test ./...

lint:
	go vet ./...

schema:
	@echo "Schema validation handled in code tests"

