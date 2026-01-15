package wx

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/originaleric/digeino/config"
)

var (
	tokenMutex sync.Mutex // 用于保护 Token 文件读写的互斥锁
)

// getTokenFilePath 获取 Token 文件完整路径
func getTokenFilePath() string {
	cfg := config.Get()
	if cfg.WeChat.TokenFilePath != "" {
		return cfg.WeChat.TokenFilePath
	}
	// 默认路径：当前工作目录下的 storage/app/wechat/access_token.json
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	return filepath.Join(wd, "storage/app/wechat/access_token.json")
}

// loadAccessTokenFromFile 从文件加载 AccessToken
func loadAccessTokenFromFile(filePath string) (*AccessTokenData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tokenData AccessTokenData
	err = json.Unmarshal(data, &tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token file: %w", err)
	}

	return &tokenData, nil
}

// saveAccessTokenToFile 保存 AccessToken 到文件
func saveAccessTokenToFile(filePath string, tokenData *AccessTokenData) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// isTokenExpired 检查 Token 是否过期
func isTokenExpired(tokenData *AccessTokenData) bool {
	// 提前 5 分钟刷新，避免边界情况
	bufferTime := int64(300) // 5 分钟
	return time.Now().Unix() >= (tokenData.ExpiresAt - bufferTime)
}

// fetchAccessTokenFromWeChat 从微信 API 获取新的 AccessToken
func fetchAccessTokenFromWeChat(ctx context.Context, appID, appSecret string) (*AccessTokenData, error) {
	// 微信获取 AccessToken 的 API
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		appID, appSecret)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call WeChat API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("WeChat API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResp WeChatTokenResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查是否有错误
	if apiResp.ErrCode != 0 {
		return nil, fmt.Errorf("WeChat API error: errcode=%d, errmsg=%s", apiResp.ErrCode, apiResp.ErrMsg)
	}

	if apiResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access_token in response")
	}

	// 构建 Token 数据
	now := time.Now().Unix()
	tokenData := &AccessTokenData{
		AccessToken: apiResp.AccessToken,
		ExpiresIn:   apiResp.ExpiresIn,
		CreatedAt:   now,
		ExpiresAt:   now + apiResp.ExpiresIn,
	}

	return tokenData, nil
}

// GetAccessToken 获取 AccessToken（带缓存和自动刷新）
func GetAccessToken(ctx context.Context) (string, error) {
	tokenMutex.Lock()
	defer tokenMutex.Unlock()

	cfg := config.Get()
	if cfg.WeChat.AppID == "" || cfg.WeChat.AppSecret == "" {
		return "", fmt.Errorf("WeChat AppID or AppSecret not configured")
	}

	// 1. 从配置读取文件路径
	tokenFilePath := getTokenFilePath()

	// 2. 尝试读取文件
	tokenData, err := loadAccessTokenFromFile(tokenFilePath)
	if err == nil && !isTokenExpired(tokenData) {
		// 文件存在且未过期，直接返回
		return tokenData.AccessToken, nil
	}

	// 3. 需要获取新 token
	appID := cfg.WeChat.AppID
	appSecret := cfg.WeChat.AppSecret

	// 4. 调用微信 API 获取新 token
	newTokenData, err := fetchAccessTokenFromWeChat(ctx, appID, appSecret)
	if err != nil {
		return "", err
	}

	// 5. 保存到文件
	err = saveAccessTokenToFile(tokenFilePath, newTokenData)
	if err != nil {
		// 即使保存失败，也返回 token（至少本次可以使用）
		fmt.Printf("[WeChat Warning] Failed to save token to file: %v\n", err)
		return newTokenData.AccessToken, nil
	}

	return newTokenData.AccessToken, nil
}
