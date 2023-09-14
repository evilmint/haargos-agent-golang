APP_NAME := haargos
DIST_DIR := dist

# List of OS and architectures for cross-compilation
OS_ARCH := \
    linux/386   linux/amd64 \
    windows/386 windows/amd64 \
    darwin/amd64

VERSION := $(shell cat VERSION)

distribute: $(OS_ARCH)

# Rule to create the distribution directory
$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Rule to build for each OS and architecture
$(OS_ARCH): $(DIST_DIR)
	GOOS=$(firstword $(subst /, ,$@)) GOARCH=$(lastword $(subst /, ,$@)) \
    go build -o $(DIST_DIR)/$(APP_NAME)-$(VERSION)-$(firstword $(subst /, ,$@))-$(lastword $(subst /, ,$@))
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
	@go build -ldflags "-X 'client.API_URL=${API_URL}'" -o haargos-prod
	@echo "Reloading daemons"
	@systemctl daemon-reload
	@echo "Stopping service..."
	@systemctl stop haargos.service
	@cp haargos-prod /usr/local/bin/haargos
	@echo "Starting service..."	
	@systemctl start haargos.service
	@echo "Haargos service replaced"

.PHONY: distribute clean
