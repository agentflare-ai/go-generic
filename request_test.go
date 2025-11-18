package generic

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRequestWithContext_NewRequestWithContext(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		ctx := context.Background()
		req, err := NewRequestWithContext(ctx, "GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req == nil {
			t.Fatal("expected non-nil request")
		}
		if req.Method != "GET" {
			t.Errorf("expected method GET, got %s", req.Method)
		}
		if req.URL.String() != "http://example.com" {
			t.Errorf("expected URL http://example.com, got %s", req.URL.String())
		}
	})

	t.Run("with body", func(t *testing.T) {
		ctx := context.Background()
		body := strings.NewReader("test body")
		req, err := NewRequestWithContext(ctx, "POST", "http://example.com", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if req.Method != "POST" {
			t.Errorf("expected method POST, got %s", req.Method)
		}
		if req.Body == nil {
			t.Error("expected non-nil body")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		ctx := context.Background()
		_, err := NewRequestWithContext(ctx, "GET", ":", nil)
		if err == nil {
			t.Fatal("expected error for invalid URL")
		}
	})
}

func TestRequestWithContext_Context(t *testing.T) {
	t.Run("successful context access", func(t *testing.T) {
		type CustomContext struct {
			context.Context
			UserID string
		}

		baseCtx := context.Background()
		customCtx := CustomContext{Context: baseCtx, UserID: "user123"}

		req, err := NewRequestWithContext(customCtx, "GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrievedCtx := req.Context()
		customRetrieved, ok := retrievedCtx.(CustomContext)
		if !ok {
			t.Fatal("expected CustomContext type")
		}
		if customRetrieved.UserID != "user123" {
			t.Errorf("expected UserID 'user123', got %s", customRetrieved.UserID)
		}
	})

	t.Run("context type mismatch panics", func(t *testing.T) {
		// Create request with standard context
		baseCtx := context.Background()
		req, err := NewRequestWithContext(baseCtx, "GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// This should panic because we're trying to access as a different type
		defer func() {
			if r := recover(); r == nil {
				t.Fatal("expected panic for context type mismatch")
			}
		}()

		type OtherContext struct {
			context.Context
		}
		_ = req.Context().(OtherContext) // This will panic in the method
	})
}

func TestRequestWithContext_ForwardedMethods(t *testing.T) {
	ctx := context.Background()
	req, err := NewRequestWithContext(ctx, "GET", "http://example.com/test?param=value", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("method access", func(t *testing.T) {
		if req.Method != "GET" {
			t.Errorf("expected method GET, got %s", req.Method)
		}
		if req.URL.String() != "http://example.com/test?param=value" {
			t.Errorf("expected URL 'http://example.com/test?param=value', got %s", req.URL.String())
		}
		if req.Proto != "HTTP/1.1" {
			t.Errorf("expected proto HTTP/1.1, got %s", req.Proto)
		}
	})

	t.Run("cookies", func(t *testing.T) {
		// Add a cookie
		cookie := &http.Cookie{Name: "test", Value: "value", Path: "/"}
		req.AddCookie(cookie)

		// Retrieve cookie
		retrieved, err := req.Cookie("test")
		if err != nil {
			t.Fatalf("unexpected error getting cookie: %v", err)
		}
		if retrieved.Value != "value" {
			t.Errorf("expected cookie value 'value', got %s", retrieved.Value)
		}

		// Get all cookies
		cookies := req.Cookies()
		if len(cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(cookies))
		}
	})

	t.Run("form values", func(t *testing.T) {
		// Create a request with form data
		form := url.Values{}
		form.Add("key1", "value1")
		form.Add("key2", "value2")
		body := strings.NewReader(form.Encode())
		req, err := NewRequestWithContext(ctx, "POST", "http://example.com", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Parse form
		if err := req.ParseForm(); err != nil {
			t.Fatalf("unexpected error parsing form: %v", err)
		}

		// Test form values
		if val := req.FormValue("key1"); val != "value1" {
			t.Errorf("expected form value 'value1', got %s", val)
		}
		if val := req.PostFormValue("key2"); val != "value2" {
			t.Errorf("expected post form value 'value2', got %s", val)
		}
	})

	t.Run("path values", func(t *testing.T) {
		req.SetPathValue("id", "123")
		if val := req.PathValue("id"); val != "123" {
			t.Errorf("expected path value '123', got %s", val)
		}
	})

	t.Run("basic auth", func(t *testing.T) {
		req.SetBasicAuth("user", "pass")
		username, password, ok := req.BasicAuth()
		if !ok {
			t.Fatal("expected basic auth to be set")
		}
		if username != "user" {
			t.Errorf("expected username 'user', got %s", username)
		}
		if password != "pass" {
			t.Errorf("expected password 'pass', got %s", password)
		}
	})

	t.Run("user agent and referer", func(t *testing.T) {
		req.Header.Set("User-Agent", "test-agent")
		req.Header.Set("Referer", "http://referer.com")

		if ua := req.UserAgent(); ua != "test-agent" {
			t.Errorf("expected user agent 'test-agent', got %s", ua)
		}
		if ref := req.Referer(); ref != "http://referer.com" {
			t.Errorf("expected referer 'http://referer.com', got %s", ref)
		}
	})

	t.Run("protocol version", func(t *testing.T) {
		if !req.ProtoAtLeast(1, 0) {
			t.Error("expected protocol at least 1.0")
		}
		if !req.ProtoAtLeast(1, 1) {
			t.Error("expected protocol at least 1.1")
		}
		if req.ProtoAtLeast(2, 0) {
			t.Error("expected protocol not at least 2.0")
		}
	})

	t.Run("clone", func(t *testing.T) {
		newCtx := context.WithValue(ctx, "key", "value")
		cloned := req.Clone(newCtx)
		if cloned.Method != req.Method {
			t.Errorf("expected cloned method %s, got %s", req.Method, cloned.Method)
		}
		if cloned.URL.String() != req.URL.String() {
			t.Errorf("expected cloned URL %s, got %s", req.URL.String(), cloned.URL.String())
		}
	})

	t.Run("with context", func(t *testing.T) {
		newCtx := context.WithValue(ctx, "newkey", "newvalue")
		withCtx := req.WithContext(newCtx)
		if withCtx.Method != req.Method {
			t.Errorf("expected with context method %s, got %s", req.Method, withCtx.Method)
		}
	})
}

func TestRequestWithContext_MultipartForm(t *testing.T) {
	ctx := context.Background()

	t.Run("multipart reader", func(t *testing.T) {
		// Create a multipart form request
		body := strings.NewReader("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\n\r\ncontent\r\n--boundary--\r\n")
		req, err := NewRequestWithContext(ctx, "POST", "http://example.com", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

		reader, err := req.MultipartReader()
		if err != nil {
			t.Fatalf("unexpected error creating multipart reader: %v", err)
		}
		if reader == nil {
			t.Fatal("expected non-nil multipart reader")
		}
	})

	t.Run("parse multipart form", func(t *testing.T) {
		// Create a separate multipart form request
		body := strings.NewReader("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"test.txt\"\r\n\r\ncontent\r\n--boundary--\r\n")
		req, err := NewRequestWithContext(ctx, "POST", "http://example.com", body)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		req.Header.Set("Content-Type", "multipart/form-data; boundary=boundary")

		if err := req.ParseMultipartForm(1024); err != nil {
			t.Fatalf("unexpected error parsing multipart form: %v", err)
		}
	})
}

func TestRequestWithContext_Write(t *testing.T) {
	ctx := context.Background()
	req, err := NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Run("write to buffer", func(t *testing.T) {
		var buf strings.Builder
		if err := req.Write(&buf); err != nil {
			t.Fatalf("unexpected error writing request: %v", err)
		}
		written := buf.String()
		if !strings.Contains(written, "GET") {
			t.Error("expected written request to contain GET method")
		}
		if !strings.Contains(written, "example.com") {
			t.Error("expected written request to contain host")
		}
	})

	t.Run("write proxy", func(t *testing.T) {
		var buf strings.Builder
		if err := req.WriteProxy(&buf); err != nil {
			t.Fatalf("unexpected error writing proxy request: %v", err)
		}
		written := buf.String()
		if !strings.Contains(written, "GET") {
			t.Error("expected written proxy request to contain GET method")
		}
	})
}

func TestRequestWithContext_CustomContextTypes(t *testing.T) {
	t.Run("with cancel context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		req, err := NewRequestWithContext(ctx, "GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved := req.Context()
		if retrieved != ctx {
			t.Error("expected same context instance")
		}
	})

	t.Run("with value context", func(t *testing.T) {
		type testKey string
		const key testKey = "testkey"
		ctx := context.WithValue(context.Background(), key, "testvalue")

		req, err := NewRequestWithContext(ctx, "GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved := req.Context()
		if val := retrieved.Value(testKey("testkey")); val != "testvalue" {
			t.Errorf("expected context value 'testvalue', got %v", val)
		}
	})

	t.Run("with timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 0) // Already expired
		defer cancel()

		req, err := NewRequestWithContext(ctx, "GET", "http://example.com", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		retrieved := req.Context()
		select {
		case <-retrieved.Done():
			// Expected - context is already expired
		default:
			t.Error("expected context to be done")
		}
	})
}

func TestRequestWithContext_Integration(t *testing.T) {
	t.Run("http test server integration", func(t *testing.T) {
		// Create a test server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify we can access the request
			if r.Method != "POST" {
				t.Errorf("expected method POST, got %s", r.Method)
			}
			if r.URL.Path != "/test" {
				t.Errorf("expected path /test, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		// Create request with typed context
		type RequestContext struct {
			context.Context
			RequestID string
		}

		baseCtx := context.Background()
		ctx := RequestContext{Context: baseCtx, RequestID: "req-123"}

		req, err := NewRequestWithContext(ctx, "POST", server.URL+"/test", strings.NewReader("test"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify context access
		reqCtx := req.Context()
		customCtx, ok := reqCtx.(RequestContext)
		if !ok {
			t.Fatal("expected RequestContext type")
		}
		if customCtx.RequestID != "req-123" {
			t.Errorf("expected RequestID 'req-123', got %s", customCtx.RequestID)
		}

		// Make the request
		client := &http.Client{}
		resp, err := client.Do((*http.Request)(req))
		if err != nil {
			t.Fatalf("unexpected error making request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})
}

// Benchmark tests for performance
func BenchmarkRequestWithContext_NewRequestWithContext(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	}
}

func BenchmarkRequestWithContext_Context(b *testing.B) {
	ctx := context.Background()
	req, _ := NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.Context()
	}
}

func BenchmarkRequestWithContext_FormValue(b *testing.B) {
	ctx := context.Background()
	req, _ := NewRequestWithContext(ctx, "GET", "http://example.com?key=value", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = req.FormValue("key")
	}
}

func BenchmarkRequestWithContext_Cookie(b *testing.B) {
	ctx := context.Background()
	req, _ := NewRequestWithContext(ctx, "GET", "http://example.com", nil)
	req.AddCookie(&http.Cookie{Name: "test", Value: "value"})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = req.Cookie("test")
	}
}
