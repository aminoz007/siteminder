
BINARY_NAME := $(shell basename $(shell pwd))
VERSION := $$(cat $(BINARY_NAME).go | grep agentVersion | head -1 | cut -d'"' -f 2)
PKG_NAME_LINUX := $(BINARY_NAME)_linux-v$(VERSION)
PKG_NAME_WINDOWS := $(BINARY_NAME)_windows-v$(VERSION)
PKG_NAME_DARWIN := $(BINARY_NAME)_darwin-v$(VERSION)
ARCH ?= $$(uname -s | tr A-Z a-z)
LINE_BREAK := "--------------------------------------------------------------------"

setup: 
	@echo "### This may take some time the first time..."
	@echo ${LINE_BREAK}
	dep ensure
	@echo ${LINE_BREAK}

build-all: build-linux build-darwin build-windows

build: setup
	@echo "### Building for Current OS"
	@echo ${LINE_BREAK}
	GOOS=$(ARCH) go build -o bin/$(ARCH)/$(BINARY_NAME)
	@echo ${LINE_BREAK}

build-linux: setup
	@echo "### Building Linux binary"
	@echo ${LINE_BREAK}
	GOOS=linux go build -o bin/linux/$(BINARY_NAME)
	@echo ${LINE_BREAK}

build-darwin: setup
	@echo "### Building Darwin binary"
	@echo ${LINE_BREAK}
	GOOS=darwin go build -o bin/darwin/$(BINARY_NAME)
	@echo ${LINE_BREAK}

build-windows: setup
	@echo "### Building Windows binary"
	@echo ${LINE_BREAK}
	GOOS=windows go build -o bin/windows/$(BINARY_NAME)
	@echo ${LINE_BREAK}

package-all: package-linux package-darwin package-windows

package-linux: build-linux
	@echo "### Packaging into $(PKG_NAME_LINUX).tar"
	@echo ${LINE_BREAK}
	@rm -rf $(BINARY_NAME)_linux-*
	@mkdir $(PKG_NAME_LINUX)
	@cp ./bin/linux/$(BINARY_NAME) ./$(PKG_NAME_LINUX)/
	@cp ./siteminder.yml ./$(PKG_NAME_LINUX)/
	@cp ./README.md ./$(PKG_NAME_LINUX)/
	@tar -cvf $(PKG_NAME_LINUX).tar $(PKG_NAME_LINUX)/
	@rm -rf $(PKG_NAME_LINUX)
	@echo "Completed packaging: $(PKG_NAME_LINUX).tar"
	@echo ${LINE_BREAK}

package-darwin: build-darwin
	@echo "### Packaging into $(PKG_NAME_DARWIN).tar"
	@echo ${LINE_BREAK}
	@rm -rf $(BINARY_NAME)_darwin-*
	@mkdir $(PKG_NAME_DARWIN)
	@cp ./bin/darwin/$(BINARY_NAME) ./$(PKG_NAME_DARWIN)/
	@cp ./siteminder.yml ./$(PKG_NAME_DARWIN)/
	@cp ./README.md ./$(PKG_NAME_DARWIN)/
	@tar -cvf $(PKG_NAME_DARWIN).tar $(PKG_NAME_DARWIN)/
	@rm -rf $(PKG_NAME_DARWIN)
	@echo "Completed packaging: $(PKG_NAME_DARWIN).tar"
	@echo ${LINE_BREAK}

package-windows: build-windows
	@echo "### Packaging into $(PKG_NAME_WINDOWS).tar"
	@echo ${LINE_BREAK}
	@rm -rf $(BINARY_NAME)_windows-*
	@mkdir $(PKG_NAME_WINDOWS)
	@cp ./bin/windows/$(BINARY_NAME) ./$(PKG_NAME_WINDOWS)/
	@cp ./siteminder.yml ./$(PKG_NAME_WINDOWS)/
	@cp ./README.md ./$(PKG_NAME_WINDOWS)/
	@tar -cvf $(PKG_NAME_WINDOWS).tar $(PKG_NAME_WINDOWS)/
	@rm -rf $(PKG_NAME_WINDOWS)
	@echo "Completed packaging: $(PKG_NAME_WINDOWS).tar"
	@echo ${LINE_BREAK}

version:
	@echo $(VERSION)

clean:
	@echo "### Removing folders: vendor, bin, coverage"
	@echo ${LINE_BREAK}
	rm -rf vendor bin coverage.out
	@echo ${LINE_BREAK}

test: 
	@echo "### Testing locally"
	@echo ${LINE_BREAK}
	go test -cover -coverprofile=coverage.out

run: build
	@echo "### Running agent..."
	@echo "### ctrl+c to exit"
	@echo ${LINE_BREAK}
	go run *.go
	@echo ${LINE_BREAK}

.PHONY : setup build build-linux