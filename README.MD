# wgAdmin

A GUI application for managing WireGuard VPN configurations on Linux.

## Features

- Create, edit, and delete WireGuard tunnels
- Activate/deactivate interfaces with one click
- Generate key pairs and preshared keys
- Manage multiple peers per tunnel
- Network scanner for discovering hosts in a CIDR range
- Auto-backup before deletion

## Requirements

- Linux with WireGuard installed (`wg`, `wg-quick`)
- Root privileges

## Installation
Either download the latest release from the releases [page](https://github.com/MrVasquez96/wgAdmin/releases) or build from source:
```bash
git clone https://github.com/MrVasquez96/wgAdmin
cd wgAdmin
go build -o wgAdmin
sudo ./wgAdmin
```

## Usage

Run with root privileges:

```bash
sudo ./wgAdmin
```

The application reads and writes configuration files from `/etc/wireguard/`.

## License

MIT
