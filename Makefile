fmt:
	go fmt ./...

vet: fmt
	go vet ./...

test: vet
	go test ./...

run: test
	go run main/main.go

# Targets that do not represent actual files
.PHONY: fmt test vet run
