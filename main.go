package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type config struct {
	Port  string       `koanf:"port"`
	Auth  auth         `koanf:"auth"`
	Feeds []feedConfig `koanf:"feeds" `
}

type auth struct {
	HashKey  string `koanf:"hash_key"`
	BlockKey string `koanf:"block_key"`
}

type feedConfig struct {
	Name     string `koanf:"name"`
	Url      string `koanf:"url"`
	Username string `koanf:"username"`
	Password string `koanf:"password"`
}

func (f feedConfig) HasCredentials() bool {
	return f.Username != "" && f.Password != ""
}

var k = koanf.New(".")

func main() {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String("port", "8080", "port to listen on")
	configPath := fs.String("config", "config.yml", "config file to load")
	generateKeys := fs.Bool("generate-keys", false, "generate cookie signing keys and exit")
	if err := fs.Parse(os.Args[1:]); err != nil {
		log.Fatal(err)
	}

	if *generateKeys {
		displayKeys()
		os.Exit(0)
	}

	// Load config file from disk.
	// Feed options must be defined here.
	if err := k.Load(file.Provider(*configPath), yaml.Parser()); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Flags take precedence over config file.
	if err := k.Load(basicflag.Provider(fs, "."), nil); err != nil {
		log.Fatal(err)
	}

	config := config{}
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

	server, err := NewServer(&config)
	if err != nil {
		log.Fatal(err)
	}
	server.Serve()
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
