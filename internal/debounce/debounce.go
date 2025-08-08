package debounce

import (
	"crypto/md5"
	"encoding/hex"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	"github.com/evan-buss/opds-proxy/internal/cache"
	"github.com/evan-buss/opds-proxy/internal/httpx"
	"golang.org/x/sync/singleflight"
)

func NewDebounceMiddleware(debounce time.Duration) func(next http.HandlerFunc) http.HandlerFunc {
	responseCache := cache.NewCache[httptest.ResponseRecorder](cache.CacheConfig{CleanupInterval: time.Second, TTL: debounce})
	singleflight := singleflight.Group{}

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)
			hash := md5.Sum([]byte(ip + r.URL.Path + r.URL.RawQuery))
			key := string(hex.EncodeToString(hash[:]))

			if entry, exists := responseCache.Get(key); exists {
				w.Header().Set("X-Debounce", "true")
				httpx.WriteRecorder(entry, w)
				return
			}

			rw, _, shared := singleflight.Do(key, func() (interface{}, error) {
				rw := httptest.NewRecorder()
				next(rw, r)
				return rw, nil
			})

			recorder := rw.(*httptest.ResponseRecorder)
			responseCache.Set(key, recorder)

			w.Header().Set("X-Shared", strconv.FormatBool(shared))
			httpx.WriteRecorder(recorder, w)
		}
	}
}
