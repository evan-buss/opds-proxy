package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type ProxyConfig struct {
	Port      string       `koanf:"port"`
	Auth      AuthConfig   `koanf:"auth"`
	Feeds     []FeedConfig `koanf:"feeds" `
	isDevMode bool
}

type AuthConfig struct {
	HashKey  string `koanf:"hash_key"`
	BlockKey string `koanf:"block_key"`
}

type FeedConfig struct {
	Name     string `koanf:"name"`
	Url      string `koanf:"url"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
}

func main() {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	// These aren't mapped to the config file.
	configPath := fs.String("config", "config.yml", "config file to load")
	generateKeys := fs.Bool("generate-keys", false, "generate cookie signing keys and exit")
	isDevMode := fs.Bool("dev", false, "enable development mode")

	port := fs.String("port", "8080", "port to listen on")
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	if *generateKeys {
		displayKeys()
		os.Exit(0)
	}

	var k = koanf.New(".")

	// Load config file from disk.
	// Feed options must be defined here.
	if err := k.Load(file.Provider(*configPath), yaml.Parser()); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Selectively add command line options to the config. Overriding the config file.
	if err := k.Load(confmap.Provider(map[string]interface{}{
		"port": *port,
	}, "."), nil); err != nil {
		log.Fatal(err)
	}

	config := ProxyConfig{}
	k.Unmarshal("", &config)

	if len(config.Feeds) == 0 {
		log.Fatal("No feeds defined in config")
	}

	if config.Auth.HashKey == "" || config.Auth.BlockKey == "" {
		log.Println("Generating new cookie signing credentials")
		hashKey, blockKey := displayKeys()

		config.Auth.HashKey = hashKey
		config.Auth.BlockKey = blockKey
	}

	// This should only be set by the command line flag,
	// so we don't use koanf to set this.
	config.isDevMode = *isDevMode
	server, err := NewServer(&config)
	if err != nil {
		log.Fatal(err)
	}

	if err = server.Serve(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func displayKeys() (string, string) {
	hashKey := hex.EncodeToString(securecookie.GenerateRandomKey(32))
	blockKey := hex.EncodeToString(securecookie.GenerateRandomKey(32))

	log.Println("Set these values in your config file to persist authentication between server restarts.")
	fmt.Println("auth:")
	fmt.Printf("  hash_key: %s\n", hashKey)
	fmt.Printf("  block_key: %s\n", blockKey)

	return hashKey, blockKey
}
