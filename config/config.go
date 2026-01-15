package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config 根配置结构
type Config struct {
	HttpServer HttpServerConfig `yaml:"HttpServer" json:"HttpServer"`
	Status     StatusConfig     `yaml:"Status" json:"Status"`
	WeChat     WeChatConfig     `yaml:"WeChat" json:"WeChat"`
}

// HttpServerConfig HTTP 服务配置
type HttpServerConfig struct {
	Api struct {
		Port string `yaml:"Port" json:"Port"`
	} `yaml:"Api" json:"Api"`
}

// StatusConfig 状态相关配置：包含 Webhook 与 Store
type StatusConfig struct {
	Webhook AppWebhookConfig  `yaml:"Webhook" json:"Webhook"`
	Store   StatusStoreConfig `yaml:"Store" json:"Store"`
}

// AppWebhookConfig 应用级 Webhook 配置
type AppWebhookConfig struct {
	Enabled *bool         `yaml:"Enabled" json:"Enabled,omitempty"`
	URL     string        `yaml:"URL" json:"URL,omitempty"`
	Config  WebhookConfig `yaml:",inline" json:"Config"`
}

// WebhookConfig 基础 Webhook 配置
type WebhookConfig struct {
	URL        string            `yaml:"url" json:"url"`
	Method     string            `yaml:"method" json:"method,omitempty"`
	Headers    map[string]string `yaml:"headers" json:"headers,omitempty"`
	Secret     string            `yaml:"secret" json:"secret,omitempty"`
	Timeout    int               `yaml:"timeout" json:"timeout,omitempty"`
	RetryCount int               `yaml:"retry_count" json:"retry_count,omitempty"`
	RetryDelay int               `yaml:"retry_delay" json:"retry_delay,omitempty"`
	Events     []string          `yaml:"events" json:"events,omitempty"`
}

// StatusStoreConfig 状态存储配置
type StatusStoreConfig struct {
	Enabled *bool       `yaml:"Enabled" json:"Enabled,omitempty"`
	Type    string      `yaml:"Type" json:"Type,omitempty"` // memory, mysql
	MySQL   MySQLConfig `yaml:"MySQL" json:"MySQL"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	Host        string `yaml:"Host"`
	Port        int    `yaml:"Port"`
	User        string `yaml:"User"`
	Password    string `yaml:"Password"`
	Database    string `yaml:"Database"`
	ExecTable   string `yaml:"ExecTable"`
	StatusTable string `yaml:"StatusTable"`
}

// WeChatConfig 微信推送配置
type WeChatConfig struct {
	Enabled     *bool    `yaml:"Enabled" json:"Enabled,omitempty"`         // 是否启用微信推送功能
	AppID       string   `yaml:"AppID" json:"AppID,omitempty"`             // 微信服务号 AppID
	AppSecret   string   `yaml:"AppSecret" json:"AppSecret,omitempty"`     // 微信服务号 AppSecret
	OpenIDs     []string `yaml:"OpenIDs" json:"OpenIDs,omitempty"`          // 默认接收消息的用户 openid 列表
	TokenFilePath string `yaml:"TokenFilePath" json:"TokenFilePath,omitempty"` // AccessToken 存储文件路径（相对于项目根目录）
}

var (
	// DefaultConfig 全局默认配置
	currentConfig *Config
)

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	currentConfig = &cfg
	return &cfg, nil
}

// Get 获取当前配置
func Get() *Config {
	if currentConfig == nil {
		currentConfig = Default()
	}
	return currentConfig
}

// Set 设置当前配置
func Set(cfg *Config) {
	currentConfig = cfg
}

// Default 返回默认配置
func Default() *Config {
	enabled := true
	wechatDisabled := false
	return &Config{
		HttpServer: HttpServerConfig{
			Api: struct {
				Port string `yaml:"Port" json:"Port"`
			}{
				Port: ":20201",
			},
		},
		Status: StatusConfig{
			Webhook: AppWebhookConfig{
				Enabled: &enabled,
				Config: WebhookConfig{
					Method:     "POST",
					Timeout:    5,
					RetryCount: 3,
					RetryDelay: 1000,
				},
			},
			Store: StatusStoreConfig{
				Enabled: &enabled,
				Type:    "memory",
			},
		},
		WeChat: WeChatConfig{
			Enabled:       &wechatDisabled,
			TokenFilePath: "storage/app/wechat/access_token.json",
		},
	}
}
