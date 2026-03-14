.PHONY: all build install

TMP_DIR     := ./tmp_build
BUILD_DIR   := ./build
BUILD_DONE  := $(BUILD_DIR)/.build_done

INSTALL_REQUIREMENTS := \
    $(BUILD_DIR)/usr/local/bin/wgAdmin \
    $(BUILD_DIR)/usr/local/share/applications/wgAdmin.desktop \
    $(BUILD_DIR)/usr/local/share/pixmaps/wgAdmin.png

all: build install

build: $(BUILD_DONE)

$(BUILD_DONE):
	@echo "Starting build..."
	mkdir -p $(BUILD_DIR) $(TMP_DIR)
	~/go/bin/fyne release --target linux
	tar -xvf wgAdmin.tar.xz -C $(TMP_DIR)/ 
	rm -rf $(BUILD_DIR)/*
	mv $(TMP_DIR)/wgAdmin/* $(BUILD_DIR)/
	rm -rf $(TMP_DIR)
	touch $(BUILD_DONE)

install: $(BUILD_DONE)
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
	-rm -rf $(HOME)/.local/share/icons/$(Icon)
	sudo rm -rf usr/share/applications/wgAdmin.desktop
	sudo rm -rf usr/bin/wgAdmin
	sudo rm -rf usr/share/pixmaps/wgAdmin.png
clean:
	rm -rf $(BUILD_DIR) $(TMP_DIR) *.tar.xz