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
	Kafka       KafkaConfig    `yaml:"kafka"`
	EnrollmentSync EnrollmentSyncConfig `yaml:"enrollment_sync"`
	SessionService SessionServiceConfig `yaml:"session_service"`
}

type SessionServiceConfig struct {
	Enabled    bool   `yaml:"enabled"`
	GRPCTarget string `yaml:"grpc_target"` // host:port, plaintext (operator adds TLS as needed).
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

type KafkaConfig struct {
	Enabled     bool     `yaml:"enabled"`
	Brokers     []string `yaml:"brokers"`
	Topic       string   `yaml:"topic"`
	GroupID     string   `yaml:"group_id"`
	ClientID    string   `yaml:"client_id"`
	// Tuning knobs for consumer reads.
	MinBytes int `yaml:"min_bytes"`
	MaxBytes int `yaml:"max_bytes"`
}

type EnrollmentSyncConfig struct {
	// Batching for users.synchronization events.
	BatchSize int `yaml:"batch_size"`
	// How many outbox events to publish per poll iteration.
	OutboxBatchSize int `yaml:"outbox_batch_size"`
	OutboxPollInterval time.Duration `yaml:"outbox_poll_interval"`
	OutboxLockLease time.Duration `yaml:"outbox_lock_lease"`
	OutboxMaxAttempts int `yaml:"outbox_max_attempts"`

	// How often to pick up roster sync runs.
	RosterWorkerPollInterval time.Duration `yaml:"roster_worker_poll_interval"`
	RosterLockLease time.Duration `yaml:"roster_lock_lease"`

	// Consumer side: apply events into consumer inbox and domain tables.
	ConsumerEnabled bool `yaml:"consumer_enabled"`
	ConsumerConcurrency int `yaml:"consumer_concurrency"`
	ConsumerReadBatchSize int `yaml:"consumer_read_batch_size"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var raw bootFile
	if err := yaml.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	out := raw.App
	out.Kafka = raw.Kafka
	out.EnrollmentSync = raw.EnrollmentSync
	return &out, nil
}

type bootFile struct {
	App            Config               `yaml:"app"`
	Kafka          KafkaConfig          `yaml:"kafka"`
	EnrollmentSync EnrollmentSyncConfig `yaml:"enrollment_sync"`
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
