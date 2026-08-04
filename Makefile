.PHONY: build run clean

# Build the trail binary
build:
	go build -o trail .

# Run with last week analysis
run: build
	./trail run --last-week

# Clean build artifacts
clean:
	rm -f trail

# Run tests
test:
	go test ./...

# Show help
help:
	@echo "Usage:"
	@echo "  make build     - Build the trail binary"
	@echo "  make run       - Build and run for last week"
	@echo "  make clean     - Remove build artifacts"
	@echo "  make test      - Run tests"
	@echo ""
	@echo "Examples:"
	@echo "  ./trail auth"
	@echo "  ./trail run --last-week"
	@echo "  ./trail run --start 2024-01-15 --end 2024-01-21"
