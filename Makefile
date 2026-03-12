# GOLANG Build Targets

WIN_ARCH ?= amd64
 # or 386
BUILD_OS ?= LINUX
.PHONY: all
all: linux

.PHONY: dev
dev:
	go build -o ./bin/wgAdmin .
	sudo ./bin/wgAdmin

.PHONY: linux 
linux:
	mkdir -p bin 
	~/go/bin/fyne release --target linux
	tar -xvf wgAdmin.tar.xz -C bin/

