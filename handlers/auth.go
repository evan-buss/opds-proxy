package handlers

import (
	"io"
	"net/http"
	"net/url"

	"github.com/evan-buss/opds-proxy/internal/auth"
	"github.com/evan-buss/opds-proxy/view"
	"github.com/gorilla/securecookie"
)

func Auth(s *securecookie.SecureCookie) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		returnUrl := r.URL.Query().Get("return")
		if returnUrl == "" {
			http.Error(w, "No return URL specified", http.StatusBadRequest)
			return
		}

		if r.Method == "GET" {
			view.Render(w, func(buf io.Writer) error { return view.Login(buf, view.LoginParams{ReturnURL: returnUrl}) })
			return
		}

		if r.Method == "POST" {
			username := r.FormValue("username")
			password := r.FormValue("password")

			rUrl, err := url.Parse(returnUrl)
			if err != nil {
				http.Error(w, "Invalid return URL", http.StatusBadRequest)
				return
			}
			domain, err := url.Parse(rUrl.Query().Get("q"))
			if err != nil {
				http.Error(w, "Invalid site", http.StatusBadRequest)
				return
			}

			value := map[string]auth.Credentials{
				domain.Hostname(): {Username: username, Password: password},
			}

			encoded, err := s.Encode(auth.CookieName, value)
			if err != nil {
				http.Error(w, "Failed to encode credentials", http.StatusInternalServerError)
				return
			}
			cookie := &http.Cookie{
				Name:  auth.CookieName,
				Value: encoded,
				Path:  "/",
				// Kobo fails to set cookies with HttpOnly or Secure flags
				Secure:   false,
				HttpOnly: false,
			}

			http.SetCookie(w, cookie)
			http.Redirect(w, r, returnUrl, http.StatusFound)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
