package ui_ux

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/originaleric/digeino/config"
)

func TestNewChatModelFromConfigRequiresExplicitConfig(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	config.Set(config.Default())

	_, err := NewChatModelFromConfig(context.Background())
	if err == nil {
		t.Fatal("expected missing ChatModel.Type to fail")
	}
	if !strings.Contains(err.Error(), "ChatModel.Type is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewChatModelFromConfigFailsWhenEnvPlaceholderMissing(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	const envName = "DIGEINO_TEST_MISSING_QWEN_KEY"
	origEnv, hadEnv := os.LookupEnv(envName)
	_ = os.Unsetenv(envName)
	defer func() {
		if hadEnv {
			_ = os.Setenv(envName, origEnv)
		} else {
			_ = os.Unsetenv(envName)
		}
	}()

	cfg := config.Default()
	cfg.ChatModel = config.ChatModelConfig{
		Type: "qwen",
		Config: map[string]interface{}{
			"ApiKey":  "${" + envName + "}",
			"BaseUrl": "https://example.invalid/v1",
			"Model":   "qwen-test",
		},
	}
	config.Set(cfg)

	_, err := NewChatModelFromConfig(context.Background())
	if err == nil {
		t.Fatal("expected missing env placeholder to fail")
	}
	if !strings.Contains(err.Error(), "references missing environment variable "+envName) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewChatModelFromConfigRequiresExplicitModelAndBaseURL(t *testing.T) {
	orig := config.Get()
	defer config.Set(orig)

	cfg := config.Default()
	cfg.ChatModel = config.ChatModelConfig{
		Type: "qwen",
		Config: map[string]interface{}{
			"ApiKey": "test-key",
		},
	}
	config.Set(cfg)

	_, err := NewChatModelFromConfig(context.Background())
	if err == nil {
		t.Fatal("expected missing model/base URL to fail")
	}
	if !strings.Contains(err.Error(), "ChatModel.Config.BaseUrl is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
