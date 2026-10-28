package coworkingview

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Version is this client's version, sent in the User-Agent header.
const Version = "0.1.0"

// DefaultBaseURL is the origin of the public CoworkingView API.
const DefaultBaseURL = "https://api.coworkingview.com"

// DefaultTimeout bounds every request made by a client built without
// [WithHTTPClient]. An unreachable API must never wedge a caller.
const DefaultTimeout = 15 * time.Second

// Client reads the CoworkingView catalogue. The zero value is not usable; build
// one with [NewClient]. A Client is safe for concurrent use.
type Client struct {
	baseURL    string
	httpClient *http.Client
	userAgent  string
	locale     Locale
}

// Option configures a [Client].
type Option func(*Client)

// WithBaseURL points the client at another origin, such as a local gateway.
func WithBaseURL(base string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(base, "/") }
}

// WithHTTPClient replaces the default *http.Client (and its timeout).
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) { c.httpClient = h }
}

// WithUserAgent prefixes the User-Agent header, so the API's operators can tell
// your integration apart, e.g. "acme-office-finder/2.1".
func WithUserAgent(ua string) Option {
	return func(c *Client) { c.userAgent = ua + " " + c.userAgent }
}

// WithLocale sets the language used when a call's params leave Locale empty.
// The default is [English].
func WithLocale(l Locale) Option {
	return func(c *Client) { c.locale = l }
}

// NewClient returns a client for the public API at [DefaultBaseURL].
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: DefaultTimeout},
		userAgent:  "coworkingview-go/" + Version + " (+https://coworkingview.com)",
		locale:     English,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// APIError is a non-2xx answer from the API. The API speaks RFC 9457
// problem+json, so Title and Detail are human-readable and Code is stable.
type APIError struct {
	StatusCode int          `json:"status"`
	Type       string       `json:"type"`
	Title      string       `json:"title"`
	Detail     string       `json:"detail,omitempty"`
	Code       string       `json:"code,omitempty"`
	Errors     []FieldError `json:"errors,omitempty"`
}

// FieldError names one rejected query parameter in a validation failure.
type FieldError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	msg := e.Detail
	if msg == "" {
		msg = e.Title
	}
	for _, fe := range e.Errors {
		msg += fmt.Sprintf("; %s: %s", fe.Path, fe.Message)
	}
	return fmt.Sprintf("coworkingview: %d %s", e.StatusCode, msg)
}

// IsNotFound reports whether err is the API saying the slug does not exist (404)
// or no longer exists (410, e.g. a space that closed).
func IsNotFound(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && (apiErr.StatusCode == http.StatusNotFound || apiErr.StatusCode == http.StatusGone)
}

func (c *Client) localeOr(l Locale) string {
	if l != "" {
		return string(l)
	}
	return string(c.locale)
}

// get calls path with query and decodes the {data, meta} envelope into out.
func (c *Client) get(ctx context.Context, path string, query url.Values, out any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("coworkingview: request failed: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 32<<20))
	if err != nil {
		return fmt.Errorf("coworkingview: reading response: %w", err)
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		apiErr := &APIError{StatusCode: res.StatusCode, Title: http.StatusText(res.StatusCode)}
		_ = json.Unmarshal(body, apiErr)
		apiErr.StatusCode = res.StatusCode
		return apiErr
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("coworkingview: decoding %s: %w", path, err)
	}
	return nil
}

// Response is the envelope every read returns: the payload and its metadata.
type Response[T any] struct {
	Data T    `json:"data"`
	Meta Meta `json:"meta"`
}

// Health reports whether the API is up and which build is serving it.
func (c *Client) Health(ctx context.Context) (*Response[HealthStatus], error) {
	var out Response[HealthStatus]
	if err := c.get(ctx, "/v1/health", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// HealthStatus is the payload of [Client.Health].
type HealthStatus struct {
	Status string `json:"status"`
	SHA    string `json:"sha"`
}
