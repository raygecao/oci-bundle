BINDIR ?= $(CURDIR)/bin

GIT_COMMIT    = $(shell git rev-parse HEAD)
GIT_SHA       = $(shell git rev-parse --short HEAD)
GIT_TAG       = $(shell git describe --tags --abbrev=0 --exact-match 2>/dev/null)
GIT_DIRTY     = $(shell test -n "`git status --porcelain`" && echo "dirty" || echo "clean")
VER_PATH      = ocibundle/internal/version

LDFLAGS := -w -s

ifdef VERSION
	BINARY_VERSION = $(VERSION)
endif
BINARY_VERSION ?= ${GIT_TAG}

# Only set Version if building a tag or VERSION is set
ifneq ($(BINARY_VERSION),)
	LDFLAGS += -X ${VER_PATH}.version=${BINARY_VERSION}
endif

VERSION_METADATA =
# Clear the "unreleased" string in BuildMetadata
ifneq ($(GIT_TAG),)
	VERSION_METADATA =
endif

LDFLAGS += -X ${VER_PATH}.metadata=${VERSION_METADATA}
LDFLAGS += -X ${VER_PATH}.gitCommit=${GIT_COMMIT}
LDFLAGS += -X ${VER_PATH}.gitTreeState=${GIT_DIRTY}


all: client

client: $(BINDIR)/cb

prepare:
	go generate ./...
	go mod tidy

clean:
	rm -r $(BINDIR)

.PHONY: all prepare clean

$(BINDIR)/cb: prepare
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)' -tags '$(TAGS)' -o $@ ./cmd

