BINARIES_PATH ?= binaries

.PHONY: all clean build-all

all: build-all

binaries:
ifeq ($(OS),Windows_NT)
	if not exist "${BINARIES_PATH}" mkdir "${BINARIES_PATH}"
else
	mkdir -p ${BINARIES_PATH}
endif

build-all: binaries build-windows build-windows32 build-linux build-linux-arm build-linux32 build-macos build-macos-arm

ifeq ($(OS),Windows_NT)
build-current: binaries
	set CGO_ENABLED=0&&  go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr.exe ./cmd

build-windows: binaries
	set GOOS=windows&& set GOARCH=amd64&& set CGO_ENABLED=0&&  go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr.exe ./cmd

build-windows32: binaries
	set GOOS=windows&& set GOARCH=386&& set CGO_ENABLED=0&&  go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr32.exe ./cmd

build-linux: binaries
	set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux ./cmd

build-linux-arm: binaries
	set GOOS=linux&& set GOARCH=arm&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux-arm ./cmd

build-linux32: binaries
	set GOOS=linux&& set GOARCH=386&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux32 ./cmd

build-macos: binaries
	set GOOS=darwin&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos ./cmd

build-macos-arm: binaries
	set GOOS=darwin&& set GOARCH=arm64&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos-arm ./cmd

build-macos32: binaries
	set GOOS=darwin&& set GOARCH=386&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos-32 ./cmd
else
build-current: binaries
	CGO_ENABLED=0 go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr ./cmd

build-windows: binaries
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr.exe ./cmd

build-windows32: binaries
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w -extldflags \"-static\"" -o ${BINARIES_PATH}/wbcr32.exe ./cmd

build-linux: binaries
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux ./cmd

build-linux-arm: binaries
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux-arm ./cmd

build-linux32: binaries
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-linux32 ./cmd

build-macos: binaries
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos ./cmd

build-macos-arm: binaries
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o ${BINARIES_PATH}/wbcr-macos-arm ./cmd

endif

clean:
ifeq ($(OS),Windows_NT)
	rmdir /s /q ${BINARIES_PATH} 2>nul || exit 0
else
	rm -rf ${BINARIES_PATH}/
endif
	go clean

deps:
	go mod download
	go mod tidy

test:
	go test -v ./...

test-race:
	go test -v -race ./...

test-coverage:
	go test -v -coverprofile=coverage.out ./...

lint:
	golangci-lint run

ci-build: build-all

install-lint: deps
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest