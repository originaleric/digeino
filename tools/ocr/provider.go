package ocr

import (
	"context"
	"strings"
	"sync"
)

// OCRProvider OCR 模型提供者接口。
type OCRProvider interface {
	Recognize(ctx context.Context, req *OCRRequest, img *OCRImage) (*OCRResponse, error)
	Name() string
}

// OCRRegistry 管理 OCR provider，便于宿主或后续 provider 扩展。
type OCRRegistry interface {
	Register(provider OCRProvider)
	Get(name string) (OCRProvider, bool)
	Default() OCRProvider
}

type providerRegistry struct {
	mu          sync.RWMutex
	defaultName string
	providers   map[string]OCRProvider
}

// NewOCRRegistry 创建 OCR provider registry。
func NewOCRRegistry(defaultName string) OCRRegistry {
	return &providerRegistry{
		defaultName: normalizeProviderName(defaultName),
		providers:   map[string]OCRProvider{},
	}
}

func (r *providerRegistry) Register(provider OCRProvider) {
	if r == nil || provider == nil {
		return
	}
	name := normalizeProviderName(provider.Name())
	if name == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[name] = provider
	if r.defaultName == "" {
		r.defaultName = name
	}
}

func (r *providerRegistry) Get(name string) (OCRProvider, bool) {
	if r == nil {
		return nil, false
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[normalizeProviderName(name)]
	return p, ok
}

func (r *providerRegistry) Default() OCRProvider {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if p, ok := r.providers[r.defaultName]; ok {
		return p
	}
	for _, p := range r.providers {
		return p
	}
	return nil
}

func normalizeProviderName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func (r *providerRegistry) snapshot(defaultName string) *providerRegistry {
	clone := &providerRegistry{
		defaultName: normalizeProviderName(defaultName),
		providers:   map[string]OCRProvider{},
	}
	if r == nil {
		return clone
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if clone.defaultName == "" {
		clone.defaultName = r.defaultName
	}
	for name, provider := range r.providers {
		clone.providers[name] = provider
	}
	return clone
}

var globalProviders = NewOCRRegistry("")

// RegisterOCRProvider 注册包级 OCR provider，NewClient 会按 Tools.OCR.Provider 解析使用。
func RegisterOCRProvider(provider OCRProvider) {
	globalProviders.Register(provider)
}

// GetOCRProvider 获取已注册的包级 OCR provider。
func GetOCRProvider(name string) (OCRProvider, bool) {
	return globalProviders.Get(name)
}

func configuredProviderRegistry(defaultName string) *providerRegistry {
	if r, ok := globalProviders.(*providerRegistry); ok {
		return r.snapshot(defaultName)
	}
	return NewOCRRegistry(defaultName).(*providerRegistry)
}
