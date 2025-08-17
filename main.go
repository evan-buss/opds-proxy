package main

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/evan-buss/opds-proxy/internal/envextended"
	"github.com/gorilla/securecookie"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	flag "github.com/spf13/pflag"
)

// Version information set at build time
var version = "dev"
var commit = "unknown"
var date = "unknown"

type ProxyConfig struct {
	Port      string       `koanf:"port"`
	Auth      AuthConfig   `koanf:"auth"`
	Feeds     []FeedConfig `koanf:"feeds" `
	DebugMode bool         `koanf:"debug"`
}

type AuthConfig struct {
	HashKey  string `koanf:"hash_key"`
	BlockKey string `koanf:"block_key"`
}

type FeedConfig struct {
	Name string          `koanf:"name"`
	Url  string          `koanf:"url"`
	Auth *FeedConfigAuth `koanf:"auth"`
}

type FeedConfigAuth struct {
	Username  string `koanf:"username"`
	Password  string `koanf:"password"`
	LocalOnly bool   `koanf:"local_only"`
}

func main() {
	var k = koanf.New(".")

	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.StringP("port", "p", "8080", "port to listen on")
	fs.StringP("config", "c", "config.yml", "config file to load")
	fs.Bool("generate-keys", false, "generate cookie signing keys and exit")
	fs.BoolP("version", "v", false, "print version and exit")
	fs.Usage = func() {
		fmt.Println("Usage: opds-proxy [flags]")
		fmt.Println(fs.FlagUsages())
		os.Exit(0)
	}
	if err := fs.Parse(os.Args[1:]); err != nil {
		slog.Error("error parsing flags", slog.Any("error", err))
		os.Exit(1)
	}

	if showVersion, _ := fs.GetBool("version"); showVersion {
		fmt.Println("opds-proxy")
		fmt.Printf("  Version: %s\n", version)
		fmt.Printf("  Commit: %s\n", commit)
		fmt.Printf("  Build Date: %s\n", date)
		os.Exit(0)
	}

	if generate, _ := fs.GetBool("generate-keys"); generate {
		displayKeys()
		os.Exit(0)
	}

	// YAML Config
	configPath, _ := fs.GetString("config")
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil && !os.IsNotExist(err) {
		slog.Error("error loading config file", slog.Any("error", err))
		os.Exit(1)
	}

	// Environment Variables Config
	if err := k.Load(envextended.ProviderWithValue("OPDS", ".", envCallback), json.Parser()); err != nil {
		slog.Error("error loading environment variables", slog.Any("error", err))
		os.Exit(1)
	}

	// CLI Flags Config
	if err := k.Load(posflag.Provider(fs, ".", k), nil); err != nil {
		slog.Error("error loading CLI flags", slog.Any("error", err))
		os.Exit(1)
	}

	config := ProxyConfig{}
	k.Unmarshal("", &config)

	if config.Auth.HashKey == "" || config.Auth.BlockKey == "" {
		slog.Info("Generating new cookie signing credentials")
		hashKey, blockKey := displayKeys()

		config.Auth.HashKey = hashKey
		config.Auth.BlockKey = blockKey
	}

	if err := config.Validate(); err != nil {
		slog.Error("invalid configuration", slog.Any("error", err))
		os.Exit(1)
	}

	server, err := NewServer(&config)
	if err != nil {
		slog.Error("error creating server", slog.Any("error", err))
		os.Exit(1)
	}

	if err = server.Serve(); err != nil && err != http.ErrServerClosed {
		slog.Error("error serving", slog.Any("error", err))
		os.Exit(1)
	}
}

func displayKeys() (string, string) {
	hashKey := hex.EncodeToString(securecookie.GenerateRandomKey(32))
	blockKey := hex.EncodeToString(securecookie.GenerateRandomKey(32))

	slog.Info("Set these values in your config file to persist authentication between server restarts")
	fmt.Println("auth:")
	fmt.Printf("  hash_key: %s\n", hashKey)
	fmt.Printf("  block_key: %s\n", blockKey)

	return hashKey, blockKey
}

func envCallback(key string, value string) (string, interface{}) {
	key = strings.TrimPrefix(key, "OPDS__")
	key = strings.ReplaceAll(key, "__", ".")
	key = strings.ToLower(key)
	return key, value
}

func (c *ProxyConfig) Validate() error {
	if c.Port == "" {
		return errors.New("port is required")
	}

	if c.Auth.HashKey == "" || c.Auth.BlockKey == "" {
		return errors.New("auth.hash_key and auth.block_key are required")
	}

	if len(c.Feeds) == 0 {
		return errors.New("at least one feed must be defined")
	}

	for _, feed := range c.Feeds {
		if feed.Name == "" {
			return errors.New("feed.name is required")
		}

		if feed.Url == "" {
			return errors.New("feed.url is required")
		}
	}

	return nil
}
