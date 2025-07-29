.PHONY: all clean build-all build-current build-windows build-windows32 build-linux build-linux-arm build-linux32 build-macos build-macos-arm build-macos32

all: build-all

compiled:
	mkdir compiled

build-all: compiled build-windows build-windows32 build-linux build-linux-arm build-linux32 build-macos build-macos-arm build-macos32

ifeq ($(OS),Windows_NT)
build-current: compiled
	set CGO_ENABLED=0&&  go build -ldflags="-s -w -extldflags \"-static\"" -o compiled/wbcm.exe ./cmd

build-windows: compiled
	set GOOS=windows&& set GOARCH=amd64&& set CGO_ENABLED=0&&  go build -ldflags="-s -w -extldflags \"-static\"" -o compiled/wbcm.exe ./cmd

build-windows32: compiled
	set GOOS=windows&& set GOARCH=386&& set CGO_ENABLED=0&&  go build -ldflags="-s -w -extldflags \"-static\"" -o compiled/wbcm32.exe ./cmd

build-linux: compiled
	set GOOS=linux&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o compiled/wbcm-linux ./cmd

build-linux-arm: compiled
	set GOOS=linux&& set GOARCH=arm&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o compiled/wbcm-linux-arm ./cmd

build-linux32: compiled
	set GOOS=linux&& set GOARCH=386&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o compiled/wbcm-linux32 ./cmd

build-macos: compiled
	set GOOS=darwin&& set GOARCH=amd64&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o compiled/wbcm-macos ./cmd

build-macos-arm: compiled
	set GOOS=darwin&& set GOARCH=arm64&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o compiled/wbcm-macos-arm ./cmd

build-macos32: compiled
	set GOOS=darwin&& set GOARCH=386&& set CGO_ENABLED=0&& go build -ldflags="-s -w" -o compiled/wbcm-macos-32 ./cmd
else
build-current: compiled
	CGO_ENABLED=0 go build -ldflags="-s -w -extldflags \"-static\"" -o compiled/wbcm ./cmd

build-windows: compiled
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w -extldflags \"-static\"" -o compiled/wbcm.exe ./cmd

build-windows32: compiled
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w -extldflags \"-static\"" -o compiled/wbcm32.exe ./cmd

build-linux: compiled
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o compiled/wbcm-linux ./cmd

build-linux-arm: compiled
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -ldflags="-s -w" -o compiled/wbcm-linux-arm ./cmd

build-linux32: compiled
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o compiled/wbcm-linux32 ./cmd

build-macos: compiled
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o compiled/wbcm-macos ./cmd

build-macos-arm: compiled
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o compiled/wbcm-macos-arm ./cmd

build-macos32: compiled
	GOOS=darwin GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o compiled/wbcm-macos-32 ./cmd
endif

clean:
ifeq ($(OS),Windows_NT)
	rmdir /s /q compiled 2>nul || exit 0
else
	rm -rf compiled/
endif
	go clean

deps:
	go mod download
	go mod tidy