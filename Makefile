.PHONY: fmt vet test tidy build clean help

help:
	@echo "Targets:"
	@echo "  fmt    — gofmt -w ."
	@echo "  vet    — go vet ./..."
	@echo "  test   — go test ./..."
	@echo "  tidy   — go mod tidy"
	@echo "  build  — go build ./..."
	@echo "  clean  — remove built binaries"
	@echo ""
	@echo "Run an example:"
	@echo "  make run SECTION=01-cli-tools EX=01-flag-basics"

fmt:
	gofmt -w .

vet:
	go vet ./...

test:
	go test ./...

tidy:
	go mod tidy

build:
	go build ./...

clean:
	find . -type f -perm -u+x -not -path "*/.git/*" -not -name "*.sh" -delete 2>/dev/null || true
	rm -rf bin/

run:
	@if [ -z "$(SECTION)" ] || [ -z "$(EX)" ]; then \
		echo "Usage: make run SECTION=<dir> EX=<subdir>"; \
		exit 1; \
	fi
	go run ./$(SECTION)/$(EX)
