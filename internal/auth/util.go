package auth

import (
	"net/http"
	"net/url"

	"github.com/evan-buss/opds-proxy/internal/reqctx"
	"github.com/gorilla/securecookie"
)

type Credentials struct {
	Username string
	Password string
}

const CookieName = "auth-creds"

func GetCredentials(rawUrl string, req *http.Request, feeds []FeedConfigLike, s *securecookie.SecureCookie) *Credentials {
	requestUrl, err := url.Parse(rawUrl)
	if err != nil {
		return nil
	}

	// Try to get credentials from the config first
	for _, feed := range feeds {
		feedUrl, err := url.Parse(feed.GetURL())
		if err != nil {
			continue
		}

		if feedUrl.Hostname() != requestUrl.Hostname() {
			continue
		}

		cfg := feed.GetAuth()
		if cfg == nil || cfg.Username == "" || cfg.Password == "" {
			continue
		}

		// Only set feed credentials for local requests
		// when the auth config has local_only flag
		isLocal := reqctx.IsLocal(req.Context())
		if !isLocal && cfg.LocalOnly {
			continue
		}

		return &Credentials{Username: cfg.Username, Password: cfg.Password}
	}

	// Otherwise, try to get credentials from the cookie
	cookie, err := req.Cookie(CookieName)
	if err != nil {
		return nil
	}

	value := make(map[string]*Credentials)
	if err = s.Decode(CookieName, cookie.Value, &value); err != nil {
		return nil
	}

	return value[requestUrl.Hostname()]
}

// FeedAuth mirrors config authentication for a feed
// and is used by the FeedConfigLike adapter interface.
type FeedAuth struct {
	Username  string
	Password  string
	LocalOnly bool
}

type FeedConfigLike interface {
	GetName() string
	GetURL() string
	GetAuth() *FeedAuth
}
