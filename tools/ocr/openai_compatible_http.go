package ocr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func postOpenAICompatibleJSON(ctx context.Context, client *http.Client, apiKey, url string, body any) (string, *OCRUsage, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return "", nil, newOCRError(CodeInvalidInput, "invalid OCR provider payload: "+err.Error())
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", nil, newOCRError(CodeConfigMissing, "invalid OCR provider URL: "+err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return "", nil, newOCRError(CodeProviderTimeout, err.Error())
		}
		return "", nil, newOCRError(CodeProviderError, err.Error())
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
	if err != nil {
		return "", nil, newOCRError(CodeProviderError, err.Error())
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", nil, newOCRError(CodeProviderError, fmt.Sprintf("HTTP %d: %s", resp.StatusCode, truncate(string(respBody), 512)))
	}

	var chatResp chatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err == nil && len(chatResp.Choices) > 0 {
		text := strings.TrimSpace(chatResp.Choices[0].Message.Content)
		var usage *OCRUsage
		if chatResp.Usage.PromptTokens > 0 || chatResp.Usage.CompletionTokens > 0 {
			usage = &OCRUsage{
				InputTokens:  chatResp.Usage.PromptTokens,
				OutputTokens: chatResp.Usage.CompletionTokens,
			}
		}
		return text, usage, nil
	}
	return string(respBody), nil, nil
}
