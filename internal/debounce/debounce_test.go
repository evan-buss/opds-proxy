package debounce

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestDebounceMiddleware(t *testing.T) {
	setup := func() (http.Handler, *int) {
		// Mock handler that simulates a slow response
		handlerCallCount := 0
		mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCallCount++
			time.Sleep(100 * time.Millisecond) // Simulate some work
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		middleware := NewDebounceMiddleware(500 * time.Millisecond)
		wrappedHandler := middleware(mockHandler)

		return wrappedHandler, &handlerCallCount
	}

	t.Run("Caching Behavior", func(t *testing.T) {
		wrappedHandler, handlerCallCount := setup()

		// First request
		req1 := httptest.NewRequest("GET", "/test", nil)
		rec1 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec1, req1)

		if *handlerCallCount != 1 {
			t.Errorf("Expected handler to be called once, got %d", handlerCallCount)
		}

		// Second request within debounce period
		req2 := httptest.NewRequest("GET", "/test", nil)
		rec2 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec2, req2)

		if *handlerCallCount != 1 {
			t.Errorf("Expected handler to still be called once, got %d", handlerCallCount)
		}

		if rec2.Header().Get("X-Debounce") != "true" {
			t.Error("Expected second response to be debounced")
		}

		// Wait for debounce period to expire
		time.Sleep(600 * time.Millisecond)

		// Third request after debounce period
		req3 := httptest.NewRequest("GET", "/test", nil)
		rec3 := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec3, req3)

		if *handlerCallCount != 2 {
			t.Errorf("Expected handler to be called twice, got %d", handlerCallCount)
		}
	})

	t.Run("Singleflight Behavior", func(t *testing.T) {
		wrappedHandler, handlerCallCount := setup()

		var wg sync.WaitGroup
		requestCount := 10

		for i := 0; i < requestCount; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest("GET", "/test", nil)
				rec := httptest.NewRecorder()
				wrappedHandler.ServeHTTP(rec, req)
			}()
		}

		wg.Wait()

		if *handlerCallCount != 1 {
			t.Errorf("Expected handler to be called once for concurrent requests, got %d", handlerCallCount)
		}
	})

	t.Run("Different Paths", func(t *testing.T) {
		wrappedHandler, handlerCallCount := setup()

		// Request to path A
		reqA := httptest.NewRequest("GET", "/testA", nil)
		recA := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recA, reqA)

		// Request to path B
		reqB := httptest.NewRequest("GET", "/testB", nil)
		recB := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(recB, reqB)

		if *handlerCallCount != 2 {
			t.Errorf("Expected handler to be called twice for different paths, got %d", handlerCallCount)
		}
	})
}
