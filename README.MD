# wgAdmin

A GUI application for managing WireGuard VPN configurations on Linux.

## Features

- Create, edit, import and delete WireGuard tunnels
- Activate/deactivate interfaces
- Generate key pairs and preshared keys
- Manage multiple peers per tunnel
- Network scanner for discovering hosts in a CIDR range
- Auto-backup before deletion
- Restore from backup config. 

## Requirements

- ~~Linux with WireGuard installed (`wg`, `wg-quick`)~~
- Linux with wireguard module loaded
- Root privileges

## Installation

```bash
git clone https://github.com/MrVasquez96/wgAdmin
cd wgAdmin
make install
```
1. Checks if `build/.build_done` file exists. 
2. Builds from source if missing and install destop app for linux.
4. Kopies binary, destop file and logo/ocon to dedicated folders for desktop apps in linux. 
3. The app should now be available in your desktop applications selection under `wgAdmin`.

## License

MIT
