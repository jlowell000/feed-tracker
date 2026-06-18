BINDIR ?= ./bin
RACE   ?= 1

ifeq ($(RACE), 1)
RACE_FLAG = -race
else
RACE_FLAG =
endif

.PHONY: all build test vet lint tidy clean run-cli run-tui install

all: build vet test

build:
	@mkdir -p $(BINDIR)
	go build -o $(BINDIR)/ ./cmd/...

test:
	go test $(RACE_FLAG) ./... -count=1

vet:
	go vet ./...

lint:
	@which golangci-lint > /dev/null 2>&1 && golangci-lint run || echo "golangci-lint not installed, skipping"

tidy:
	go mod tidy

clean:
	rm -rf $(BINDIR)

run-cli:
	go run ./cmd/cli/...

run-tui:
	go run ./cmd/tui/...

install:
	go install ./cmd/...
