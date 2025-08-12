package config

import "github.com/evan-buss/opds-proxy/internal/auth"

type FeedConfig struct {
	Name string
	Url  string
	Auth *FeedAuth
}

type FeedAuth = auth.FeedAuth

// Adapter methods (Get* names) to avoid field/method collisions
func (f FeedConfig) GetName() string    { return f.Name }
func (f FeedConfig) GetURL() string     { return f.Url }
func (f FeedConfig) GetAuth() *FeedAuth { return f.Auth }
