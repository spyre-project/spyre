# make newline character available as a variable
define \n


endef

# Architectures
# -------------
# Determine host architecture:
3rdparty_NATIVE_ARCH := $(shell gcc -dumpmachine)
# Determine target architectures:
# On Linux, we can cross-build for Linux and Windows.
# On MacOSX, we can only build for MacOSX.
$(or \
	$(if $(findstring -linux-gnu,$(3rdparty_NATIVE_ARCH)),\
		$(eval 3rdparty_ARCHS=i386-linux-musl x86_64-linux-musl i686-w64-mingw32 x86_64-w64-mingw32)\
		$(foreach arch,i686-w64-mingw32 x86_64-w64-mingw32,\
			$(if $(not $(shell which $(arch)-gcc)),$(error $(arch)-gcc not found)))),\
	$(if $(findstring -apple-darwin,$(3rdparty_NATIVE_ARCH)),\
		$(eval 3rdparty_ARCHS=x86_64-apple-darwin)),\
	$(error Unknown native triplet $(3rdparty_NATIVE_ARCH)))

# Package definitions
# -------------------
3rdparty_JOBS    := 8
3rdparty_TARGETS := yara musl openssl

yara_VERSION := 3.10.0
yara_URL     := https://github.com/VirusTotal/yara/archive/v$(yara_VERSION).tar.gz
yara_ARCHS   := $(3rdparty_ARCHS)
# This is executed in the source directory
yara_PREP    := ./bootstrap.sh

musl_VERSION := 1.1.22
musl_URL     := http://www.musl-libc.org/releases/musl-$(musl_VERSION).tar.gz
musl_ARCHS   := $(filter %-linux-musl,$(3rdparty_ARCHS))

openssl_VERSION := 1.1.0k
openssl_URL     := https://www.openssl.org/source/openssl-$(openssl_VERSION).tar.gz
openssl_ARCHS   := $(3rdparty_ARCHS)

# Rules/Templates
# ---------------

# Only set variables when building the dependencies above
_3rdparty/build/%/.build-stamp: \
	export private PATH := $(abspath _3rdparty/tgt/bin):$(PATH)
_3rdparty/build/%/.build-stamp: \
	export private PKG_CONFIG_PATH := $(abspath _3rdparty/tgt/$1/lib/pkgconfig)

# Download tarball for package $1
define download_TEMPLATE
_3rdparty/archive/$1-$($1_VERSION).tar.gz:
	@mkdir -p $$(@D)
	wget -O $$@.t $$($1_URL)
	mv $$@.t $$@
endef


# Unpack tarball for package $1
define unpack_TEMPLATE
_3rdparty/src/$1-$($1_VERSION)/.unpack-stamp: _3rdparty/archive/$1-$($1_VERSION).tar.gz
	@mkdir -p $$(@D)
	tar --strip=1 -xzf $$^ -C $$(@D)
	$(foreach patch,$($1_PATCHES),patch -p1 -d $$(@D) < _3rdparty/$(patch)$(\n))
	$(if $($1_PREP),cd $$(@D) && $($1_PREP))
	touch $$@
endef


# Out-of-tree build, installation for musl, architecture $1
# Notes:
# - syslibdir contains the dynamic linker, this has to be a custom
#   location so we can use any resulting dynamically linked binaries
#   later (yarac).
# - the .spec file needs to be patched so *-musl-gcc will use
#   -m32/-m64 as appropriate.
define build_musl_TEMPLATE
_3rdparty/build/$1/musl-$(musl_VERSION)/.build-stamp: _3rdparty/src/musl-$(musl_VERSION)/.unpack-stamp
	@mkdir -p $$(@D)
	cd $$(@D) && $$(abspath $$(<D))/configure \
		--host=$1 \
		--prefix=$(abspath _3rdparty/tgt/$1) \
		--syslibdir=$(abspath _3rdparty/tgt/$1/lib) \
		CC=gcc \
		CROSS_COMPILE= \
		CFLAGS=$(if $(findstring x86_64,$1),-m64,-m32)
	$(MAKE) -j$(3rdparty_JOBS) -C $$(@D) AR=ar RANLIB=ranlib
	$(MAKE) -C $$(@D) install
	$(abspath _3rdparty)/patch-musl-spec.sh $(abspath _3rdparty/tgt/$1)
	# Make gcc wrapper available as <triplet>-gcc
	@mkdir -p _3rdparty/tgt/bin
	ln -sf $(abspath _3rdparty/tgt/$1)/bin/musl-gcc _3rdparty/tgt/bin/$1-gcc
	touch $$@
endef

# Out-of-tree build for yara, architecture $1, with dependency on musl
# where appropriate
define build_yara_TEMPLATE
$(if $(findstring linux-musl,$1),\
_3rdparty/build/$1/yara-$(yara_VERSION)/.build-stamp: _3rdparty/build/$1/musl-$(musl_VERSION)/.build-stamp\
)

_3rdparty/build/$1/yara-$(yara_VERSION)/.build-stamp: _3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp

_3rdparty/build/$1/yara-$(yara_VERSION)/.build-stamp: _3rdparty/src/yara-$(yara_VERSION)/.unpack-stamp
	@mkdir -p $$(@D)
	cd $$(@D) && $$(abspath $$(<D))/configure \
		--host=$1 \
		--prefix=$(abspath _3rdparty/tgt/$1) \
		--disable-shared \
		--disable-magic --disable-cuckoo --enable-dotnet \
		CC=$$(or $$(shell PATH=$$(PATH) which $1-gcc),$$(shell PATH=$$(PATH) which gcc)) \
		CPPFLAGS="-I$(abspath _3rdparty/tgt/$1/include)" \
		CFLAGS="$(if $(findstring -linux-musl,$1),-static)" \
		LDFLAGS="-L$(abspath _3rdparty/tgt/$1/lib) $$(shell \
			pkg-config --static --libs libcrypto \
			| sed -e 's/-ldl//g' )"
	$(MAKE) -j$(3rdparty_JOBS) -C $$(@D)/libyara
	$(MAKE) -C $$(@D)/libyara install
	$(if $(findstring $(patsubst %-linux-gnu,%-linux-musl,$(3rdparty_NATIVE_ARCH)),$1),\
		ln -sf $(patsubst %,$(abspath _3rdparty/tgt/$1)/bin/%,yarac yara) _3rdparty/tgt//bin)
	touch $$@
endef

define build_openssl_TEMPLATE
$(if $(findstring linux-musl,$1),\
_3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp: _3rdparty/build/$1/musl-$(musl_VERSION)/.build-stamp\
)

_3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp: \
	private export CC=$$(or $$(shell PATH=$$(PATH) which $1-gcc),$$(shell PATH=$$(PATH) which gcc))
_3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp: \
	private export CFLAGS=$(if $(findstring -linux-musl,$1),-static) $(if $(findstring x86_64,$1),-m64,-m32)
_3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp: \
	private export MACHINE=$(if $(findstring x86_64,$1),x86_64,i386)
_3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp: \
	private export SYSTEM=$(or \
		$(if $(findstring mingw,$1),$(if $(findstring x86_64,$1),MINGW64,MINGW32)),\
		$(if $(findstring linux,$1),linux2),\
		$(if $(findstring darwin,$1),Darwin),\
		$(error what should we set MACHINE for the OpenSSL build to?))
_3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp: _3rdparty/src/openssl-$(openssl_VERSION)/.unpack-stamp
	@mkdir -p $$(@D)
	cd $$(@D) && $$(abspath $$(<D))/config \
		no-afalgeng \
		no-async \
		no-capieng \
		no-dso \
		no-shared \
		no-sock \
		no-ui \
		--prefix=$(abspath _3rdparty/tgt/$1)
	$(MAKE) -j$(3rdparty_JOBS) -C $$(@D)
	$(MAKE) -C $$(@D) install_sw
	touch $$@

endef

# Template expansion
$(foreach pkg,$(3rdparty_TARGETS),$(eval $(call download_TEMPLATE,$(pkg))))
$(foreach pkg,$(3rdparty_TARGETS),$(eval $(call unpack_TEMPLATE,$(pkg))))
$(foreach pkg,$(3rdparty_TARGETS),\
	$(foreach arch,$($(pkg)_ARCHS),\
		$(eval $(call build_$(pkg)_TEMPLATE,$(arch)))))

# Targets
# -------

# Build everything
3rdparty-all: 3rdparty-all.stamp
3rdparty-all.stamp: $(foreach pkg,$(3rdparty_TARGETS),\
	$(foreach arch,$($(pkg)_ARCHS),\
		_3rdparty/build/$(arch)/$(pkg)-$($(pkg)_VERSION)/.build-stamp))
	touch $@
# Clean unpacked sources, build, install directory
3rdparty-clean:
	rm -rf 3rdparty-all.stamp _3rdparty/build _3rdparty/src _3rdparty/tgt
# Clean everything
3rdparty-distclean: 3rdparty-clean
	rm -rf _3rdparty/archive

# Debug help
3rdparty-dump-templates:
	$(foreach pkg,$(3rdparty_TARGETS),$(info $(call download_TEMPLATE,$(pkg))))
	$(foreach pkg,$(3rdparty_TARGETS),$(info $(call unpack_TEMPLATE,$(pkg))))
	$(foreach pkg,$(3rdparty_TARGETS),\
		$(foreach arch,$($(pkg)_ARCHS),\
			$(info $(call build_$(pkg)_TEMPLATE,$(arch)))))

.PHONY: 3rdparty-all 3rdparty-clean 3rdparty-distclean 3rdparty-dump-templates
