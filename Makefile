APP_NAME := haargos
DIST_DIR := dist

# List of OS and architectures for cross-compilation
OS_ARCH := \
    linux/386   linux/amd64 \
    windows/386 windows/amd64 \
    darwin/amd64

distribute: $(OS_ARCH)

# Rule to create the distribution directory
$(DIST_DIR):
	mkdir -p $(DIST_DIR)

# Rule to build for each OS and architecture
$(OS_ARCH): $(DIST_DIR)
	GOOS=$(firstword $(subst /, ,$@)) GOARCH=$(lastword $(subst /, ,$@)) \
    go build -o $(DIST_DIR)/$(APP_NAME)-$(firstword $(subst /, ,$@))-$(lastword $(subst /, ,$@))

# Rule to clean up the distribution directory
clean:
	rm -rf $(DIST_DIR)

.PHONY: distribute clean
