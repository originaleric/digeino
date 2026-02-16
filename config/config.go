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
	WeCom      WeComConfig      `yaml:"WeCom" json:"WeCom"`
	ChatModel  ChatModelConfig  `yaml:"ChatModel" json:"ChatModel"`
	UIUX       UIUXConfig       `yaml:"UIUX" json:"UIUX"`
	Tools      ToolsConfig      `yaml:"Tools" json:"Tools"`
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
	Enabled       *bool    `yaml:"Enabled" json:"Enabled,omitempty"`             // 是否启用微信推送功能
	AppID         string   `yaml:"AppID" json:"AppID,omitempty"`                 // 微信服务号 AppID
	AppSecret     string   `yaml:"AppSecret" json:"AppSecret,omitempty"`         // 微信服务号 AppSecret
	OpenIDs       []string `yaml:"OpenIDs" json:"OpenIDs,omitempty"`             // 默认接收消息的用户 openid 列表
	TokenFilePath string   `yaml:"TokenFilePath" json:"TokenFilePath,omitempty"` // AccessToken 存储文件路径（相对于项目根目录）
	// 小程序相关配置（用于发送小程序卡片消息）
	MiniProgram MiniProgramConfig `yaml:"MiniProgram" json:"MiniProgram,omitempty"`
}

// MiniProgramConfig 小程序配置
type MiniProgramConfig struct {
	AppID        string `yaml:"AppID" json:"AppID,omitempty"`               // 小程序 AppID
	DefaultPath  string `yaml:"DefaultPath" json:"DefaultPath,omitempty"`   // 默认页面路径，例如 "pages/index/index"
	ThumbMediaID string `yaml:"ThumbMediaID" json:"ThumbMediaID,omitempty"` // 小程序卡片封面图片的 media_id
}

// WeComConfig 企业微信配置
type WeComConfig struct {
	Enabled       *bool              `yaml:"Enabled" json:"Enabled,omitempty"`
	CorpID        string             `yaml:"CorpID" json:"CorpID,omitempty"`
	QYAPIHost     string             `yaml:"QYAPIHost" json:"QYAPIHost,omitempty"`
	TokenFilePath string             `yaml:"TokenFilePath" json:"TokenFilePath,omitempty"`
	Applications  []WeComApplication `yaml:"Applications" json:"Applications,omitempty"`
}

// WeComApplication 企业微信应用配置
type WeComApplication struct {
	AgentID            int64  `yaml:"AgentID" json:"AgentID"`
	AgentSecret        string `yaml:"AgentSecret" json:"AgentSecret,omitempty"`
	ManageAllKFSession bool   `yaml:"ManageAllKFSession" json:"ManageAllKFSession,omitempty"` // 是否管理所有客服会话，用于发送客服消息给个人微信用户
}

// ChatModelConfig ChatModel 配置（参考 DigFlow 的配置方式）
type ChatModelConfig struct {
	Type   string                 `yaml:"Type" json:"Type"`     // qwen, openai
	Config map[string]interface{} `yaml:"Config" json:"Config"` // 模型配置（ApiKey, Model, BaseUrl 等）
}

// UIUXConfig UI/UX 工具配置
type UIUXConfig struct {
	Storage UIUXStorageConfig `yaml:"Storage" json:"Storage"`
}

// UIUXStorageConfig UI/UX 存储配置
type UIUXStorageConfig struct {
	BaseDir string `yaml:"BaseDir" json:"BaseDir"` // 存储基础目录，默认为 "storage/app/ui_ux"
	// 如果不同应用/agent 需要隔离，可以在调用时传入 AppName
	// 存储路径为: {BaseDir}/{app-name}/design-system/{project}/MASTER.md
	// 如果未指定 app-name，则为: {BaseDir}/design-system/{project}/MASTER.md
}

// ToolsConfig 工具配置集
type ToolsConfig struct {
	Firecrawl    FirecrawlConfig    `yaml:"Firecrawl" json:"Firecrawl"`
	WebSearch    WebSearchConfig    `yaml:"WebSearch" json:"WebSearch"`
	Unstructured UnstructuredConfig `yaml:"Unstructured" json:"Unstructured"`
	Pinecone     PineconeConfig     `yaml:"Pinecone" json:"Pinecone"`
	Embedding    EmbeddingConfig    `yaml:"Embedding" json:"Embedding"`
}

// FirecrawlConfig Firecrawl 深度爬取配置
type FirecrawlConfig struct {
	ApiKey string `yaml:"ApiKey" json:"ApiKey"`
}

// WebSearchConfig 网页搜索配置
type WebSearchConfig struct {
	Engine     string           `yaml:"Engine" json:"Engine"` // bocha, serpapi, google, bing, duckduckgo
	Bocha      BochaConfig      `yaml:"Bocha" json:"Bocha"`
	SerpApi    SerpApiConfig    `yaml:"SerpApi" json:"SerpApi"`
	Google     GoogleConfig     `yaml:"Google" json:"Google"`
	Bing       BingConfig       `yaml:"Bing" json:"Bing"`
	DuckDuckGo DuckDuckGoConfig `yaml:"DuckDuckGo" json:"DuckDuckGo"`
}

// BochaConfig 博查搜索配置
type BochaConfig struct {
	ApiKey  string `yaml:"ApiKey" json:"ApiKey"`
	BaseUrl string `yaml:"BaseUrl" json:"BaseUrl"`
}

// SerpApiConfig SerpApi 搜索配置
type SerpApiConfig struct {
	ApiKey  string `yaml:"ApiKey" json:"ApiKey"`
	BaseUrl string `yaml:"BaseUrl" json:"BaseUrl"`
}

// GoogleConfig Google Custom Search 配置
type GoogleConfig struct {
	ApiKey  string `yaml:"ApiKey" json:"ApiKey"`
	Cx      string `yaml:"Cx" json:"Cx"` // Custom Search Engine ID
	BaseUrl string `yaml:"BaseUrl" json:"BaseUrl"`
}

// BingConfig Bing 搜索配置
type BingConfig struct {
	ApiKey     string `yaml:"ApiKey" json:"ApiKey"`
	MaxResults int    `yaml:"MaxResults" json:"MaxResults"`
	Region     string `yaml:"Region" json:"Region"`
	SafeSearch string `yaml:"SafeSearch" json:"SafeSearch"`
}

// DuckDuckGoConfig DuckDuckGo 搜索配置
type DuckDuckGoConfig struct {
	MaxResults int    `yaml:"MaxResults" json:"MaxResults"`
	Region     string `yaml:"Region" json:"Region"`
	SafeSearch string `yaml:"SafeSearch" json:"SafeSearch"`
}

// UnstructuredConfig Unstructured.io 文档解析配置
type UnstructuredConfig struct {
	ApiKey  string `yaml:"ApiKey" json:"ApiKey"`
	BaseUrl string `yaml:"BaseUrl" json:"BaseUrl"` // 默认为 https://api.unstructured.io/general/v0/general
}

// PineconeConfig Pinecone 向量数据库配置
type PineconeConfig struct {
	ApiKey    string `yaml:"ApiKey" json:"ApiKey"`
	Host      string `yaml:"Host" json:"Host"`
	Namespace string `yaml:"Namespace" json:"Namespace"`
	IndexName string `yaml:"IndexName" json:"IndexName"`
}

// EmbeddingConfig 向量化模型配置
type EmbeddingConfig struct {
	Type       string `yaml:"Type" json:"Type"` // openai, qwen
	ApiKey     string `yaml:"ApiKey" json:"ApiKey"`
	BaseUrl    string `yaml:"BaseUrl" json:"BaseUrl"`
	Model      string `yaml:"Model" json:"Model"`
	Dimensions int    `yaml:"Dimensions" json:"Dimensions"`
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
		WeCom: WeComConfig{
			Enabled:       &wechatDisabled,
			QYAPIHost:     "https://qyapi.weixin.qq.com",
			TokenFilePath: "storage/app/wecom/access_token.json",
		},
		ChatModel: ChatModelConfig{
			Type: "qwen",
			Config: map[string]interface{}{
				"ApiKey":  "",
				"Model":   "qwen-max",
				"BaseUrl": "https://dashscope.aliyuncs.com/compatible-mode/v1",
			},
		},
		UIUX: UIUXConfig{
			Storage: UIUXStorageConfig{
				BaseDir: "storage/app/ui_ux",
			},
		},
		Tools: ToolsConfig{
			Firecrawl: FirecrawlConfig{
				ApiKey: "",
			},
			WebSearch: WebSearchConfig{
				Engine: "bocha",
				Bocha: BochaConfig{
					BaseUrl: "https://api.bochaai.com",
				},
				SerpApi: SerpApiConfig{
					BaseUrl: "https://serpapi.com/search",
				},
				Google: GoogleConfig{
					BaseUrl: "https://www.googleapis.com/customsearch/v1",
				},
				Bing: BingConfig{
					MaxResults: 10,
				},
				DuckDuckGo: DuckDuckGoConfig{
					MaxResults: 10,
				},
			},
			Unstructured: UnstructuredConfig{
				ApiKey:  "",
				BaseUrl: "https://api.unstructured.io/general/v0/general",
			},
			Pinecone: PineconeConfig{
				ApiKey:    "",
				Namespace: "default",
			},
			Embedding: EmbeddingConfig{
				Type:       "qwen",
				Model:      "text-embedding-v2",
				Dimensions: 1024,
				BaseUrl:    "https://dashscope.aliyuncs.com/compatible-mode/v1",
			},
		},
	}
}
