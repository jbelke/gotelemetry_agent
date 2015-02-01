# List building
ALL_LIST = agent.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test -race
GOFMT=gofmt -w

PREBUILD_LIST = $(foreach int, $(ALL_LIST), $(int)_prebuild)
POSTBUILD_LIST = $(foreach int, $(ALL_LIST), $(int)_postbuild)
BUILD_LIST_OSX = $(foreach int, $(ALL_LIST), $(int)_build_osx)
BUILD_LIST_WIN = $(foreach int, $(ALL_LIST), $(int)_build_win)
BUILD_LIST_LINUX = $(foreach int, $(ALL_LIST), $(int)_build_linux)
TEST_LIST = $(foreach int, $(ALL_LIST), $(int)_test)
FMT_TEST = $(foreach int, $(ALL_LIST), $(int)_fmt)
RUN_LIST = $(foreach int, $(ALL_LIST), $(int)_run)

# All are .PHONY for now because dependencyness is hard
.PHONY: $(CLEAN_LIST) $(TEST_LIST) $(FMT_LIST) $(BUILD_LIST)

all: build
build: prebuild $(BUILD_LIST_OSX) $(BUILD_LIST_WIN) $(BUILD_LIST_LINUX) postbuild
build_osx: prebuild $(BUILD_LIST_OSX) postbuild
build_win: prebuild $(BUILD_LIST_WIN) postbuild
build_linux: prebuild $(BUILD_LIST_LINUX) postbuild
clean: $(CLEAN_LIST)
test: $(TEST_LIST)
fmt: $(FMT_TEST)
run: $(RUN_LIST)

prebuild:
	@if [ -f ./prebuild ]; then \
		echo "Running prebuild script in release mode..." ; \
		./prebuild --release ; \
	else  \
		echo "No pre-build script found in pre-build phase; skipping." ; \
	fi

postbuild:
	@if [ -f ./prebuild ]; then \
		echo "Running prebuild script in debug mode..." ; \
		./prebuild --debug ; \
	else  \
		echo "No pre-build script found in post-build phase; skipping." ; \
	fi

$(BUILD_LIST_OSX): %_build_osx: %_fmt
	@echo "Building Darwin AMD64..."
	@GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 $(GOBUILD) -o bin/darwin-amd64/$*
	@echo "Building complete."

$(BUILD_LIST_WIN): %_build_win: %_fmt
	@echo "Building Windows 386..."
	@GOARCH=386 CGO_ENABLED=1 GOOS=windows CC="i686-w64-mingw32-gcc -fno-stack-protector -D_FORTIFY_SOURCE=0 -lssp -D_localtime32=localtime" $(GOBUILD) -o bin/windows-386/$* 
	@echo "Building complete."

$(BUILD_LIST_LINUX): %_build_linux: %_fmt
	@echo "Building Linux AMD64..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=1 CC="gcc" $(GOBUILD) -o bin/linux-amd64/$*
	@echo "Building complete."

$(TEST_LIST): %_test:
	@echo "Running go test..."
	@$(GOTEST) ./...

$(FMT_TEST): %_fmt:
	@echo "Running go fmt..."
	@$(GOFMT) agent.go agent plugin
