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
	~/go/bin/fyne package \
		-os linux \
		-icon icon.png \
		-name wgAdmin
	tar -xvf wgAdmin.tar.xz
	mv wgAdmin/usr/local/bin/wgAdmin bin/wgAdmin
