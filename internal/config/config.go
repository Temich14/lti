package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	LTIConfig   LTIConfig      `yaml:"lti"`
	Database    DatabaseConfig `yaml:"database"`
	OAuthConfig OAuthConfig    `yaml:"oauth"`
	RedisConfig RedisConfig    `yaml:"redis"`
}
type RedisConfig struct {
	TTL             time.Duration `yaml:"ttl"`
	Addr            string        `yaml:"addr"`
	Password        string        `yaml:"password"`
	DB              int           `yaml:"db"`
	LoginSessionKey string        `yaml:"session_key"`
	MaxSize         int           `yaml:"max_size"`
}
type LTIConfig struct {
	ToolName              string `yaml:"tool_name"`
	Domain                string `yaml:"domain"`
	RegistrationClientUri string `yaml:"registration_client_uri"`
	LoginClientUri        string `yaml:"login_client_uri"`
	ApplicationType       string `yaml:"application_type"`
}
type OAuthConfig struct {
	PrivateKeyPath string `yaml:"private_key_path"`
	KeyID          string `yaml:"key_id"`
	ClientID       string
	RedirectURI    string   `yaml:"redirect_uri"`
	JWKSUri        string   `yaml:"jwks_uri"`
	Scopes         []string `yaml:"scopes"`
}
type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
	MaxConns int    `yaml:"max_conns"`
	MinConns int    `yaml:"min_conns"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg AppWrapper
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg.App, nil
}

type AppWrapper struct {
	App Config `yaml:"app"`
}

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("failed to decode pem")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}
