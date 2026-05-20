package gateway

import (
	"time"

	"github.com/originaleric/digeino/config"
	"github.com/originaleric/digeino/gateway/artifact"
	"github.com/originaleric/digeino/gateway/executor"
	"github.com/originaleric/digeino/gateway/registry"
	"github.com/originaleric/digeino/gateway/runtime"
)

// RegistryOptions configures tool registration.
type RegistryOptions struct {
	ConfigDomains    []string
	AllowedReadPaths []string
	ArtifactStore    artifact.Store
}

// NewRegistry builds the default gateway tool registry from config.
func NewRegistry(cfg *config.Config, opts RegistryOptions) *registry.Registry {
	reg := registry.New()
	domains := opts.ConfigDomains
	if len(domains) == 0 {
		domains = cfg.Tools.LocalBrowser.AllowedDomains
	}
	store := opts.ArtifactStore
	readPaths := opts.AllowedReadPaths
	if len(readPaths) == 0 {
		readPaths = cfg.Gateway.AllowedReadPaths
	}

	reg.Register(executor.BrowserBrowseEntry(domains, store))
	reg.Register(executor.BrowserSnapshotEntry(domains))
	reg.Register(executor.BrowserActionEntry(domains))
	reg.Register(executor.WechatArticleReadEntry(domains, store))
	reg.Register(executor.XiaohongshuNoteReadEntry(domains, store))
	reg.Register(executor.DouyinVideoReadEntry(domains, store))
	reg.Register(executor.XPostReadEntry(domains, store))
	if len(readPaths) > 0 {
		reg.Register(executor.FileReadEntry(readPaths))
	}
	return reg
}

// NewArtifactStore creates a disk artifact store from gateway config.
func NewArtifactStore(cfg *config.Config) (artifact.Store, error) {
	dir := cfg.Gateway.ArtifactDir
	if dir == "" {
		dir = "storage/app/gateway_artifacts"
	}
	ttl := time.Duration(cfg.Gateway.ArtifactTTLMinutes) * time.Minute
	if ttl <= 0 {
		ttl = time.Hour
	}
	return artifact.NewDiskStore(dir, ttl)
}

// NewRuntime creates a runtime with registry and gateway options from config.
func NewRuntime(cfg *config.Config) *runtime.Runtime {
	gw := cfg.Gateway
	instanceID := gw.InstanceID
	if instanceID == "" {
		instanceID = "digeino-local"
	}
	var store artifact.Store
	if gw.ArtifactEnabled == nil || *gw.ArtifactEnabled {
		if s, err := NewArtifactStore(cfg); err == nil {
			store = s
		}
	}
	reg := NewRegistry(cfg, RegistryOptions{ArtifactStore: store})
	return runtime.New(reg, runtime.Options{
		InstanceID:    instanceID,
		AllowedTools:  gw.AllowedTools,
		ConfigDomains: cfg.Tools.LocalBrowser.AllowedDomains,
		ArtifactStore: store,
	})
}

// NewCollectorRuntime creates a runtime for the local Collector process.
func NewCollectorRuntime(cfg *config.Config) *runtime.Runtime {
	col := cfg.Collector
	instanceID := col.InstanceID
	if instanceID == "" {
		instanceID = cfg.Gateway.InstanceID
	}
	if instanceID == "" {
		instanceID = "digeino-collector"
	}
	allowed := col.AllowedTools
	if len(allowed) == 0 {
		allowed = cfg.Gateway.AllowedTools
	}
	var store artifact.Store
	if s, err := NewArtifactStore(cfg); err == nil {
		store = s
	}
	reg := NewRegistry(cfg, RegistryOptions{ArtifactStore: store})
	return runtime.New(reg, runtime.Options{
		InstanceID:    instanceID,
		AllowedTools:  allowed,
		ConfigDomains: cfg.Tools.LocalBrowser.AllowedDomains,
		ArtifactStore: store,
	})
}
