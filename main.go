package main

import (
	"flag"
	"log"
	"os"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type config struct {
	Port  string       `koanf:"port"`
	Feeds []feedConfig `koanf:"feeds" `
}

type feedConfig struct {
	Name string `koanf:"name"`
	Url  string `koanf:"url"`
}

var k = koanf.New(".")

func main() {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.String("port", "8080", "port to listen on")
	configPath := fs.String("config", "config.yml", "config file to load")
	if err := fs.Parse(os.Args[1:]); err != nil {
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

	server := NewServer(&config)
	server.Serve()
}
