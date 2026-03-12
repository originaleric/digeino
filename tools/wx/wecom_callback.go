package wx

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/originaleric/digeino/config"
)

// SyncCustomerMessages 拉取企业微信客服消息
func SyncCustomerMessages(ctx context.Context, req SyncMessageRequest) (SyncMessageResponse, error) {
	cfg := config.Get()
	if cfg.WeCom.CorpID == "" {
		return SyncMessageResponse{}, fmt.Errorf("WeCom CorpID not configured")
	}

	// 获取 access_token
	accessToken, err := getWeComCustomerAccessToken(ctx)
	if err != nil {
		return SyncMessageResponse{}, fmt.Errorf("failed to get access token: %w", err)
	}

	// 构建请求
	baseURL := getWeComAPIHost()
	url := fmt.Sprintf("%s/cgi-bin/kf/sync_msg?access_token=%s", baseURL, accessToken)

	requestBody := map[string]interface{}{
		"token": req.Token,
	}
	if req.Limit > 0 {
		requestBody["limit"] = req.Limit
	} else {
		requestBody["limit"] = 1000 // 默认值
	}
	if req.Cursor != "" {
		requestBody["cursor"] = req.Cursor
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return SyncMessageResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return SyncMessageResponse{}, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return SyncMessageResponse{}, fmt.Errorf("call sync_msg API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return SyncMessageResponse{}, fmt.Errorf("read response: %w", err)
	}

	var apiResp SyncMessageResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return SyncMessageResponse{}, fmt.Errorf("parse response: %w", err)
	}

	if apiResp.ErrCode != 0 {
		return SyncMessageResponse{}, fmt.Errorf("sync_msg API errcode=%d errmsg=%s", apiResp.ErrCode, apiResp.ErrMsg)
	}

	return apiResp, nil
}

// WeComCallbackHandler 企业微信回调处理器
type WeComCallbackHandler struct {
	onMessage func(CustomerMessage) error // 消息处理回调函数
}

// NewWeComCallbackHandler 创建回调处理器
func NewWeComCallbackHandler() *WeComCallbackHandler {
	return &WeComCallbackHandler{}
}

// OnMessage 设置消息处理回调函数
func (h *WeComCallbackHandler) OnMessage(fn func(CustomerMessage) error) {
	h.onMessage = fn
}

// VerifyURL 验证回调 URL（GET 请求）
func (h *WeComCallbackHandler) VerifyURL(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	callbackCfg := cfg.WeCom.Callback

	if callbackCfg.Enabled == nil || !*callbackCfg.Enabled {
		http.Error(w, "callback not enabled", http.StatusBadRequest)
		return
	}

	// 获取查询参数
	msgSignature := r.URL.Query().Get("msg_signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")
	echostr := r.URL.Query().Get("echostr")

	if msgSignature == "" || timestamp == "" || nonce == "" || echostr == "" {
		http.Error(w, "missing required parameters", http.StatusBadRequest)
		return
	}

	// 验证签名
	if !verifySignature(callbackCfg.Token, timestamp, nonce, echostr, msgSignature) {
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	// 解密 echostr
	decrypted, err := decrypt(callbackCfg.EncodingAESKey, cfg.WeCom.CorpID, echostr)
	if err != nil {
		http.Error(w, fmt.Sprintf("decrypt failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 返回解密后的字符串
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(decrypted))
}

// HandleMessage 处理回调消息（POST 请求）
func (h *WeComCallbackHandler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	cfg := config.Get()
	callbackCfg := cfg.WeCom.Callback

	if callbackCfg.Enabled == nil || !*callbackCfg.Enabled {
		http.Error(w, "callback not enabled", http.StatusBadRequest)
		return
	}

	// 获取查询参数
	msgSignature := r.URL.Query().Get("msg_signature")
	timestamp := r.URL.Query().Get("timestamp")
	nonce := r.URL.Query().Get("nonce")

	if msgSignature == "" || timestamp == "" || nonce == "" {
		http.Error(w, "missing required parameters", http.StatusBadRequest)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("read body failed: %v", err), http.StatusBadRequest)
		return
	}

	// 解析 XML
	var encryptedMsg struct {
		XMLName    xml.Name `xml:"xml"`
		ToUserName string   `xml:"ToUserName"`
		Encrypt    string   `xml:"Encrypt"`
	}
	if err := xml.Unmarshal(body, &encryptedMsg); err != nil {
		http.Error(w, fmt.Sprintf("parse XML failed: %v", err), http.StatusBadRequest)
		return
	}

	// 验证签名
	if !verifySignature(callbackCfg.Token, timestamp, nonce, encryptedMsg.Encrypt, msgSignature) {
		http.Error(w, "invalid signature", http.StatusBadRequest)
		return
	}

	// 解密消息
	decryptedXML, err := decrypt(callbackCfg.EncodingAESKey, cfg.WeCom.CorpID, encryptedMsg.Encrypt)
	if err != nil {
		http.Error(w, fmt.Sprintf("decrypt failed: %v", err), http.StatusInternalServerError)
		return
	}

	// 解析解密后的 XML
	var msg struct {
		XMLName      xml.Name `xml:"xml"`
		ToUserName   string   `xml:"ToUserName"`
		FromUserName string   `xml:"FromUserName"`
		CreateTime   int64    `xml:"CreateTime"`
		MsgType      string   `xml:"MsgType"`
		Event        string   `xml:"Event"`
		Content      string   `xml:"Content"`
		MsgID        string   `xml:"MsgId"`
		Token        string   `xml:"Token"`
		OpenKfId     string   `xml:"OpenKfId"`
		// 其他消息类型字段...
	}

	if err := xml.Unmarshal([]byte(decryptedXML), &msg); err != nil {
		http.Error(w, fmt.Sprintf("parse decrypted XML failed: %v", err), http.StatusBadRequest)
		return
	}

	// 处理消息
	if msg.MsgType == "event" && msg.Event == "kf_msg_or_event" {
		// 事件通知，需要调用拉取接口
		if msg.Token != "" {
			syncReq := SyncMessageRequest{
				Token: msg.Token,
				Limit: 1000,
			}
			syncResp, err := SyncCustomerMessages(r.Context(), syncReq)
			if err != nil {
				http.Error(w, fmt.Sprintf("sync messages failed: %v", err), http.StatusInternalServerError)
				return
			}

			// 处理拉取到的消息
			if h.onMessage != nil {
				for _, customerMsg := range syncResp.MsgList {
					if err := h.onMessage(customerMsg); err != nil {
						// 记录错误但继续处理其他消息
						// TODO: 添加日志记录
					}
				}
			}
		}
	} else {
		// 直接包含消息内容，实时处理
		if h.onMessage != nil && msg.MsgType == "text" {
			customerMsg := CustomerMessage{
				MsgID:          msg.MsgID,
				ExternalUserID: msg.FromUserName,
				SendTime:       msg.CreateTime,
				Origin:         3, // 客户发送
				MsgType:        "text",
				Text: &TextMessage{
					Content: msg.Content,
				},
			}
			if msg.OpenKfId != "" {
				customerMsg.OpenKfId = msg.OpenKfId
			}

			if err := h.onMessage(customerMsg); err != nil {
				http.Error(w, fmt.Sprintf("handle message failed: %v", err), http.StatusInternalServerError)
				return
			}
		}
	}

	// 返回成功响应（必须在5秒内返回）
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("success"))
}

// verifySignature 验证签名
func verifySignature(token, timestamp, nonce, encryptedMsg, msgSignature string) bool {
	// 对 token、timestamp、nonce 和 encryptedMsg 进行字典序排序
	arr := []string{token, timestamp, nonce, encryptedMsg}
	sort.Strings(arr)

	// 拼接字符串
	str := strings.Join(arr, "")

	// SHA1 加密
	h := sha1.New()
	h.Write([]byte(str))
	hash := h.Sum(nil)

	// 转换为十六进制字符串
	signature := fmt.Sprintf("%x", hash)

	return signature == msgSignature
}

// decrypt 解密消息
func decrypt(encodingAESKey, corpID, encryptedMsg string) (string, error) {
	// Base64 解码 AESKey（43位Base64编码，需要补1个=）
	var aesKey []byte
	var err error
	// 尝试直接解码
	aesKey, err = base64.StdEncoding.DecodeString(encodingAESKey)
	if err != nil {
		// 如果失败，尝试补一个=
		aesKey, err = base64.StdEncoding.DecodeString(encodingAESKey + "=")
		if err != nil {
			return "", fmt.Errorf("decode AESKey failed: %w", err)
		}
	}

	if len(aesKey) != 32 {
		return "", fmt.Errorf("invalid AESKey length: expected 32 bytes, got %d", len(aesKey))
	}

	// Base64 解码加密消息
	encrypted, err := base64.StdEncoding.DecodeString(encryptedMsg)
	if err != nil {
		return "", fmt.Errorf("decode encrypted message failed: %w", err)
	}

	// AES 解密
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", fmt.Errorf("new cipher failed: %w", err)
	}

	// IV 取 AESKey 前16字节
	iv := aesKey[:16]
	mode := cipher.NewCBCDecrypter(block, iv)

	// 解密
	if len(encrypted)%aes.BlockSize != 0 {
		return "", fmt.Errorf("encrypted message length is not a multiple of block size")
	}
	decrypted := make([]byte, len(encrypted))
	mode.CryptBlocks(decrypted, encrypted)

	// PKCS7 去填充
	decrypted = pkcs7Unpad(decrypted)

	// 提取消息内容（去掉随机数和长度）
	// 格式：random(16字节) + msg_len(4字节，网络字节序) + msg + corpID
	if len(decrypted) < 20 {
		return "", fmt.Errorf("decrypted message too short")
	}

	// 解析消息长度（网络字节序，大端）
	msgLen := int(decrypted[16])<<24 | int(decrypted[17])<<16 | int(decrypted[18])<<8 | int(decrypted[19])
	if msgLen < 0 || len(decrypted) < 20+msgLen {
		return "", fmt.Errorf("invalid message length: %d", msgLen)
	}

	msg := string(decrypted[20 : 20+msgLen])

	// 验证 corpID（最后部分）
	if len(decrypted) < 20+msgLen+len(corpID) {
		return "", fmt.Errorf("invalid message format: missing corpID")
	}
	corpIDInMsg := string(decrypted[20+msgLen:])
	if corpIDInMsg != corpID {
		return "", fmt.Errorf("corpID mismatch: expected %s, got %s", corpID, corpIDInMsg)
	}

	return msg, nil
}

// pkcs7Unpad PKCS7 去填充
func pkcs7Unpad(data []byte) []byte {
	if len(data) == 0 {
		return data
	}

	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return data
	}

	// 验证填充
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return data
		}
	}

	return data[:len(data)-padding]
}
