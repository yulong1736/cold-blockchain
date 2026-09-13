package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	JWT        JWTConfig        `yaml:"jwt"`
	Session    SessionConfig    `yaml:"session"`
	SMTP       SMTPConfig       `yaml:"smtp"`
	Blockchain BlockchainConfig `yaml:"blockchain"`
	Logging    LoggingConfig    `yaml:"logging"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Mode string `yaml:"mode"`
}

type DatabaseConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	DBName         string `yaml:"dbname"`
	SSLMode        string `yaml:"sslmode"`
	MaxConnections int    `yaml:"max_connections"`
}

type JWTConfig struct {
	Secret       string `yaml:"secret"`
	ExpiresHours int    `yaml:"expires_hours"`
}

type SessionConfig struct {
	IdleTimeoutSeconds int `yaml:"idle_timeout_seconds"`
}

type SMTPConfig struct {
	Host               string `yaml:"host"`
	Port               int    `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	From               string `yaml:"from"`
	UseTLS             bool   `yaml:"use_tls"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify"`
}

type BlockchainConfig struct {
	FISCOBCOS FISCOBCOSConfig `yaml:"fisco_bcos"`
}

type FISCOBCOSConfig struct {
	NodeURL         string `yaml:"node_url"`
	GroupID         int    `yaml:"group_id"`
	ChainID         int    `yaml:"chain_id"`
	PrivateKey      string `yaml:"private_key"`
	ContractAddress string `yaml:"contract_address"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

var AppConfig *Config

func LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	AppConfig = &Config{}
	if err := yaml.Unmarshal(data, AppConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Docker/环境变量覆盖：便于 docker-compose 中通过 environment 指定数据库等
	if v := os.Getenv("DB_HOST"); v != "" {
		AppConfig.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			AppConfig.Database.Port = p
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		AppConfig.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		AppConfig.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		AppConfig.Database.DBName = v
	}
	if v := os.Getenv("SMTP_HOST"); v != "" {
		AppConfig.SMTP.Host = v
	}
	if v := os.Getenv("SMTP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			AppConfig.SMTP.Port = p
		}
	}
	if v := os.Getenv("SMTP_USER"); v != "" {
		AppConfig.SMTP.User = v
	}
	if v := os.Getenv("SMTP_PASSWORD"); v != "" {
		AppConfig.SMTP.Password = v
	}
	if v := os.Getenv("SMTP_FROM"); v != "" {
		AppConfig.SMTP.From = v
	}
	if v := os.Getenv("SMTP_USE_TLS"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			AppConfig.SMTP.UseTLS = b
		}
	}
	if v := os.Getenv("SMTP_INSECURE_SKIP_VERIFY"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			AppConfig.SMTP.InsecureSkipVerify = b
		}
	}
	if v := os.Getenv("BLOCKCHAIN_NODE_URL"); v != "" {
		AppConfig.Blockchain.FISCOBCOS.NodeURL = v
	}
	if v := os.Getenv("BLOCKCHAIN_PRIVATE_KEY"); v != "" {
		AppConfig.Blockchain.FISCOBCOS.PrivateKey = v
	}
	if v := os.Getenv("BLOCKCHAIN_CONTRACT_ADDRESS"); v != "" {
		AppConfig.Blockchain.FISCOBCOS.ContractAddress = v
	}

	return nil
}

func GetDSN() string {
	db := AppConfig.Database
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		db.Host, db.Port, db.User, db.Password, db.DBName, db.SSLMode)
}
