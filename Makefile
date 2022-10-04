
ROOT := $(shell git rev-parse --show-toplevel)

echo:
	echo ROOT=$(ROOT)

build-dir:
	mkdir -p .build

build-go: build-dir
	cd go && go build -o ../.build/server ./cmd/server/...
	cd go && go build -o ../.build/cli ./cmd/cli/...

build-e2e: build-dir
	cd go && go build -o ../.build/e2e ./pkg/testing/e2e

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

.PHONY: protos
protos:
	./build_protos.sh

test: test-go

# run the e2e test
e2e: build-go build-e2e
	# Delete any tasks from previous run
	rm -f /tmp/tasks.json
	# This isn't a very robust way to deal with cleaning up logs 
	# Should we do it inside the test? Should we pass a timestamped directory?
	rm -rf /tmp/flaapE2ELogs	
	PYTHONPATH=$(ROOT)/py .build/e2e \
		--taskstore=$(ROOT)/.build/server \
		--json-logs=false --level=debug \
		--port=8081 --terminate=true


# Cleanup e2e processes
cleanup-e2e:
	-pkill -x server
	-pkill -x e2e
	-pkill -f "python3.*flaap.*"