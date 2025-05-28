# NGhost - NKN-based Decentralized VPN

NGhost is a cross-platform VPN application that uses NKN (New Kind of Network) technology for decentralized peer-to-peer networking, similar to Tailscale but built on NKN's decentralized infrastructure.

![NGhost Logo](https://img.shields.io/badge/NGhost-NKN%20VPN-blue?style=for-the-badge)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey?style=flat)

## ✨ Features

- 🌐 **Decentralized Networking**: Uses NKN for peer discovery and communication
- 🔄 **Cross-Platform**: Supports Linux, macOS, and Windows (WSL for Windows)
- 🚪 **Exit Node Support**: Can act as an exit node for internet traffic
- 🔒 **End-to-End Encryption**: All traffic is encrypted by default through NKN
- 🛡️ **No Public IP Required**: Works behind NATs and firewalls
- ⚡ **Network Agnostic**: No port forwarding needed
- 📊 **Peer Management**: Discover, manage and monitor VPN peers
- 🔧 **Auto-Discovery**: Automatic peer announcement and discovery
- 🌍 **Internet Routing**: Route traffic through exit nodes for internet access

## Architecture

- **NKN Layer**: Handles peer-to-peer communication and discovery
- **VPN Engine**: Manages TUN interface and packet routing
- **TUN Interface**: Cross-platform virtual network interface
- **Configuration**: JSON-based configuration management

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** for building from source
- **Root/Administrator privileges** for TUN interface creation
- **Linux/macOS** (Windows via WSL)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd nghost

# Build the application
go mod download
go build -o nghost

# Run demo to verify installation
./scripts/demo.sh
```

### Basic Usage

```bash
# Start VPN client (creates config.json on first run)
sudo ./nghost

# Run as exit node for others
sudo ./nghost -exit-node

# Run in background daemon mode
sudo ./nghost -daemon

# List discovered peers
./nghost -list-peers

# Add peer manually
./nghost -add-peer <nkn-address>

# Export peers to JSON
./nghost -export-peers peers.json
```

### First Time Setup

1. **Build the application**: `go build -o nghost`
2. **Generate config**: `./nghost -list-peers` (creates default config)
3. **Start VPN**: `sudo ./nghost`
4. **Share your NKN address** with peers or join a network

## Configuration

NGhost creates a default `config.json` file on first run:

```json
{
  "nkn": {
    "seedRPCServerAddr": [
      "http://seed1.nkn.org:30003",
      "http://seed2.nkn.org:30003",
      "http://seed3.nkn.org:30003"
    ]
  },
  "vpn": {
    "interfaceName": "nghost0",
    "cidr": "10.100.0.0/16",
    "mtu": 1420,
    "dns": ["1.1.1.1", "8.8.8.8"],
    "exitNodes": []
  }
}
```

## Platform-Specific Notes

### Linux
- Requires `/dev/net/tun` access
- Uses `ip` command for interface configuration

### macOS
- Uses `utun` devices
- Requires `route` and `ifconfig` commands

### Windows
- Requires WinTUN driver (not yet implemented)
- Use WSL for now

## 🔧 Advanced Usage

### Network Scenarios

**Scenario 1: Direct Peer Connection**
```bash
# On Machine A
sudo ./nghost
# Note your NKN address from output

# On Machine B  
./nghost -add-peer <machine-a-nkn-address>
sudo ./nghost
```

**Scenario 2: Exit Node Setup**
```bash
# On exit node (server with internet access)
sudo ./nghost -exit-node

# On clients (will auto-discover exit node)
sudo ./nghost
```

**Scenario 3: Network Monitoring**
```bash
# Real-time peer status
watch -n 2 './nghost -list-peers'

# Export network topology
./nghost -export-peers network-$(date +%Y%m%d).json
```

### Troubleshooting

- **Permission denied**: Ensure you're running with `sudo` for TUN interface
- **No peers found**: Check firewall settings and NKN connectivity
- **Routing issues**: Verify IP forwarding is enabled on exit nodes
- **macOS specific**: May need to approve network extensions in System Preferences

## 🌐 NKN Integration

NGhost leverages NKN's unique features:

- **Network Agnostic**: No need for public IPs or port forwarding
- **High Performance**: ~100ms latency, 10+ Mbps throughput
- **Decentralized**: No central servers or infrastructure
- **Secure**: End-to-end encryption by default
- **Global Reach**: Access to NKN's worldwide network infrastructure

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details