.PHONY: tidy fmt vet test ci

tidy:
	go mod tidy
	git diff --exit-code

fmt:
	go fmt ./...
	git diff --exit-code

vet:
	go vet ./...

test:
	go test -race -v ./...

ci: tidy fmt vet test
