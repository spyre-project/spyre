# make newline character available as a variable
define \n


endef

# Architectures
# -------------
# Determine host architecture:
3rdparty_NATIVE_ARCH := $(shell cc -dumpmachine)
# Determine target architectures:
# On Linux, we can cross-build for Windows.
# On other systems, we only build for the respective host architecture
$(or \
	$(if $(or $(findstring x86_64-linux-gnu,$(3rdparty_NATIVE_ARCH)),\
		  $(findstring x86_64-redhat-linux,$(3rdparty_NATIVE_ARCH))),\
		$(eval 3rdparty_ARCHS=\
				i386-linux-musl x86_64-linux-musl \
				i686-w64-mingw32 x86_64-w64-mingw32)\
		$(foreach arch,i686-w64-mingw32 x86_64-w64-mingw32,\
			$(if $(not $(shell which $(arch)-gcc)),$(error $(arch)-gcc not found)))),\
	$(if $(findstring -linux-gnu,$(3rdparty_NATIVE_ARCH)),\
		$(eval 3rdparty_ARCHS=\
				$(patsubst %-linux-gnu,%-linux-musl,$(3rdparty_NATIVE_ARCH)) \
				i686-w64-mingw32 x86_64-w64-mingw32)\
		$(foreach arch,i686-w64-mingw32 x86_64-w64-mingw32,\
			$(if $(not $(shell which $(arch)-gcc)),$(error $(arch)-gcc not found)))),\
	$(if $(or $(findstring -linux-gnu,$(3rdparty_NATIVE_ARCH))),\
		$(eval 3rdparty_ARCHS=$(patsubst %-linux-gnu,%-linux-musl,$(3rdparty_NATIVE_ARCH)))),\
	$(if $(or $(findstring -apple-darwin,$(3rdparty_NATIVE_ARCH)),\
		  $(findstring -freebsd,$(3rdparty_NATIVE_ARCH))),\
		$(eval 3rdparty_ARCHS=$(3rdparty_NATIVE_ARCH))),\
	$(error (Currently) unsupported native triplet $(3rdparty_NATIVE_ARCH)))

# DEPENDS(pkg,dependency,[architectures])
# Declare dependency so that dependency has been built/installed
# before pkg is built. Limit to architectures if set.
define DEPENDS
$(foreach arch,$($1_ARCHS),\
	$(if $(or $(if $3,,x),\
		  $(findstring $(arch),$3)),\
_3rdparty/build/$(arch)/$1-$($1_VERSION)/.build-stamp: _3rdparty/build/$(arch)/$2-$($2_VERSION)/.build-stamp\
)
)
endef

# Package definitions
# -------------------
3rdparty_JOBS    := 8
3rdparty_TARGETS := yara musl openssl

yara_VERSION := 4.5.4
yara_URL     := https://github.com/VirusTotal/yara/archive/v$(yara_VERSION).tar.gz
yara_ARCHS   := $(3rdparty_ARCHS)
# This is executed in the source directory
yara_PREP    := ./bootstrap.sh
yara_PATCHES := yara-winxp-compat.patch

musl_VERSION := 1.2.5
musl_URL     := https://musl.libc.org/releases/musl-$(musl_VERSION).tar.gz
musl_ARCHS   := $(filter %-linux-musl,$(3rdparty_ARCHS))
musl_PATCHES := getauxval.patch

openssl_VERSION := 3.5.0
openssl_URL     := https://www.openssl.org/source/openssl-$(openssl_VERSION).tar.gz
openssl_ARCHS   := $(3rdparty_ARCHS)

# Declare dependencies
$(eval $(call DEPENDS,yara,openssl,))
$(eval $(call DEPENDS,yara,musl,i386-linux-musl x86_64-linux-musl aarch64-linux-musl))
$(eval $(call DEPENDS,openssl,musl,i386-linux-musl x86_64-linux-musl aarch64-linux-musl))

# Rules/Templates
# ---------------

# Only set variables when building the dependencies above
_3rdparty/build/%/.build-stamp: \
	export PATH := $(abspath _3rdparty/tgt/bin):$(PATH)
_3rdparty/build/%/.build-stamp: \
	export PKG_CONFIG_PATH := $(abspath _3rdparty/tgt/$1/lib/pkgconfig)

# Download tarball for package $1
define download_TEMPLATE
_3rdparty/archive/$1-$($1_VERSION).tar.gz:
	@mkdir -p $$(@D)
	wget -q -O $$@.t $$($1_URL)
	mv $$@.t $$@
endef


# Unpack tarball for package $1
define unpack_TEMPLATE
_3rdparty/src/$1-$($1_VERSION)/.unpack-stamp: _3rdparty/archive/$1-$($1_VERSION).tar.gz
	@mkdir -p $$(@D)
	$(TAR) --strip=1 -xzf $$^ -C $$(@D)
	$(foreach patch,$($1_PATCHES),patch -p1 -d $$(@D) < _3rdparty/$(patch); )
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
		CFLAGS="-fPIC $(or $(if $(findstring x86_64,$1),-m64),\
                                   $(if $(or $(findstring i386,$1),$(findstring i686,$1)),-m32))"
	$(MAKE) -s -j$(3rdparty_JOBS) -C $$(@D) AR=ar RANLIB=ranlib
	$(MAKE) -s -C $$(@D) install
	$(abspath _3rdparty)/patch-musl-spec.sh $(abspath _3rdparty/tgt/$1)
	# Make gcc wrapper available as <triplet>-gcc
	@mkdir -p _3rdparty/tgt/bin
	ln -sf $(abspath _3rdparty/tgt/$1)/bin/musl-gcc _3rdparty/tgt/bin/$1-gcc
	$(foreach tool,ar ranlib ld, ln -sf $(shell which $(tool)) _3rdparty/tgt/bin/$1-$(tool); )
	touch $$@
endef

# Out-of-tree build for yara, architecture $1, with dependency on musl
# where appropriate
define build_yara_TEMPLATE
_3rdparty/build/$1/yara-$(yara_VERSION)/.build-stamp: export PKG_CONFIG_PATH=$(abspath _3rdparty/tgt/$1/lib/pkgconfig)
_3rdparty/build/$1/yara-$(yara_VERSION)/.build-stamp: _3rdparty/src/yara-$(yara_VERSION)/.unpack-stamp
	@mkdir -p $$(@D)
	cd $$(@D) && $$(abspath $$(<D))/configure \
		--quiet \
		--host=$1 \
		--prefix=$(abspath _3rdparty/tgt/$1) \
		--disable-shared --with-crypto \
		--disable-magic --disable-cuckoo --enable-macho --enable-dex \
		CC=$$(firstword $$(shell PATH=$$(PATH) which $1-gcc gcc cc)) \
		CPPFLAGS="-I$(abspath _3rdparty/tgt/$1/include)" \
		CFLAGS="$(if $(findstring -linux-musl,$1),-static)" \
		LDFLAGS="-L$(abspath _3rdparty/tgt/$1/lib)"
	$(if $(findstring -w64-mingw32,$1),\
		$(SED) -i 's/-DHAVE__MKGMTIME=1//g' _3rdparty/build/$1/yara-$(yara_VERSION)/Makefile)
	$(MAKE) -s -C $$(@D) uninstall
	$(MAKE) -s -j$(3rdparty_JOBS) -C $$(@D)
	$(MAKE) -s -C $$(@D) install
	$(if $(or $(findstring $(patsubst %-linux-gnu,%-linux-musl,$(3rdparty_NATIVE_ARCH)),$1),
		  $(findstring $(patsubst %-redhat-linux,%-linux-musl,$(3rdparty_NATIVE_ARCH)),$1)),\
		mkdir -p _3rdparty/tgt/bin && ln -sf $(patsubst %,$(abspath _3rdparty/tgt/$1)/bin/%,yarac yara) _3rdparty/tgt//bin)
	$(SED) -i -e '/Libs.private:/ s/ *$$$$/ -lm/' _3rdparty/tgt/$1/lib/pkgconfig/yara.pc
	touch $$@
endef

define build_openssl_TEMPLATE
_3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp: \
	export CC=$(or \
		$(if $(shell which gcc),gcc),\
		$(if $(shell which cc),cc),\
		$(error 3rdparty/openssl: gcc or cc not found))
_3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp: \
	export CFLAGS=$(if $(findstring -linux-musl,$1),-static) $(or $(if $(findstring x86_64,$1),-m64),\
                                                                              $(if $(or $(findstring i386,$1),$(findstring i686,$1)),-m32))
_3rdparty/build/$1/openssl-$(openssl_VERSION)/.build-stamp: _3rdparty/src/openssl-$(openssl_VERSION)/.unpack-stamp
	@mkdir -p $$(@D)
	cd $$(@D) && $$(abspath $$(<D))/config \
		no-afalgeng \
		no-atexit \
		no-autoalginit \
		no-autoerrinit \
		no-apps \
		no-async \
		no-capieng \
		no-dso \
		no-engine \
		no-ocsp \
		no-posix-io \
		no-shared \
		no-sock \
		no-ssl \
		no-static-engine \
		no-tls \
		no-threads \
		no-ui-console \
		no-winstore \
		$(or $(if $(findstring i386-linux-musl,$1),linux-x86),\
		     $(if $(findstring x86_64-linux-musl,$1),linux-x86_64),\
		     $(if $(findstring aarch64-linux-musl,$1),linux-aarch64),\
		     $(if $(findstring i686-w64-mingw32,$1),mingw),\
		     $(if $(findstring x86_64-w64-mingw32,$1),mingw64))\
		$(if $(findstring $1,$(3rdparty_NATIVE_ARCH)),,--cross-compile-prefix=$1-) \
		-DOPENSSL_NO_SECURE_MEMORY \
		--prefix=$(abspath _3rdparty/tgt/$1)
	$(MAKE) -s -j$(3rdparty_JOBS) -C $$(@D)
	# LIBDIR overrides "multilib" settings
	$(MAKE) -s -C $$(@D) install_sw LIBDIR=lib
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
	find _3rdparty/tgt -type f -path '*darwin*/*.dylib' -delete
	touch $@
# Clean unpacked sources, build, install directory
3rdparty-clean:
	rm -rf 3rdparty-all.stamp _3rdparty/build _3rdparty/src _3rdparty/tgt
# Clean everything
3rdparty-distclean: 3rdparty-clean
	rm -rf _3rdparty/archive

# Save, restore binary artifacts into single tarfile (developer help)
# 3rdparty-artifact-id := \
# 	$(shell echo $(foreach pkg,$(3rdparty_TARGETS),$(pkg)=$($(pkg)_VERSION):) \
# 		| md5sum | awk '{print $$1}')
3rdparty-artifact-archive := _3rdparty/archive/artifacts-$(3rdparty-artifact-id).tar.gz

3rdparty-save-artifacts: 3rdparty-all
	$(TAR) -czf $(3rdparty-artifact-archive) \
		_3rdparty/tgt

3rdparty-restore-artifacts: $(3rdparty-artifact-archive)
	rm -rf _3rdparty/tgt
	$(TAR) -xzf $< _3rdparty/tgt
	# Regenerate stamp files so Make knows everything is current.
	mkdir -p \
		$(foreach pkg,$(3rdparty_TARGETS),_3rdparty/src/$(pkg)-$($(pkg)_VERSION)/)\
		$(foreach pkg,$(3rdparty_TARGETS),\
			$(foreach arch,$($(pkg)_ARCHS),\
				_3rdparty/build/$(arch)/$(pkg)-$($(pkg)_VERSION)/))
	touch \
		$(foreach pkg,$(3rdparty_TARGETS),_3rdparty/src/$(pkg)-$($(pkg)_VERSION)/.unpack-stamp)\
		$(foreach pkg,$(3rdparty_TARGETS),\
			$(foreach arch,$($(pkg)_ARCHS),\
				_3rdparty/build/$(arch)/$(pkg)-$($(pkg)_VERSION)/.build-stamp))
	touch 3rdparty-all.stamp

.PHONY: 3rdparty-save-artifacts 3rdparty-restore-artifacts

# Debug help
3rdparty-dump-templates:
	$(foreach pkg,$(3rdparty_TARGETS),$(info $(call download_TEMPLATE,$(pkg))))
	$(foreach pkg,$(3rdparty_TARGETS),$(info $(call unpack_TEMPLATE,$(pkg))))
	$(foreach pkg,$(3rdparty_TARGETS),\
		$(foreach arch,$($(pkg)_ARCHS),\
			$(info $(call build_$(pkg)_TEMPLATE,$(arch)))))

.PHONY: 3rdparty-all 3rdparty-clean 3rdparty-distclean 3rdparty-dump-templates
