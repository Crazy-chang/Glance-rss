// Package config 负责加载后端配置文件 config.yaml
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 全局配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	JWT      JWTConfig      `yaml:"jwt"`
	WeChat   WeChatConfig   `yaml:"wechat"`
	Fetch    FetchConfig    `yaml:"fetch"`
}

type ServerConfig struct {
	Port    string `yaml:"port"`
	BaseURL string `yaml:"base_url"`
	// RefreshInterval 全局定时抓取间隔（30m/1h/2h），不暴露给客户端
	RefreshInterval string `yaml:"refresh_interval"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	ExpireDays int    `yaml:"expire_days"`
}

type WeChatConfig struct {
	AppID  string `yaml:"appid"`
	Secret string `yaml:"secret"`
}

type FetchConfig struct {
	TimeoutSeconds int `yaml:"timeout_seconds"`
	Concurrency    int `yaml:"concurrency"`
}

// MustLoad 加载配置，失败直接 panic（服务启动阶段）
// 可通过环境变量 GLANCE_CONFIG 指定配置文件路径
func MustLoad() *Config {
	path := os.Getenv("GLANCE_CONFIG")
	if path == "" {
		path = "config/config.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("读取配置文件失败 %s: %v", path, err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("解析配置文件失败: %v", err))
	}

	// 默认值兜底
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	switch cfg.Server.RefreshInterval {
	case "30m", "1h", "2h":
	default:
		cfg.Server.RefreshInterval = "1h"
	}
	if cfg.JWT.ExpireDays <= 0 {
		cfg.JWT.ExpireDays = 30
	}
	if cfg.Fetch.TimeoutSeconds <= 0 {
		cfg.Fetch.TimeoutSeconds = 10
	}
	if cfg.Fetch.Concurrency <= 0 {
		cfg.Fetch.Concurrency = 5
	}

	return &cfg
}
