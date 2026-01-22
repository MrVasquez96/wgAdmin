# GOLANG Build Targets

WIN_ARCH ?= amd64
 # or 386
BUILD_OS ?= LINUX
 
# --- Windows Cross-Compilation Target ---
.PHONY: win_build
# Builds the application for Windows (64-bit) and names the executable wgAdmin.exe.
# NOTE: CGO_ENABLED=0 is set, and the build tag `-tags=osd` is used to force Fyne
# to use its pure-Go software renderer, bypassing CGO dependencies (like go-gl)
# that require a C cross-compiler (mingw-w64).
# --- Existing Targets (Original) ---
# Rebuilds and runs the application in the current OS environment (e.g., Linux/macOS
   
# Builds the application for the current OS and runs with sudo.
linux:
	mkdir -p bin
	~/go/bin/fyne package \
		-os linux \
		-icon icon.png \
		-name wgAdmin
	tar -xvf wgAdmin.tar.xz
	mv wgAdmin/usr/local/bin/wgAdmin bin/wgAdmin


win_build:
	mkdir -p bin
	~/go/bin/fyne-cross \
		windows \
		-arch=$(WIN_ARCH) \
		-icon icon.png
	unzip -o fyne-cross/dist/windows-$(WIN_ARCH)/wgAdmin.exe.zip
	mv *.exe bin/
