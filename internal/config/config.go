package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	NKN NKNConfig `json:"nkn"`
	VPN VPNConfig `json:"vpn"`
}

type NKNConfig struct {
	SeedRPCServerAddr []string `json:"seedRPCServerAddr"`
	ClientConfig      struct {
		SeedRPCServerAddr []string `json:"seedRPCServerAddr"`
		RPCTimeout        int      `json:"rpcTimeout"`
		RPCConcurrency    int      `json:"rpcConcurrency"`
	} `json:"clientConfig"`
}

type VPNConfig struct {
	InterfaceName string   `json:"interfaceName"`
	CIDR          string   `json:"cidr"`
	MTU           int      `json:"mtu"`
	DNS           []string `json:"dns"`
	ExitNodes     []string `json:"exitNodes"`
}

func Load(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := &Config{
			NKN: NKNConfig{
				SeedRPCServerAddr: []string{
					"http://seed1.nkn.org:30003",
					"http://seed2.nkn.org:30003",
					"http://seed3.nkn.org:30003",
				},
			},
			VPN: VPNConfig{
				InterfaceName: "nghost0",
				CIDR:          "10.100.0.0/16",
				MTU:           1420,
				DNS:           []string{"1.1.1.1", "8.8.8.8"},
				ExitNodes:     []string{},
			},
		}
		return cfg, Save(cfg, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Save(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}