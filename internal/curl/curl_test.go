package curl

import (
	"context"
	"faast-go/internal/config"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewCurlConfig(t *testing.T) {
	yamlConfig := &config.YamlConfig{
		ValidateType: "code",
		SizeDefault:  100,
		CodeDefault:  404,
		Cookies:      []string{"session=abc123", "user=john"},
		Endpoint:     "http://example.com",
		RateLimit:    10,
		Timeout:      5,
		Fields:       []string{"field1", "field2"},
		StaticValues: []string{"static1"},
	}

	curlConfig, err := NewCurlConfig(yamlConfig)
	if err != nil {
		t.Fatalf("NewCurlConfig failed: %v", err)
	}

	// Check if the CurlConfig fields are set correctly
	if curlConfig.ValidateType != yamlConfig.ValidateType {
		t.Errorf("ValidateType mismatch. Got %s, want %s", curlConfig.ValidateType, yamlConfig.ValidateType)
	}
	if curlConfig.SizeDefault != yamlConfig.SizeDefault {
		t.Errorf("SizeDefault mismatch. Got %d, want %d", curlConfig.SizeDefault, yamlConfig.SizeDefault)
	}
	if curlConfig.CodeDefault != yamlConfig.CodeDefault {
		t.Errorf("CodeDefault mismatch. Got %d, want %d", curlConfig.CodeDefault, yamlConfig.CodeDefault)
	}
	if len(curlConfig.Cookies) != len(yamlConfig.Cookies) {
		t.Errorf("Cookies length mismatch. Got %d, want %d", len(curlConfig.Cookies), len(yamlConfig.Cookies))
	}
	if curlConfig.URL != yamlConfig.Endpoint {
		t.Errorf("URL mismatch. Got %s, want %s", curlConfig.URL, yamlConfig.Endpoint)
	}
	if curlConfig.Client.Timeout != time.Duration(yamlConfig.Timeout)*time.Second {
		t.Errorf("Client Timeout mismatch. Got %v, want %v", curlConfig.Client.Timeout, time.Duration(yamlConfig.Timeout)*time.Second)
	}

	// Test invalid cookie format
	yamlConfig.Cookies = []string{"invalid_cookie"}
	_, err = NewCurlConfig(yamlConfig)
	if err == nil {
		t.Error("NewCurlConfig should have returned an error for invalid cookie format")
	}
}

func TestValidateResponse(t *testing.T) {
	tests := []struct {
		name         string
		validateType string
		sizeDefault  int
		codeDefault  int
		response     *http.Response
		want         bool
	}{
		{
			name:         "Validate Size Success",
			validateType: "size",
			sizeDefault:  100,
			response:     &http.Response{ContentLength: 100},
			want:         true,
		},
		{
			name:         "Validate Size Failure",
			validateType: "size",
			sizeDefault:  100,
			response:     &http.Response{ContentLength: 200},
			want:         false,
		},
		{
			name:         "Validate Code Success",
			validateType: "code",
			codeDefault:  404,
			response:     &http.Response{StatusCode: 404},
			want:         true,
		},
		{
			name:         "Validate Code Failure",
			validateType: "code",
			codeDefault:  404,
			response:     &http.Response{StatusCode: 200},
			want:         false,
		},
		{
			name:         "Invalid Validate Type",
			validateType: "invalid",
			response:     &http.Response{},
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CurlConfig{
				ValidateType: tt.validateType,
				SizeDefault:  tt.sizeDefault,
				CodeDefault:  tt.codeDefault,
			}
			if got := c.ValidateResponse(tt.response); got != tt.want {
				t.Errorf("ValidateResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendCurl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the User-Agent is set correctly
		if r.Header.Get("User-Agent") != "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36" {
			t.Errorf("User-Agent not set correctly")
		}

		// Check if the cookie is set correctly
		cookie, err := r.Cookie("session")
		if err != nil || cookie.Value != "abc123" {
			t.Errorf("Cookie not set correctly")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test response"))
	}))
	defer server.Close()

	c := &CurlConfig{
		URL:    server.URL,
		Client: &http.Client{},
		Cookies: []http.Cookie{
			{Name: "session", Value: "abc123"},
		},
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
	}

	resp, err := c.SendCurl(context.Background(), strings.NewReader("test body"))
	if err != nil {
		t.Fatalf("SendCurl failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code. Got %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if string(body) != "Test response" {
		t.Errorf("Unexpected response body. Got %s, want %s", string(body), "Test response")
	}
}

func TestConstructPayload(t *testing.T) {
	tests := []struct {
		name         string
		fields       []string
		staticValues []string
		permutation  []string
		want         string
		wantErr      bool
	}{
		{
			name:         "Valid Payload",
			fields:       []string{"field1", "field2", "field3"},
			staticValues: []string{"static1"},
			permutation:  []string{"value1", "value2"},
			want:         "field1=value1&field2=value2&field3=static1",
			wantErr:      false,
		},
		{
			name:         "Mismatched Lengths",
			fields:       []string{"field1", "field2"},
			staticValues: []string{"static1"},
			permutation:  []string{"value1", "value2"},
			want:         "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CurlConfig{
				Fields:       tt.fields,
				StaticValues: tt.staticValues,
			}
			got, err := c.ConstructPayload(tt.permutation)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConstructPayload() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				gotString, _ := io.ReadAll(got)
				if string(gotString) != tt.want {
					t.Errorf("ConstructPayload() = %v, want %v", string(gotString), tt.want)
				}
			}
		})
	}
}
