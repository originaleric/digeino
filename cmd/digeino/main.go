package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/originaleric/digeino/config"
	"github.com/originaleric/digeino/gateway"
	"github.com/originaleric/digeino/gateway/collector"
	"github.com/originaleric/digeino/gateway/devhost"
	httpgw "github.com/originaleric/digeino/gateway/http"
	mcpgw "github.com/originaleric/digeino/gateway/mcp"
	stdiogw "github.com/originaleric/digeino/gateway/stdio"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "gateway":
		runGateway(os.Args[2:])
	case "collector":
		runCollector(os.Args[2:])
	case "dev-host":
		runDevHost(os.Args[2:])
	case "mcp":
		runMCP(os.Args[2:])
	case "stdio":
		runStdio(os.Args[2:])
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `DigEino — Universal Agent Plugin Runtime

Usage:
  digeino gateway [flags]     HTTP Tool Gateway
  digeino collector [flags]   WebSocket reverse connector
  digeino mcp [flags]         MCP server (stdio, for IDE)
  digeino stdio [flags]       JSON-line gateway on stdin/stdout
  digeino dev-host [flags]    Local reference host (dev only)
  digeino help

Host projects can import: github.com/originaleric/digeino/gateway/client

`)
}

func runGateway(args []string) {
	fs := flag.NewFlagSet("gateway", flag.ExitOnError)
	configPath := fs.String("config", "config/config.yaml", "path to config.yaml")
	addr := fs.String("addr", "", "listen address (overrides Gateway.ListenAddr)")
	_ = fs.Parse(args)

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	listen := cfg.Gateway.ListenAddr
	if *addr != "" {
		listen = *addr
	}
	if listen == "" {
		listen = ":8787"
	}

	rt := gateway.NewRuntime(cfg)
	srv := httpgw.NewServer(rt, rt.ArtifactStore(), cfg.Gateway.AuthToken)
	log.Printf("DigEino HTTP gateway listening on %s (instance=%s)", listen, cfg.Gateway.InstanceID)
	if err := srv.ListenAndServe(listen); err != nil {
		log.Fatalf("gateway server: %v", err)
	}
}

func runCollector(args []string) {
	fs := flag.NewFlagSet("collector", flag.ExitOnError)
	configPath := fs.String("config", "config/config.yaml", "path to config.yaml")
	server := fs.String("server", "", "host base URL (overrides Collector.ServerURL)")
	token := fs.String("token", "", "auth token (overrides Collector.Token)")
	instanceID := fs.String("instance-id", "", "collector instance id")
	pullSec := fs.Int("pull-interval", -1, "pull interval seconds; -1 uses config")
	_ = fs.Parse(args)

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	if *server != "" {
		cfg.Collector.ServerURL = *server
	}
	if *token != "" {
		cfg.Collector.Token = *token
	}
	if *instanceID != "" {
		cfg.Collector.InstanceID = *instanceID
	}
	if *pullSec >= 0 {
		cfg.Collector.PullIntervalSec = *pullSec
	}

	opts := collector.OptionsFromConfig(cfg)
	if opts.ServerURL == "" {
		log.Fatal("collector: --server or Collector.ServerURL is required")
	}

	rt := gateway.NewCollectorRuntime(cfg)
	client := collector.NewClient(opts, rt)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("DigEino collector connecting to %s instance=%s pull=%s",
		opts.ServerURL, opts.InstanceID, opts.PullInterval)
	if err := client.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("collector: %v", err)
	}
}

func runDevHost(args []string) {
	fs := flag.NewFlagSet("dev-host", flag.ExitOnError)
	addr := fs.String("addr", ":8790", "listen address")
	token := fs.String("token", "dev", "auth token for collectors")
	wsPath := fs.String("ws-path", "/digeino/v1/collector/ws", "WebSocket path")
	_ = fs.Parse(args)

	srv := devhost.NewServer(*token, *wsPath)
	log.Printf("DigEino dev-host (reference only — implement production host in your project)")
	if err := srv.ListenAndServe(*addr); err != nil {
		log.Fatalf("dev-host: %v", err)
	}
}

func runMCP(args []string) {
	fs := flag.NewFlagSet("mcp", flag.ExitOnError)
	configPath := fs.String("config", "config/config.yaml", "path to config.yaml")
	_ = fs.Parse(args)

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	rt := gateway.NewRuntime(cfg)
	log.Printf("DigEino MCP server on stdio (instance=%s)", cfg.Gateway.InstanceID)
	if err := mcpgw.ServeStdio(rt); err != nil {
		log.Fatalf("mcp: %v", err)
	}
}

func runStdio(args []string) {
	fs := flag.NewFlagSet("stdio", flag.ExitOnError)
	configPath := fs.String("config", "config/config.yaml", "path to config.yaml")
	_ = fs.Parse(args)

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	rt := gateway.NewRuntime(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := stdiogw.NewServer(rt).Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("stdio: %v", err)
	}
}

func loadConfig(path string) (*config.Config, error) {
	if path == "" {
		return config.Get(), nil
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Printf("config file %q not found, using defaults", path)
			return config.Get(), nil
		}
		return nil, err
	}
	return config.Load(path)
}
