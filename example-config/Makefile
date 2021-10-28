all: spyre.zip

clean:
	rm -f spyre.zip spyre-self-contained.exe spyre-self-contained

spyre.zip: spyre.yaml file-rules.yara proc-rules.yara common.yara
	zip $@ -Pinfected $^

self-contained: spyre-self-contained.exe spyre-self-contained

spyre-self-contained: ../_build/x86_64-linux-musl/spyre spyre.zip
	cat $^ > $@.t
	chmod 755 $@.t
	mv $@.t $@

spyre-self-contained.exe: ../_build/i686-w64-mingw32/spyre.exe spyre.zip
	cat $^ > $@.t
	mv $@.t $@

.PHONY: all clean self-contained
