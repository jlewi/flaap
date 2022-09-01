tidy-go:
	cd go && gofmt -s -w .
	cd go && goimports -w .
	
lint-go:
	# golangci-lint automatically searches up the root tree for configuration files.
	cd go && golangci-lint run

test-go:
	cd go && go test -v ./...

tidy-py:
  # Sort the imports; black doesn't appear to sort imports
	isort ./py
	# Remove unused imports and other issues
	# Don't reformat the proto files because that appears to remove unused imports which will likely
	# cause problems
	autoflake -r --in-place \
		--exclude *pb2.py,*pb2_grpc.py \
		--remove-unused-variables \
		--remove-all-unused-imports ./py
	black ./py

test-py:
	pytest ./py

protos:
	./build_protos.sh

test: test-go