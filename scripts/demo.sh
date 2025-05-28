#!/bin/bash

# NGhost Demo Script
# This script demonstrates the basic functionality of NGhost VPN

echo "=== NGhost VPN Demo ==="
echo

# Build the application
echo "Building NGhost..."
go build -o nghost
echo "✅ Build complete"
echo

# Show help
echo "📖 Available commands:"
./nghost -h
echo

# Test configuration creation
echo "🔧 Testing configuration..."
./nghost -list-peers
echo

# Show peer management features
echo "📡 Peer management features:"
echo "  ./nghost -list-peers           # List discovered peers"
echo "  ./nghost -add-peer <address>   # Add peer manually"
echo "  ./nghost -export-peers out.json # Export peers to JSON"
echo

# Show VPN modes
echo "🌐 VPN operation modes:"
echo "  sudo ./nghost                  # Start VPN client"
echo "  sudo ./nghost -daemon          # Run as background daemon"
echo "  sudo ./nghost -exit-node       # Run as exit node for others"
echo

echo "🚀 Ready to use! Run 'sudo ./nghost' to start the VPN."
echo "⚠️  Note: Root privileges required for TUN interface creation."