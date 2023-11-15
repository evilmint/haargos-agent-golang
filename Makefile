APP_NAME := haargos
DIST_DIR := dist

# List of OS and architectures for cross-compilation
OS_ARCH := \
    linux/386   linux/amd64

DOCKER_HUB_REPO := thaxsillion/haargos-addon
VERSION := $(shell cat VERSION)

DOCKER_IMAGE := haargos-build-image

distribute: docker-build $(OS_ARCH)

docker-push:
	# Tagging the image for Docker Hub
	docker tag $(APP_NAME)-amd64 $(DOCKER_HUB_REPO):$(VERSION)-amd64
	docker tag $(APP_NAME)-i386 $(DOCKER_HUB_REPO):$(VERSION)-i386

	# Pushing the images to Docker Hub
	docker push $(DOCKER_HUB_REPO):$(VERSION)-amd64
	docker push $(DOCKER_HUB_REPO):$(VERSION)-i386

docker-build:
	# Building the Docker images
	docker build --progress=plain -t $(APP_NAME)-amd64 -f Dockerfile.amd64 .
	docker build --progress=plain -t $(APP_NAME)-i386 -f Dockerfile.386 .

	# Create a container from the image
	docker create --name temp-container-amd64 "$(APP_NAME)-amd64"
	docker create --name temp-container-i386 "$(APP_NAME)-i386"

	# Copy and zip the compiled applications from the container to $(DIST_DIR)
	docker cp temp-container-amd64:/root/app-amd64 $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64
	docker cp temp-container-i386:/root/app-386 $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-386

	zip $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64.zip $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-amd64
	zip $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-386.zip $(DIST_DIR)/$(APP_NAME)-$(VERSION)-linux-386

	# Remove the temporary container
	docker rm temp-container-amd64
	docker rm temp-container-i386

# Rule to create the distribution directory
$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Rule to build for each OS and architecture
$(OS_ARCH): $(DIST_DIR)
	zip $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$(firstword $(subst /, ,$@))-$(lastword $(subst /, ,$@)).zip \
	    $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$(firstword $(subst /, ,$@))-$(lastword $(subst /, ,$@))

# Rule to clean up the distribution directory
clean:
	rm -rf $(DIST_DIR)

dev:
	go build -ldflags "-X 'client.API_URL=${API_URL}'" -o haargos-dev
	DEBUG=true ./haargos-dev run --ha-config /Volumes/haconfig/ha-config/

install:
	@echo "Building Haargos"
	@go build -o haargos-prod
	@echo "Reloading daemons"
	@systemctl daemon-reload
	@echo "Stopping service..."
	@systemctl stop haargos.service
	@cp haargos-prod /usr/local/bin/haargos
	@echo "Starting service..."	
	@systemctl start haargos.service
	@echo "Haargos service replaced"

.PHONY: distribute clean docker-build
