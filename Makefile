$(if $(filter 4.%,$(MAKE_VERSION)),,\
	$(error GNU make 4.0 or above is required.))

all:

include 3rdparty.mk

# Use the newest version of the Go compiler, as installed by the
# Debian packages
GOROOT ?= $(firstword $(wildcard $(patsubst %,/usr/lib/go-%,1.11 1.10 1.9 1.8)))
NAMESPACE := github.com/dcso/spyre
GOFILES := $(shell find $(CURDIR) \
		-not -path '$(CURDIR)/vendor/*' \
		-not -path '$(CURDIR)/_*' \
		-type f -name '*.go')
VERSION := $(shell < cmd/spyre/version.go sed -ne '/var version/{ s/.*"\(.*\)"/\1/;p }')

RCFILES := \
	cmd/spyre/spyre_resource_windows_amd64.syso \
	cmd/spyre/spyre_resource_windows_386.syso

ARCHS := \
	x86_64-linux-musl i386-linux-musl \
	x86_64-w64-mingw32 i686-w64-mingw32
EXE := $(foreach arch,$(ARCHS),\
	_build/$(arch)/spyre$(if $(findstring w64-mingw32,$(arch)),.exe))

.PHONY: all
all: $(EXE)
	$(info Build OK)

# Set up target-architecture-specific environment variables:
# CC, PKG_CONFIG_PATH, GOOS, GOARCH
$(foreach arch,$(ARCHS),\
	$(eval _build/$(arch)/%: private export CC=$(arch)-gcc)\
	$(eval _build/$(arch)/%: private export PKG_CONFIG_PATH=$(CURDIR)/_3rdparty/tgt/$(arch)/lib/pkgconfig)\
	$(eval _build/$(arch)/%: private export GOOS=\
		$(or $(if $(findstring linux,$(arch)),linux),\
		     $(if $(findstring mingw,$(arch)),windows),\
		     $(error Could not derive GOOS from $(arch))))\
	$(eval _build/$(arch)/%: private export GOARCH=\
		$(or $(if $(findstring x86_64,$(arch)),amd64),\
		     $(if $(or $(findstring i386,$(arch)),$(findstring i686,$(arch))),386),\
		     $(error Could not derive GOARCH from $(arch)))))

unit-test: private export CC=x86_64-linux-musl-gcc
unit-test: private export PKG_CONFIG_PATH=$(CURDIR)/_3rdparty/tgt/x86_64-linux-musl/lib/pkgconfig
unit-test: private export GOOS=linux
unit-test: private export GOARCH=amd64

$(EXE) unit-test: private export CGO_ENABLED=1
$(EXE) unit-test: private export PATH := $(CURDIR)/_3rdparty/tgt/bin:$(PATH)
$(EXE) vendor/.exists dep-% unit-test: private export GOPATH=$(CURDIR)/_gopath

# Set up vendor directory using github.com/golang/dep
_gopath/.exists: Gopkg.lock Gopkg.toml
	rm -f $(@D)/src/$(NAMESPACE)
	mkdir -p $(dir $(@D)/src/$(NAMESPACE))
	ln -sf $(CURDIR) $(@D)/src/$(NAMESPACE)
	touch $@

# Set up GOPATH via symlink
# (Technically, this does not need to be in the build directory.)
vendor/.exists: _gopath/.exists
	$(info [+] Populating vendor/ directory...)
	mkdir -p vendor
	cd _gopath/src/$(NAMESPACE) && dep ensure -vendor-only -v
	touch $@

.PHONY: dep-ensure
dep-ensure: _gopath/.exists
	cd _gopath/src/$(NAMESPACE) && dep ensure -v
dep-ensure-update:  _gopath/.exists
	cd _gopath/src/$(NAMESPACE) && dep ensure -update -v

# Build resource files
%_resource_windows_amd64.syso: %.rc
	x86_64-w64-mingw32-windres --output-format coff -o $@ -i $<
%_resource_windows_386.syso: %.rc
	i686-w64-mingw32-windres --output-format coff -o $@ -i $<

.PHONY: unit-test
unit-test:
	$(info [+] Running tests...)
	$(GOROOT)/bin/go test -v \
		-ldflags '-w -s -extldflags "-static"' \
		-tags yara_static \
		$(patsubst %,$(NAMESPACE)/%,$(shell find -not -path '*/vendor/*' \
							-not -path '*/_gopath/*' \
							-type f -name '*_test.go' \
							| xargs dirname | xargs -n1 basename))

$(EXE) unit-test: $(GOFILES) $(RCFILES) Makefile 3rdparty.mk 3rdparty-all.stamp _gopath/.exists vendor/.exists

$(EXE): export CGO_CFLAGS += $(shell PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) pkg-config --static --cflags yara)
$(EXE): export CGO_LDFLAGS += $(shell PKG_CONFIG_PATH=$(PKG_CONFIG_PATH) pkg-config --static --libs yara)
$(EXE):
	$(info [+] GOROOT=$(GOROOT) GOOS=$(GOOS) GOARCH=$(GOARCH) CC=$(CC))
	$(info [+] CGO_CFLAGS=$(CGO_CFLAGS))
	$(info [+] CGO_LDFLAGS=$(CGO_LDFLAGS))
	mkdir -p $(@D)
	$(GOROOT)/bin/go build \
		-ldflags '-w -s -linkmode=external -extldflags "-static"' \
		-tags no_pkg_config \
		-o $@ $(NAMESPACE)/cmd/spyre

.PHONY: release
release: spyre-$(VERSION).zip
spyre-$(VERSION).zip: $(EXE)
	$(info [+] Building zipfile ...)
	( cd _build && zip -r $(CURDIR)/$@ . )

.PHONY: clean distclean
clean:
	rm -rf _build $(RCFILES) spyre-$(VERSION).zip
distclean: clean 3rdparty-distclean
	rm -rf _gopath _vendor
