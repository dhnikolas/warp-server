package config

import (
	"github.com/go-playground/validator/v10"
	_ "github.com/go-playground/validator/v10"
	yaml "gopkg.in/yaml.v2"
	"os"
)

type Config struct {
	CiscoHost     string `validate:"required" yaml:"cisco_host"`
	CiscoUsername string `validate:"required" yaml:"cisco_username"`
	CiscoPassword string `validate:"required" yaml:"cisco_password"`
	LocalUsername string `validate:"required" yaml:"local_username"`
	LocalPassword string `validate:"required" yaml:"local_password"`
	LocalHost     string `validate:"required" yaml:"localhost"`
	TunnelAddress string `validate:"required" yaml:"tunnel_address"`
	DaemonMode    bool   `validate:"omitempty" yaml:"daemon_mode"`
	VpnOnly       bool   `validate:"omitempty" yaml:"vpn_only"`
}

func LoadConfig() (*Config, error) {
	homedir, _ := os.UserHomeDir()
	file, err := os.ReadFile(homedir + "/" + ".warp-server.yaml")
	if err != nil {
		return nil, err
	}

	cfg := &Config{}

	if err := yaml.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}
	validate := validator.New()
	err = validate.Struct(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
