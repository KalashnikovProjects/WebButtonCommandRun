BINARIES_PATH ?= binaries

# CGO/cross settings
CGO_ENABLED ?= 1
# ARM defaults
GOARM ?= 7

CC_LINUX_AMD64=x86_64-linux-gnu-gcc
CC_LINUX_386=i686-linux-gnu-gcc
CC_LINUX_ARM=arm-linux-gnueabihf-gcc
CC_LINUX_ARM64=aarch64-linux-gnu-gcc
CC_WINDOWS_AMD64=x86_64-w64-mingw32-gcc
CC_WINDOWS_386=i686-w64-mingw32-gcc


.PHONY: all binaries clean build-all build-current build-macos32 build-macos build-macos-arm build-linux32 build-linux build-linux-arm build-linux-arm64 build-windows32 build-windows test-race test-coverage install-lint lint test deps

all: test test-coverage test-race lint build-all

binaries:
ifeq ($(OS),Windows_NT)
	if not exist "${BINARIES_PATH}" mkdir "${BINARIES_PATH}"
else
	mkdir -p ${BINARIES_PATH}
endif

build-all: binaries build-windows build-windows32 build-linux build-linux-arm build-linux-arm64 build-linux32 build-macos build-macos-arm

ifeq ($(OS),Windows_NT)
build-current: binaries
	set CGO_ENABLED=$(CGO_ENABLED)&&  go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr.exe ./cmd

build-windows: binaries
	set GOOS=windows&& set GOARCH=amd64&& set CGO_ENABLED=$(CGO_ENABLED)&& $(if $(CC_WINDOWS_AMD64),set CC=$(CC_WINDOWS_AMD64)&& ,) go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr.exe ./cmd

build-windows32: binaries
	set GOOS=windows&& set GOARCH=386&& set CGO_ENABLED=$(CGO_ENABLED)&& $(if $(CC_WINDOWS_386),set CC=$(CC_WINDOWS_386)&& ,) go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr32.exe ./cmd

build-linux: binaries
	set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=$(CGO_ENABLED)&& $(if $(CC_LINUX_AMD64),set CC=$(CC_LINUX_AMD64)&& ,) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux ./cmd

build-linux-arm: binaries
	set GOOS=linux&& set GOARCH=arm&& set GOARM=$(GOARM)&& set CGO_ENABLED=$(CGO_ENABLED)&& $(if $(CC_LINUX_ARM),set CC=$(CC_LINUX_ARM)&& ,) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux-arm ./cmd

build-linux-arm64: binaries
	set GOOS=linux&& set GOARCH=arm64&& set CGO_ENABLED=$(CGO_ENABLED)&& $(if $(CC_LINUX_ARM64),set CC=$(CC_LINUX_ARM64)&& ,) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux-arm64 ./cmd

build-linux32: binaries
	set GOOS=linux&& set GOARCH=386&& set CGO_ENABLED=$(CGO_ENABLED)&& $(if $(CC_LINUX_386),set CC=$(CC_LINUX_386)&& ,) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux32 ./cmd

build-macos: binaries
	set GOOS=darwin&& set GOARCH=amd64&& set CGO_ENABLED=$(CGO_ENABLED)&& go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos ./cmd

build-macos-arm: binaries
	set GOOS=darwin&& set GOARCH=arm64&& set CGO_ENABLED=$(CGO_ENABLED)&& go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos-arm ./cmd

build-macos32: binaries
	set GOOS=darwin&& set GOARCH=386&& set CGO_ENABLED=1&& go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos-32 ./cmd

run:
	set CGO_ENABLED=$(CGO_ENABLED)&& go run ./cmd

test:
	set CGO_ENABLED=$(CGO_ENABLED)&& go test -v ./...

test-race:
	set CGO_ENABLED=$(CGO_ENABLED)&& go test -v -race ./...

test-coverage:
	set CGO_ENABLED=$(CGO_ENABLED)&& go test -v -coverprofile=coverage.out ./...
else
build-current: binaries
	CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr ./cmd

build-windows: binaries
	GOOS=windows GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(if $(CC_WINDOWS_AMD64),CC=$(CC_WINDOWS_AMD64)) go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr.exe ./cmd

build-windows32: binaries
	GOOS=windows GOARCH=386 CGO_ENABLED=$(CGO_ENABLED) $(if $(CC_WINDOWS_386),CC=$(CC_WINDOWS_386)) go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr32.exe ./cmd

build-linux: binaries
	GOOS=linux GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) $(if $(CC_LINUX_AMD64),CC=$(CC_LINUX_AMD64)) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux ./cmd

build-linux-arm: binaries
	GOOS=linux GOARCH=arm GOARM=$(GOARM) CGO_ENABLED=$(CGO_ENABLED) $(if $(CC_LINUX_ARM),CC=$(CC_LINUX_ARM)) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux-arm ./cmd

build-linux-arm64: binaries
	GOOS=linux GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) $(if $(CC_LINUX_ARM64),CC=$(CC_LINUX_ARM64)) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux-arm64 ./cmd

build-linux32: binaries
	GOOS=linux GOARCH=386 CGO_ENABLED=$(CGO_ENABLED) $(if $(CC_LINUX_386),CC=$(CC_LINUX_386)) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux32 ./cmd

build-macos: binaries
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos ./cmd

build-macos-arm: binaries
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=$(CGO_ENABLED) go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos-arm ./cmd

run:
	CGO_ENABLED=$(CGO_ENABLED) go run ./cmd

test:
	CGO_ENABLED=$(CGO_ENABLED) go test -v ./...

test-race:
	CGO_ENABLED=$(CGO_ENABLED) go test -v -race ./...

test-coverage:
	CGO_ENABLED=$(CGO_ENABLED) go test -v -coverprofile=coverage.out ./...
endif

clean:
ifeq ($(OS),Windows_NT)
	rmdir /s /q ${BINARIES_PATH} 2>nul || exit 0
else
	rm -rf ${BINARIES_PATH}/
endif
	go clean

lint:
	golangci-lint run

install-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
