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
	~/go/bin/fyne release \
		--target linux \
		-icon icon.png \
		-name wgAdmin-app \
		-appVersion 0.1.0 \
		-tags "wireguard,wireguardmanager,wg,wg-quick,wireguard-tools,wireguard-dkms,wireguard-go,wireguard-apple,wireguard-android,wireguard-windows,wg-gen-web,wg-dashboard,subspace,firezone,netbird,tailscale,pivpn,wg-ui,wgpia,udp,endpoint,allowed-ips,persistent-keepalive,preshared-key,roaming,nat-traversal,split-tunneling,kill-switch,mtu-optimization,ip-forwarding,site-to-site,point-to-point,mesh-network,chacha20,poly1305,curve25519,blake2s,perfect-forward-secrecy,noise-protocol-framework,cryptokey-routing,linuxserver-wireguard,kubernetes-wireguard,kilo,cloud-init-vpn,kernel-module,systemd-networkd,network-manager-wireguard,resolvconf,iptables,nftables,post-up,post-down,wg-dynamic"
	tar -xvf wgAdmin-app.tar.xz
	cp wgAdmin/usr/local/bin/wgAdmin bin/wgAdmin

