package config

import "testing"

func TestDefaultConfigYAMLDoesNotConfigureChatModel(t *testing.T) {
	orig := Get()
	defer Set(orig)

	cfg, err := Load("config.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.ChatModel.Type != "" {
		t.Fatalf("expected default config.yaml to leave ChatModel.Type empty, got %q", cfg.ChatModel.Type)
	}
	if len(cfg.ChatModel.Config) != 0 {
		t.Fatalf("expected default config.yaml to leave ChatModel.Config empty, got %+v", cfg.ChatModel.Config)
	}
}
