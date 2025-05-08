# Binary and image names
APP_NAME := wap-tool
IMAGE_NAME := $(APP_NAME)
IMAGE_TAR := docker-$(APP_NAME).tar
TAR_NAME := $(APP_NAME).tar.gz

# Build the Go binary
.PHONY: build
build:
	./sync-schema.sh
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(APP_NAME) .

# Run the binary locally with a test file
.PHONY: run
run: build
	./$(APP_NAME) ../data/test.yaml -o /tmp/wap-test.pdf

# Build the Docker image
.PHONY: docker
docker: build
	docker build -t $(IMAGE_NAME) .

# Save Docker image to tar.gz for distribution
.PHONY: dist
dist: build docker
	docker save $(IMAGE_NAME) > $(IMAGE_TAR)
	tar -czf $(APP_NAME).tar.gz $(IMAGE_TAR) ../README.md ../run-docker.sh ../data

.PHONY: lint
lint:
	golangci-lint run ./...

# Clean up all generated files
.PHONY: clean
clean:
	rm -f $(APP_NAME) $(TAR_NAME) $(IMAGE_TAR)

