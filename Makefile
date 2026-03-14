.PHONY: all rebuild build install uninstall

TMP_DIR     := ./tmp_build
BUILD_DIR   := ./build
BUILD_EXISTS  := $(BUILD_DIR)/usr

INSTALL_REQUIREMENTS := \
    $(BUILD_DIR)/usr/local/bin/wgAdmin \
    $(BUILD_DIR)/usr/local/share/applications/wgAdmin.desktop \
    $(BUILD_DIR)/usr/local/share/pixmaps/wgAdmin.png

all: rebuild uninstall install

rebuild: clean build

build: $(BUILD_DONE)

$(BUILD_EXISTS):
	@echo "Starting build..."
	mkdir -p $(BUILD_DIR) $(TMP_DIR)
	~/go/bin/fyne release --target linux
	tar -xvf wgAdmin.tar.xz -C $(TMP_DIR)/ 
	rm -rf $(BUILD_DIR)/*
	mv $(TMP_DIR)/wgAdmin/* $(BUILD_DIR)/
	rm -rf $(TMP_DIR)
	touch $(BUILD_DONE)

install: $(BUILD_EXISTS)
	@make uninstall
	@if [ -f $(BUILD_DIR)/Makefile ]; then \
		$(MAKE) -C $(BUILD_DIR) user-install; \
	else \
		echo "Error: No Makefile found in $(BUILD_DIR) to run install"; \
		exit 1; \
	fi
uninstall: 
	-rm -rf $(HOME)/.local/share/applications/wgAdmin.desktop
	-rm -rf $(HOME)/.local/bin/wgAdmin
	-rm -rf $(HOME)/.local/share/icons/wgAdmin.png
	sudo rm -rf usr/share/applications/wgAdmin.desktop
	sudo rm -rf usr/bin/wgAdmin
	sudo rm -rf usr/share/pixmaps/wgAdmin.png
clean:
	rm -rf $(BUILD_DIR) $(TMP_DIR) *.tar.xz
