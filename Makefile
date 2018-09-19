DIST := dist
EXECUTABLE := gopushserver

GO ?= go
GOFMT ?= gofmt "-s"
DEPLOY_ACCOUNT := nicholasmata
DEPLOY_IMAGE := $(EXECUTABLE)

TARGETS ?= linux
ARCHS ?= amd64

GOFILES := $(shell find . -name "*.go" -type f -not -path "./vendor/*")
SOURCES ?= $(shell find . -name "*.go" -type f)


GOVENDOR := $(GOPATH)/bin/govendor
GOX := $(GOPATH)/bin/gox
MISSPELL := $(GOPATH)/bin/misspell

all: build

$(GOVENDOR):
	$(GO) get -u github.com/kardianos/govendor

$(GOX):
	$(GO) get -u github.com/mitchellh/gox

$(MISSPELL):
	$(GO) get -u github.com/client9/misspell/cmd/misspell

build: $(EXECUTABLE)

build_linux_amd64:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags '$(TAGS)' -ldflags '$(EXTLDFLAGS)-s -w $(LDFLAGS)' -o release/linux/amd64/$(DEPLOY_IMAGE)

$(EXECUTABLE): $(SOURCES)
	$(GO) build -v -tags '$(TAGS)' -ldflags '$(EXTLDFLAGS)-s -w $(LDFLAGS)' -o release/$@

# Release helpers
release: release-dirs release-build release-copy release-check

release-dirs:
	mkdir -p $(DIST)/binaries $(DIST)/release

release-build: $(GOX)
	$(GOX) -os="$(TARGETS)" -arch="$(ARCHS)" -tags="$(TAGS)" -ldflags="$(EXTLDFLAGS)-s -w $(LDFLAGS)" -output="$(DIST)/binaries/$(EXECUTABLE)-$(VERSION)-{{.OS}}-{{.Arch}}"

release-copy:
	$(foreach file,$(wildcard $(DIST)/binaries/$(EXECUTABLE)-*),cp $(file) $(DIST)/release/$(notdir $(file));)

release-check:
	cd $(DIST)/release; $(foreach file,$(wildcard $(DIST)/release/$(EXECUTABLE)-*),gsha256sum $(notdir $(file)) > $(notdir $(file)).sha256;)

# Docker Helpers
docker_image:
	docker build -t $(DEPLOY_ACCOUNT)/$(DEPLOY_IMAGE) -f Dockerfile .

docker_release: docker_image