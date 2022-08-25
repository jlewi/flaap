tidy-go:
	cd go && gofmt -s -w .
	cd go && goimports -w .
	
lint-go:
	# golangci-lint automatically searches up the root tree for configuration files.
	cd go && golangci-lint run

test-go:
	cd go && go test -v ./...

test: test-go