package nkn

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nknorg/nkn-sdk-go"
	"nghost/internal/config"
)

type Client struct {
	config      *config.NKNConfig
	client      *nkn.Client
	multiClient *nkn.MultiClient
	peers       map[string]*Peer
	peersMutex  sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	vpnEngine   VPNEngine
}

type VPNEngine interface {
	InjectPacket(packet []byte) error
}

type Peer struct {
	Address    string    `json:"address"`
	Online     bool      `json:"online"`
	LastSeen   time.Time `json:"lastSeen"`
	IPAddress  string    `json:"ipAddress"`
	ExitNode   bool      `json:"exitNode"`
	Latency    int64     `json:"latency"`
}

type ControlMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type PeerAnnouncement struct {
	IPAddress string `json:"ipAddress"`
	ExitNode  bool   `json:"exitNode"`
}

func NewClient(cfg config.NKNConfig) (*Client, error) {
	account, err := nkn.NewAccount(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create NKN account: %w", err)
	}

	clientConfig := &nkn.ClientConfig{}

	client, err := nkn.NewClient(account, "", clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create NKN client: %w", err)
	}

	multiClient, err := nkn.NewMultiClient(account, "", 4, false, clientConfig)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create NKN multi-client: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{
		config:      &cfg,
		client:      client,
		multiClient: multiClient,
		peers:       make(map[string]*Peer),
		ctx:         ctx,
		cancel:      cancel,
	}

	go c.handleMessages()

	return c, nil
}

func (c *Client) handleMessages() {
	// Wait for connection with timeout
	select {
	case <-c.multiClient.OnConnect.C:
		fmt.Printf("🔗 NKN client connected\n")
	case <-time.After(10 * time.Second):
		fmt.Printf("⚠️  NKN connection timeout, continuing anyway...\n")
	case <-c.ctx.Done():
		return
	}

	for {
		select {
		case <-c.ctx.Done():
			return
		case msg, ok := <-c.multiClient.OnMessage.C:
			if !ok {
				fmt.Printf("📡 NKN message channel closed\n")
				return
			}
			go c.processMessage(msg)
		}
	}
}

func (c *Client) processMessage(msg *nkn.Message) {
	// Handle different message types for VPN functionality
	if msg.Encrypted {
		// Handle VPN packet data
		c.handleVPNPacket(msg)
	} else {
		// Handle control messages
		c.handleControlMessage(msg)
	}
}

func (c *Client) handleVPNPacket(msg *nkn.Message) {
	// Forward received packet to TUN interface
	if c.vpnEngine != nil {
		if err := c.vpnEngine.InjectPacket(msg.Data); err != nil {
			fmt.Printf("Failed to inject packet: %v\n", err)
		}
	}
}

func (c *Client) handleControlMessage(msg *nkn.Message) {
	var controlMsg ControlMessage
	if err := json.Unmarshal(msg.Data, &controlMsg); err != nil {
		return
	}

	switch controlMsg.Type {
	case "peer_announcement":
		c.handlePeerAnnouncement(msg.Src, controlMsg.Payload)
	case "ping":
		c.handlePing(msg.Src)
	case "pong":
		c.handlePong(msg.Src)
	}
}

func (c *Client) handlePeerAnnouncement(src string, payload interface{}) {
	data, _ := json.Marshal(payload)
	var announcement PeerAnnouncement
	if err := json.Unmarshal(data, &announcement); err != nil {
		return
	}

	c.peersMutex.Lock()
	defer c.peersMutex.Unlock()

	peer, exists := c.peers[src]
	if !exists {
		peer = &Peer{Address: src}
		c.peers[src] = peer
		fmt.Printf("🆕 New peer discovered: %s\n", src[:16]+"...")
	}

	peer.IPAddress = announcement.IPAddress
	peer.ExitNode = announcement.ExitNode
	peer.Online = true
	peer.LastSeen = time.Now()

	fmt.Printf("📢 Peer %s announced: IP=%s, ExitNode=%v\n", src[:16]+"...", announcement.IPAddress, announcement.ExitNode)

	// Notify VPN engine about new peer route
	if c.vpnEngine != nil && announcement.IPAddress != "" {
		if routeEngine, ok := c.vpnEngine.(interface{ AddPeerRoute(string, string) error }); ok {
			routeEngine.AddPeerRoute(announcement.IPAddress, src)
		}
	}
}

func (c *Client) handlePing(src string) {
	pong := ControlMessage{
		Type:    "pong",
		Payload: time.Now().Unix(),
	}
	data, _ := json.Marshal(pong)
	c.multiClient.Send(nkn.NewStringArray(src), data, nil)
}

func (c *Client) handlePong(src string) {
	c.peersMutex.Lock()
	defer c.peersMutex.Unlock()

	if peer, exists := c.peers[src]; exists {
		peer.LastSeen = time.Now()
		peer.Online = true
	}
}

func (c *Client) SendPacket(dest string, data []byte) error {
	onMessage, err := c.multiClient.Send(nkn.NewStringArray(dest), data, nil)
	if err != nil {
		return err
	}
	_ = onMessage // Handle response if needed
	return nil
}

func (c *Client) GetAddress() string {
	return c.multiClient.Address()
}

func (c *Client) AddPeer(address string) {
	c.peersMutex.Lock()
	defer c.peersMutex.Unlock()
	c.peers[address] = &Peer{
		Address: address,
		Online:  false,
	}
	fmt.Printf("🔗 Added peer manually: %s\n", address[:16]+"...")
}

func (c *Client) GetPeers() map[string]*Peer {
	c.peersMutex.RLock()
	defer c.peersMutex.RUnlock()
	peers := make(map[string]*Peer)
	for k, v := range c.peers {
		peers[k] = v
	}
	return peers
}

func (c *Client) SetVPNEngine(engine VPNEngine) {
	c.vpnEngine = engine
}

func (c *Client) AnnouncePeer(ipAddress string, isExitNode bool) error {
	announcement := ControlMessage{
		Type: "peer_announcement",
		Payload: PeerAnnouncement{
			IPAddress: ipAddress,
			ExitNode:  isExitNode,
		},
	}
	data, err := json.Marshal(announcement)
	if err != nil {
		return err
	}

	// Broadcast to discovery address for peer discovery
	discoveryAddr := "nghost.discovery"
	_, err = c.multiClient.Send(nkn.NewStringArray(discoveryAddr), data, nil)
	
	// Also send to known peers for redundancy
	c.peersMutex.RLock()
	defer c.peersMutex.RUnlock()
	for _, peer := range c.peers {
		c.multiClient.Send(nkn.NewStringArray(peer.Address), data, nil)
	}
	
	return err
}

func (c *Client) FindExitNodes() []*Peer {
	c.peersMutex.RLock()
	defer c.peersMutex.RUnlock()

	var exitNodes []*Peer
	for _, peer := range c.peers {
		if peer.ExitNode && peer.Online {
			exitNodes = append(exitNodes, peer)
		}
	}
	return exitNodes
}

func (c *Client) Close() error {
	fmt.Printf("🔌 Closing NKN client...\n")
	
	// Cancel context first to stop goroutines
	c.cancel()
	
	// Give goroutines time to exit gracefully
	time.Sleep(100 * time.Millisecond)
	
	// Close clients
	if c.multiClient != nil {
		c.multiClient.Close()
	}
	if c.client != nil {
		c.client.Close()
	}
	
	fmt.Printf("✅ NKN client closed\n")
	return nil
}