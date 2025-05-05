# Binary and image names
APP_NAME := wap-tool
IMAGE_NAME := $(APP_NAME)
TAR_NAME := $(IMAGE_NAME).tar.gz

# Build the Go binary
.PHONY: build
build:
	./sync-schema.sh
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(APP_NAME) .

# Run the binary locally with a test file
.PHONY: run
run: build
	./$(APP_NAME) ../data/det6.yaml

# Build the Docker image
.PHONY: docker
docker: build
	docker build -t $(IMAGE_NAME) .

# Save Docker image to tar.gz for distribution
.PHONY: image-tar
image-tar: docker
	docker save $(IMAGE_NAME) | gzip > $(TAR_NAME)

# Clean up all generated files
.PHONY: clean
clean:
	rm -f $(APP_NAME) $(TAR_NAME)

