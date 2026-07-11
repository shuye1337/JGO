APP_NAME    := jgo
PKG         := jgo
VERSION     := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")

ifeq ($(OS),Windows_NT)
    BUILD_DATE := $(shell powershell -Command "Get-Date -Format 'yyyy-MM-ddTHH:mm:sszzz'" 2>/dev/null || echo "unknown")
else
    BUILD_DATE := $(shell date +"%Y-%m-%dT%H:%M:%S%z")
endif
LDFLAGS     := -s -w -X '$(PKG)/cmd.Version=$(VERSION)' -X '$(PKG)/cmd.Commit=$(COMMIT)' -X '$(PKG)/cmd.BuildDate=$(BUILD_DATE)'

GO          := go
BIN_DIR     := bin

PLATFORMS := \
    windows-amd64 \
    windows-arm64 \
    linux-amd64 \
    linux-arm64 \
    darwin-amd64 \
    darwin-arm64

ifeq ($(OS),Windows_NT)
    EXE_EXT := .exe
    RM       := del /Q
    RMDIR    := rmdir /S /Q
else
    EXE_EXT :=
    RM       := rm -f
    RMDIR    := rm -rf
endif

ifeq ($(OS),Windows_NT)
    go-build = set GOOS=$(1)&& set GOARCH=$(2)&& $(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BIN_DIR)/$(APP_NAME)$(if $(filter windows,$(1)),.exe,) .
else
    go-build = GOOS=$(1) GOARCH=$(2) $(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BIN_DIR)/$(APP_NAME)$(if $(filter windows,$(1)),.exe,) .
endif

go-pack = cd $(BIN_DIR) && $(if $(filter windows,$(1)), \
	powershell -Command "Compress-Archive -Path '$(APP_NAME).exe' -DestinationPath '$(APP_NAME)-$(1)-$(2).zip' -Force", \
	tar -czf $(APP_NAME)-$(1)-$(2).tar.gz $(APP_NAME) \
)

.PHONY: all build build-all clean test lint tidy run help

all: build

build:
	$(GO) build -ldflags="$(LDFLAGS)" -trimpath -o $(BIN_DIR)/$(APP_NAME)$(EXE_EXT) .

build-all: $(addprefix build-,$(PLATFORMS))

build-windows-amd64:
	$(call go-build,windows,amd64) && $(call go-pack,windows,amd64)

build-windows-arm64:
	$(call go-build,windows,arm64) && $(call go-pack,windows,arm64)

build-linux-amd64:
	$(call go-build,linux,amd64) && $(call go-pack,linux,amd64)

build-linux-arm64:
	$(call go-build,linux,arm64) && $(call go-pack,linux,arm64)

build-darwin-amd64:
	$(call go-build,darwin,amd64) && $(call go-pack,darwin,amd64)

build-darwin-arm64:
	$(call go-build,darwin,arm64) && $(call go-pack,darwin,arm64)

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

run: build
	./$(BIN_DIR)/$(APP_NAME)$(EXE_EXT)

clean:
	-$(RM) $(BIN_DIR)/*$(EXE_EXT) 2>/dev/null
	-$(RM) $(BIN_DIR)/*.zip 2>/dev/null
	-$(RM) $(BIN_DIR)/*.tar.gz 2>/dev/null
	-$(RMDIR) $(BIN_DIR) 2>/dev/null

help:
	@echo Available targets:
	@echo   build         - Build for current platform
	@echo   build-all     - Cross-compile for all supported platforms
	@echo   test          - Run tests
	@echo   lint          - Run go vet
	@echo   tidy          - Run go mod tidy
	@echo   run           - Build and run
	@echo   clean         - Remove build artifacts
