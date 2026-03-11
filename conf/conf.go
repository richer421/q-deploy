package conf

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var C Config

type Config struct {
	Server ServerConfig `yaml:"server"`
	MySQL  MySQLConfig  `yaml:"mysql"`
	Redis  RedisConfig  `yaml:"redis"`
	Kafka  KafkaConfig  `yaml:"kafka"`
	Log    LogConfig    `yaml:"log"`
	OTel   OTelConfig   `yaml:"otel"`
}

type ServerConfig struct {
	Port int `yaml:"port"`
}

type MySQLConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	User         string `yaml:"user"`
	Password     string `yaml:"password"`
	Database     string `yaml:"database"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
	MaxLifetime  int    `yaml:"max_lifetime"`
}

func (m *MySQLConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.Database)
}

// DSNWithoutDB 返回不带数据库名的 DSN（用于创建数据库）
func (m *MySQLConfig) DSNWithoutDB() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&parseTime=True&loc=Local",
		m.User, m.Password, m.Host, m.Port)
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	PoolSize int    `yaml:"pool_size"`
}

type KafkaConfig struct {
	Brokers    []string `yaml:"brokers"`
	MaxRetries int      `yaml:"max_retries"`
}

type LogConfig struct {
	Level  string        `yaml:"level"`
	Format string        `yaml:"format"`
	File   LogFileConfig `yaml:"file"`
}

type LogFileConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Path     string `yaml:"path"`
	MaxSize  int    `yaml:"max_size"`
	MaxAge   int    `yaml:"max_age"`
	Compress bool   `yaml:"compress"`
}

type OTelConfig struct {
	Enabled     bool             `yaml:"enabled"`
	ServiceName string           `yaml:"service_name"`
	Endpoint    string           `yaml:"endpoint"`
	Prometheus  PrometheusConfig `yaml:"prometheus"`
}

type PrometheusConfig struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(data, &C); err != nil {
		return err
	}
	applyEnvOverrides()
	return nil
}

// applyEnvOverrides 从环境变量覆盖配置（用于 Docker 部署）
func applyEnvOverrides() {
	if v := os.Getenv("MYSQL_HOST"); v != "" {
		C.MySQL.Host = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		C.Redis.Addr = v
	}
	if v := os.Getenv("KAFKA_BROKERS"); v != "" {
		C.Kafka.Brokers = []string{v}
	}
	if v := os.Getenv("OTEL_ENDPOINT"); v != "" {
		C.OTel.Endpoint = v
	}
	if v := os.Getenv("OTEL_ENABLED"); v == "true" {
		C.OTel.Enabled = true
	}
}
