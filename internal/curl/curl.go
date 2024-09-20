package curl

import (
	"context"
	"faast-go/internal/config"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

type CurlConfig struct {
	ValidateType string
	SizeDefault  int
	CodeDefault  int
	Cookies      []http.Cookie
	URL          string
	RateLimiter  *rate.Limiter
	UserAgent    string
	Client       *http.Client
	Fields       []string
	StaticValues []string
}

func NewCurlConfig(config *config.YamlConfig) (*CurlConfig, error) {
	cookies := make([]http.Cookie, 0, len(config.Cookies))
	for _, cookieStr := range config.Cookies {
		parts := strings.SplitN(cookieStr, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid cookie format: %s", cookieStr)
		}
		cookies = append(cookies, http.Cookie{Name: parts[0], Value: parts[1]})
	}

	var rateLimiter *rate.Limiter
	if config.RateLimit > 0 {
		rateLimiter = rate.NewLimiter(rate.Limit(config.RateLimit), int(config.RateLimit))
	}

	return &CurlConfig{
		ValidateType: config.ValidateType,
		SizeDefault:  config.SizeDefault,
		CodeDefault:  config.CodeDefault,
		Cookies:      cookies,
		URL:          config.Endpoint,
		RateLimiter:  rateLimiter,
		// this will be a variable in the future
		UserAgent:    "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
		Client:       &http.Client{Timeout: time.Duration(config.Timeout) * time.Second},
		Fields:       config.Fields,
		StaticValues: config.StaticValues,
	}, nil
}

func (c *CurlConfig) ValidateResponse(res *http.Response) bool {
	switch c.ValidateType {
	case "size":
		return res.ContentLength == int64(c.SizeDefault)
	case "code":
		return res.StatusCode == c.CodeDefault
	default:
		fmt.Printf("Warning: invalid validate type '%s'. Defaulting to true.\n", c.ValidateType)
		return true
	}
}

func (c *CurlConfig) SendCurl(ctx context.Context, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", c.URL, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	for _, cookie := range c.Cookies {
		req.AddCookie(&cookie)
	}
	req.Header.Set("User-Agent", c.UserAgent)

	if c.RateLimiter != nil {
		if err := c.RateLimiter.Wait(ctx); err != nil {
			return nil, fmt.Errorf("rate limiter wait cancelled: %w", err)
		}
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error from response: %w", err)
	}

	return resp, nil
}

func (c *CurlConfig) ConstructPayload(permutation []string) (*strings.Reader, error) {
	if len(c.Fields) != (len(permutation) + len(c.StaticValues)) {
		return nil, fmt.Errorf("error: length of permutation and values are not equal")
	}

	var payload strings.Builder
	for i, field := range c.Fields {
		if i > 0 {
			payload.WriteString("&")
		}
		if i < len(permutation) {
			payload.WriteString(url.QueryEscape(field) + "=" + url.QueryEscape(permutation[i]))
		} else {
			payload.WriteString(url.QueryEscape(field) + "=" + url.QueryEscape(c.StaticValues[i-len(permutation)]))
		}
	}

	return strings.NewReader(payload.String()), nil
}
