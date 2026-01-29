package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SdmVersion string   `yaml:"sdm"`
	SdmProto   string   `yaml:"sdm-proto"`
	UserProtos []string `yaml:"user-protos"`
	Output     string   `yaml:"output"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
