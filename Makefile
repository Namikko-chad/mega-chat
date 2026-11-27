.PHONY: all test race vet build build-windows core-lib core-dll build-linux build-apk proto aar app-deps app-test app-run run-demo clean package-linux icons

# Every build output lands in builds/ at the repo root.
BIN := builds
SHA := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
VERSION := 0.1.0
# One stamp per make run: the app and the core built together share it (the
# app's CMake build runs `make core-lib` with PARLEY_BUILD_ID set).
STAMP := $(shell date +%m%d.%H%M%S)
BUILD_ID := $(or $(PARLEY_BUILD_ID),$(SHA)-$(STAMP))
DEFINES := --dart-define=APP_VERSION=$(VERSION) --dart-define=BUILD_SHA=$(SHA) --dart-define=BUILD_ID=$(BUILD_ID) --dart-define=ENV=$(or $(ENV),dev)
# Telegram app credentials (my.telegram.org): make TG_APP_ID=... TG_APP_HASH=...
# Without them the Telegram sign-in asks for them. Likewise the Google OAuth
# client (a "Desktop app" client): make GOOGLE_CLIENT_ID=... GOOGLE_CLIENT_SECRET=...
GO_LDFLAGS = -X github.com/SelfHold/parley/core/rpc.Version=$(VERSION) -X github.com/SelfHold/parley/core/rpc.Build=$(BUILD_ID) \
	-X github.com/SelfHold/parley/core/providers/telegram.AppVersion=$(VERSION) \
	-X github.com/SelfHold/parley/core/providers/telegram.DefaultAppID=$(TG_APP_ID) \
	-X github.com/SelfHold/parley/core/providers/telegram.DefaultAppHash=$(TG_APP_HASH) \
	-X github.com/SelfHold/parley/core/providers/google.DefaultClientID=$(GOOGLE_CLIENT_ID) \
	-X github.com/SelfHold/parley/core/providers/google.DefaultClientSecret=$(GOOGLE_CLIENT_SECRET)
NDK ?= $(lastword $(sort $(wildcard $(HOME)/Android/Sdk/ndk/*)))
PROTOC_DART := $(HOME)/.pub-cache/bin/protoc-gen-dart

# Test runs are capped at 1 GB and run one at a time. Linking a test binary
# with gotd's generated tg package does not fit in 1 GB by default: one
# compiler process, aggressive GC and no DWARF make it fit (vet runs in
# `make vet`).
CAP := systemd-run --user --scope -p MemoryMax=1G -p MemorySwapMax=0 --quiet
GOTEST := $(CAP) env GOMAXPROCS=1 GOGC=10 go test -p 1 -vet=off -ldflags=-w

all: vet test build

# --- Go core -------------------------------------------------------------

test:
	$(GOTEST) ./core/...

# The race detector needs more memory: everything that does not link Telegram (its tests
# are pure mapping code, no goroutines).
race:
	$(CAP) go test -race -p 1 $$(go list ./core/... | grep -v -e /providers/telegram -e /providers/all -e /core/embed)

vet:
	go vet ./...

# The CLI; the core itself is part of the app (core-lib, core-dll, aar).
build:
	go build -ldflags '$(GO_LDFLAGS)' -o $(BIN)/ ./cmd/parley

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags '$(GO_LDFLAGS)' -o $(BIN)/windows/ ./cmd/parley

# The desktop core (desktopbind) as a C archive, linked into the Linux app's
# executable. app/linux/CMakeLists.txt runs this, so `flutter build linux`
# and `flutter run` build it too. Without PARLEY_BUILD_ID the build id is
# "dev", like the app's.
CORE_LIB ?= app/build/core/libparley.a
core-lib: BUILD_ID = $(or $(PARLEY_BUILD_ID),dev)
core-lib:
	CGO_ENABLED=1 go build -buildmode=c-archive -ldflags '$(GO_LDFLAGS)' -o $(CORE_LIB) ./desktopbind

# The Windows core as a DLL next to parley.exe (Go cannot be linked into an
# MSVC executable). app/windows/CMakeLists.txt runs this on Windows (needs gcc,
# e.g. MSYS2 mingw-w64); elsewhere it cross-builds with x86_64-w64-mingw32-gcc.
CORE_DLL ?= $(BIN)/windows/parley_core.dll
core-dll: BUILD_ID = $(or $(PARLEY_BUILD_ID),dev)
core-dll:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(if $(filter Windows_NT,$(OS)),,CC=x86_64-w64-mingw32-gcc) \
		go build -buildmode=c-shared -ldflags '$(GO_LDFLAGS)' -o $(CORE_DLL) ./desktopbind

# Go and Dart code from proto/parley.proto (protoc-gen-go, -go-grpc and
# "dart pub global activate protoc_plugin"), and the Go client of Shard's
# Connector API from proto/connector.proto (a byte-for-byte copy of
# Shard's, see CONNECTOR.md).
proto:
	protoc -I proto \
		--go_out=core/rpc/pb --go_opt=paths=source_relative \
		--go-grpc_out=core/rpc/pb --go-grpc_opt=paths=source_relative \
		proto/parley.proto
	protoc -I proto --plugin=protoc-gen-dart=$(PROTOC_DART) --dart_out=grpc:app/lib/data/rpc/gen proto/parley.proto
	protoc -I proto \
		--go_out=core/shard/pb --go_opt=paths=source_relative --go_opt=Mconnector.proto=github.com/SelfHold/parley/core/shard/pb \
		--go-grpc_out=core/shard/pb --go-grpc_opt=paths=source_relative --go-grpc_opt=Mconnector.proto=github.com/SelfHold/parley/core/shard/pb \
		proto/connector.proto

# The core for Android (gomobile, NDK from ~/Android/Sdk/ndk). -checklinkname=0:
# pion's github.com/wlynxg/anet go:linknames net.zoneCache on Android (see its README).
aar:
	ANDROID_NDK_HOME=$(NDK) gomobile bind -target=android/arm64,android/amd64 -androidapi 26 \
		-javapkg dev.selfhold.parley -ldflags='-s -w -checklinkname=0 $(GO_LDFLAGS)' -o app/android/app/libs/parley.aar ./mobilebind

# The app with the Demo messenger, for working on it without accounts.
run-demo: app-deps
	cd app && PARLEY_DEMO=1 PARLEY_BUILD_ID=$(BUILD_ID) flutter run -d linux $(DEFINES)

# --- Flutter app ---------------------------------------------------------

app-deps:
	cd app && flutter pub get

app-test: app-deps
	cd app && flutter analyze && $(CAP) env NODE_OPTIONS=--max-old-space-size=768 flutter test --concurrency=1

# Desktop app: the core is linked into builds/linux/parley, one process.
# A CMake cache made at another checkout path (the repo was moved or renamed)
# breaks `flutter build linux`, so it is dropped first.
build-linux: app-deps
	@c=app/build/linux/x64/release/CMakeCache.txt; \
	if [ -f "$$c" ] && ! grep -qxF "CMAKE_CACHEFILE_DIR:INTERNAL=$(CURDIR)/app/build/linux/x64/release" "$$c"; then \
		echo "CMake cache from another checkout path: removing app/build/linux"; \
		rm -rf app/build/linux; \
	fi
	cd app && PARLEY_BUILD_ID=$(BUILD_ID) flutter build linux --release $(DEFINES)
	rm -rf $(BIN)/linux && mkdir -p $(BIN)/linux
	cp -r app/build/linux/x64/release/bundle/. $(BIN)/linux/
	@echo "app: $(BIN)/linux/parley"

# Linux packages from the build-linux output, described in packaging/linux/nfpm.yaml.
# PKG picks the formats (deb, rpm, archlinux; all three by default), e.g.
#   make package-linux PKG=archlinux
# Packages land in $(BIN)/packages/. nfpm runs through `go run` unless it is installed.
PKG ?= deb rpm archlinux
NFPM ?= $(or $(shell command -v nfpm 2>/dev/null),go run github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.49.0)

package-linux: build-linux
	mkdir -p $(BIN)/packages
	@for p in $(PKG); do \
		PKG_VERSION=$(VERSION) $(NFPM) package -f packaging/linux/nfpm.yaml -p $$p -t $(BIN)/packages/ || exit 1; \
	done

# Re-renders the committed PNG icons (app, tray, Android) from the SVGs in branding/.
icons:
	branding/render.sh

# APKs only for arm64 (phones) and x86_64 (emulators): the Go core is built for these.
build-apk: aar app-deps
	cd app && flutter build apk --release --split-per-abi --target-platform android-arm64,android-x64 $(DEFINES)
	mkdir -p $(BIN)/android
	cp app/build/app/outputs/flutter-apk/*-release.apk $(BIN)/android/
	@ls -lh $(BIN)/android/*.apk

app-run: app-deps
	cd app && PARLEY_BUILD_ID=$(BUILD_ID) flutter run $(DEFINES)

clean:
	rm -rf $(BIN) app/build app/android/app/libs/parley.aar app/android/app/libs/parley-sources.jar
