
test:
	find . -type f -name '*.go' | sed -E 's|^(.*)/.*\.go|\1|' | sort | uniq | xargs go test

fmt:
	find . -type f -name '*.go' -exec go fmt ";"

build: build_dir build_linux build_windows build_darwin link_build

clean:
	rm -rf _bin/*
	rm -rf _output/*

##### BUILD #####
## BUILD VARIABLE ##
bin_name := lazykube
cmds := $(wildcard cmds/*.go)

## BUILD FUNCTION ##
build = GOOS=$(1) GOARCH=$(2) go build -o _bin/$(1)_$(2)_$(bin_name) $(cmds)

build_dir:
	[ -d _bin ] || mkdir _bin

##### LINUX BUILDS #####
build_linux: build/linux_arm build/linux_arm64 build/linux_386 build/linux_amd64

build/linux_386:
	$(call build,linux,386,)

build/linux_amd64:
	$(call build,linux,amd64,)

build/linux_arm:
	$(call build,linux,arm,)

build/linux_arm64:
	$(call build,linux,arm64,)

##### DARWIN (MAC) BUILDS #####
build_darwin: build/darwin_amd64

build/darwin_amd64:
	$(call build,darwin,amd64,)

##### WINDOWS BUILDS #####
build_windows: build/windows_386 build/windows_amd64

build/windows_386:
	$(call build,windows,386,)

build/windows_amd64:
	$(call build,windows,amd64,)

##### LINK BUILD FILE #####
uname := $(shell uname -s | tr A-Z a-z)
arch := $(shell uname -m | tr A-Z a-z)
link_build:
ifneq (,$(findstring $(arch),x86_64 i86_64 86_64 amd64))
	$(eval arch="amd64")
else ifneq (,$(findstring $(arch),aarch64 arm64))
	$(eval arch="arm64")
else ifneq (,$(findstring $(arch),arm))
	$(eval arch="arm")
else ifneq (,$(findstring $(arch),i86 86))
	$(eval arch="386")
endif
	[ -L _bin/$(bin_name) -o -f _bin/$(bin_name) ] && rm _bin/$(bin_name) || true
	ln -s  $(uname)_$(arch)_$(bin_name) _bin/$(bin_name)

