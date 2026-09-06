.PHONY: all rebuild build install uninstall

FYNE    := /home/$(USER)/go/bin/fyne
APPID   := com.wireguard.manager
EXEC    := wgAdmin
ICON    := com.wireguard.manager.png
BUILDDIR := wgAdmin

all: rebuild uninstall install

rebuild:
	rm -rf $(BUILDDIR)
	@$(MAKE) build

build:
	@echo "Starting wgAdmin..."
	$(FYNE) release --target linux
	mkdir -p $(BUILDDIR)
	tar -xvf wgAdmin.tar.xz -C $(BUILDDIR)

install:
	@if [ -f $(BUILDDIR)/Makefile ]; then \
		$(MAKE) -C $(BUILDDIR) user-install; \
	else \
		echo "Error: No Makefile found in $(BUILDDIR)"; \
		exit 1; \
	fi
	rm -rf $(BUILDDIR)

uninstall:
	-rm -f $(HOME)/.local/share/applications/$(APPID).desktop
	-rm -f $(HOME)/.local/bin/$(EXEC)
	-rm -f $(HOME)/.local/share/icons/$(ICON)
	-sudo rm -f /usr/local/share/applications/$(APPID).desktop
	-sudo rm -f /usr/local/bin/$(EXEC)
	-sudo rm -f /usr/local/share/pixmaps/$(ICON)
