package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xeipuuv/gojsonschema"

	"github.com/mustafacaglarkara/webdev/pkg/helpers"
)

// Client yapılandırması
type Client struct {
	HC      *http.Client
	BaseURL string
	Headers http.Header

	// Dosyaya log
	LogDir     string
	LogSuccess bool
	LogError   bool

	// Sabit retry
	RetryAttempts int
	RetryDelay    time.Duration
	RetryStatuses map[int]struct{}

	// Exponential backoff
	BackoffInitial    time.Duration
	BackoffMax        time.Duration
	BackoffMultiplier float64
	BackoffJitter     float64

	// Log detayları
	LogHeaders             bool
	LogRequestBody         bool
	LogResponseBodyOnError bool
	RedactHeaders          map[string]struct{}
	RedactBodyFields       map[string]struct{}
	MaxLogFileSize         int64

	// Correlation
	CorrelationHeader string
	AutoCorrelation   bool

	// Policy + Logger + Validators
	RetryPolicy              RetryPolicy
	Logger                   *slog.Logger
	LogToSlogSuccess         bool
	LogToSlogError           bool
	ResponseValidators       map[string]ResponseValidator
	DefaultResponseValidator ResponseValidator
}

type Option func(*Client)

// Options
func WithTimeout(d time.Duration) Option        { return func(c *Client) { c.HC.Timeout = d } }
func WithBaseURL(u string) Option               { return func(c *Client) { c.BaseURL = strings.TrimRight(u, "/") } }
func WithHeader(k, v string) Option             { return func(c *Client) { c.Headers.Add(k, v) } }
func WithTransport(rt http.RoundTripper) Option { return func(c *Client) { c.HC.Transport = rt } }
func WithClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			c.HC = hc
		}
	}
}
func WithFileLogging(dir string, logSuccess, logError bool) Option {
	return func(c *Client) { c.LogDir, c.LogSuccess, c.LogError = dir, logSuccess, logError }
}
func WithRetry(attempts int, delay time.Duration, statuses ...int) Option {
	return func(c *Client) {
		if attempts < 1 {
			attempts = 1
		}
		c.RetryAttempts, c.RetryDelay = attempts, delay
		if len(statuses) > 0 {
			c.RetryStatuses = make(map[int]struct{}, len(statuses))
			for _, s := range statuses {
				c.RetryStatuses[s] = struct{}{}
			}
		}
	}
}
func WithExponentialBackoff(initial, max time.Duration, mult, jitter float64) Option {
	return func(c *Client) {
		c.BackoffInitial, c.BackoffMax, c.BackoffMultiplier, c.BackoffJitter = initial, max, mult, clamp01(jitter)
	}
}
func WithLoggingDetails(logHeaders, logReqBody, logRespBodyOnError bool) Option {
	return func(c *Client) {
		c.LogHeaders, c.LogRequestBody, c.LogResponseBodyOnError = logHeaders, logReqBody, logRespBodyOnError
	}
}
func WithRedactions(headerKeys []string, bodyFields []string) Option {
	return func(c *Client) {
		if len(headerKeys) > 0 {
			c.RedactHeaders = make(map[string]struct{}, len(headerKeys))
			for _, k := range headerKeys {
				c.RedactHeaders[strings.ToLower(k)] = struct{}{}
			}
		}
		if len(bodyFields) > 0 {
			c.RedactBodyFields = make(map[string]struct{}, len(bodyFields))
			for _, k := range bodyFields {
				c.RedactBodyFields[strings.ToLower(k)] = struct{}{}
			}
		}
	}
}
func WithLogRotation(maxBytes int64) Option { return func(c *Client) { c.MaxLogFileSize = maxBytes } }
func WithCorrelationHeader(name string, auto bool) Option {
	return func(c *Client) { c.CorrelationHeader, c.AutoCorrelation = name, auto }
}
func WithSlogger(l *slog.Logger, logSuccess, logError bool) Option {
	return func(c *Client) { c.Logger, c.LogToSlogSuccess, c.LogToSlogError = l, logSuccess, logError }
}
func WithRetryPolicy(p RetryPolicy) Option { return func(c *Client) { c.RetryPolicy = p } }
func WithResponseValidator(v ResponseValidator) Option {
	return func(c *Client) { c.DefaultResponseValidator = v }
}
func WithResponseValidatorForPath(path string, v ResponseValidator) Option {
	return func(c *Client) {
		if c.ResponseValidators == nil {
			c.ResponseValidators = map[string]ResponseValidator{}
		}
		c.ResponseValidators[path] = v
	}
}

// New
func New(opts ...Option) *Client {
	c := &Client{HC: &http.Client{Timeout: 30 * time.Second}, Headers: make(http.Header)}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Retry Policy
type RetryPolicy interface {
	Next(attempt int, err error, resp *http.Response) (time.Duration, bool)
}

type FixedPolicy struct {
	Delay         time.Duration
	MaxAttempts   int
	RetryStatuses map[int]struct{}
}

func (p FixedPolicy) Next(attempt int, err error, resp *http.Response) (time.Duration, bool) {
	if attempt >= p.MaxAttempts {
		return 0, false
	}
	if err != nil {
		return p.Delay, true
	}
	if resp != nil {
		if _, ok := p.RetryStatuses[resp.StatusCode]; ok {
			return p.Delay, true
		}
	}
	return 0, false
}

type JitterKind int

const (
	JitterNone JitterKind = iota
	JitterFull            // [0, delay]
)

type ExponentialPolicy struct {
	Initial, Max  time.Duration
	Multiplier    float64
	Jitter        JitterKind
	MaxAttempts   int
	RetryStatuses map[int]struct{}
}

func (p ExponentialPolicy) backoff(attempt int) time.Duration {
	if attempt <= 1 {
		return p.Initial
	}
	d := float64(p.Initial)
	for i := 1; i < attempt; i++ {
		d *= p.Multiplier
	}
	dur := time.Duration(d)
	if p.Max > 0 && dur > p.Max {
		dur = p.Max
	}
	if p.Jitter == JitterFull && dur > 0 {
		return time.Duration(rand.Int63n(dur.Nanoseconds() + 1))
	}
	return dur
}
func (p ExponentialPolicy) Next(attempt int, err error, resp *http.Response) (time.Duration, bool) {
	if attempt >= p.MaxAttempts {
		return 0, false
	}
	if err != nil {
		return p.backoff(attempt), true
	}
	if resp != nil {
		if _, ok := p.RetryStatuses[resp.StatusCode]; ok {
			return p.backoff(attempt), true
		}
	}
	return 0, false
}

// Response validation
type ResponseValidator func(resp *http.Response, body []byte) error

func NewJSONSchemaValidatorFromString(schema string) ResponseValidator {
	loader := gojsonschema.NewStringLoader(schema)
	return func(resp *http.Response, body []byte) error {
		doc := gojsonschema.NewBytesLoader(body)
		res, err := gojsonschema.Validate(loader, doc)
		if err != nil {
			return err
		}
		if !res.Valid() {
			var sb strings.Builder
			sb.WriteString("schema validation failed: ")
			for _, e := range res.Errors() {
				sb.WriteString(e.String())
				sb.WriteString("; ")
			}
			return errors.New(sb.String())
		}
		return nil
	}
}

// Helpers (URL, HTTP)
func (c *Client) cloneHeaders() http.Header {
	cp := make(http.Header, len(c.Headers))
	for k, vs := range c.Headers {
		for _, v := range vs {
			cp.Add(k, v)
		}
	}
	return cp
}
func (c *Client) resolveURL(path string) (string, error) {
	if path == "" {
		return c.BaseURL, nil
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}
	if c.BaseURL == "" {
		return path, nil
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(path, "/") {
		u.Path = strings.TrimRight(u.Path, "/") + path
	} else {
		if strings.HasSuffix(u.Path, "/") || u.Path == "" {
			u.Path = u.Path + path
		} else {
			u.Path = u.Path + "/" + path
		}
	}
	return u.String(), nil
}
func (c *Client) resolveURLWithParams(path string, params url.Values) (string, error) {
	uStr, err := c.resolveURL(path)
	if err != nil {
		return "", err
	}
	if params == nil || len(params) == 0 {
		return uStr, nil
	}
	u, err := url.Parse(uStr)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, vs := range params {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: status=%d body=%s", e.StatusCode, e.Body)
}
func readSmallBody(body io.ReadCloser) (string, error) {
	defer body.Close()
	b, err := io.ReadAll(io.LimitReader(body, 4<<10))
	return string(b), err
}

func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	if len(c.Headers) > 0 {
		for k, vs := range c.Headers {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
	}
	return c.HC.Do(req)
}

func (c *Client) doJSONWith(ctx context.Context, method, path string, params url.Values, in any, headers http.Header) (*http.Request, *http.Response, string, error) {
	u, err := c.resolveURLWithParams(path, params)
	if err != nil {
		return nil, nil, "", err
	}
	var body io.Reader
	var reqBodyPreview string
	if in != nil {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(in); err != nil {
			return nil, nil, "", err
		}
		reqBodyPreview = buf.String()
		if c.LogRequestBody {
			reqBodyPreview = c.redactBodyPreview(reqBodyPreview)
			if len(reqBodyPreview) > 1000 {
				reqBodyPreview = reqBodyPreview[:1000]
			}
		} else {
			reqBodyPreview = ""
		}
		body = buf
	}
	req, err := http.NewRequest(method, u, body)
	if err != nil {
		return nil, nil, "", err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if headers != nil {
		for k, vs := range headers {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
	}
	// Correlation
	if c.AutoCorrelation {
		h := c.CorrelationHeader
		if h == "" {
			h = "X-Correlation-ID"
		}
		if req.Header.Get(h) == "" {
			req.Header.Set(h, helpers.MustUUIDv4())
		}
	}
	resp, err := c.Do(ctx, req)
	return req, resp, reqBodyPreview, err
}

// Core exchange (policy + validator + logging)
func (c *Client) exchangeJSON(ctx context.Context, method, path string, params url.Values, headers http.Header, in, out any) error {
	start := time.Now()
	attempt := 0
	for {
		attempt++
		req, resp, reqPrev, err := c.doJSONWith(ctx, method, path, params, in, headers)
		if err != nil {
			c.logToFile(false, req, 0, start, err.Error(), reqPrev, "")
			c.logToSlog(false, req, 0, start, "", "")
			if c.shouldRetryWithPolicy(attempt, err, nil) {
				if err2 := c.sleepBackoffWithPolicy(ctx, attempt, err, nil); err2 != nil {
					return err
				}
				continue
			}
			return err
		}
		func() {
			defer resp.Body.Close()
			if resp.StatusCode >= 400 {
				b, _ := readSmallBody(resp.Body)
				var respPrev string
				if c.LogResponseBodyOnError {
					respPrev = c.redactBodyPreview(b)
				}
				c.logToFile(false, req, resp.StatusCode, start, b, reqPrev, respPrev)
				c.logToSlog(false, req, resp.StatusCode, start, b, respPrev)
				if c.shouldRetryWithPolicy(attempt, nil, resp) {
					if err2 := c.sleepBackoffWithPolicy(ctx, attempt, nil, resp); err2 != nil {
						err = err2
					} else {
						err = errors.New("retryable status")
					}
					return
				}
				err = &HTTPError{StatusCode: resp.StatusCode, Body: b}
				return
			}
			// Success body
			validator := c.pickValidator(path)
			if validator != nil || out != nil {
				body, rerr := io.ReadAll(resp.Body)
				if rerr != nil {
					err = rerr
					return
				}
				if validator != nil {
					if verr := validator(resp, body); verr != nil {
						err = verr
						return
					}
				}
				if out != nil {
					if derr := json.Unmarshal(body, &out); derr != nil {
						err = derr
						return
					}
				}
			} else {
				if out == nil {
					io.Copy(io.Discard, resp.Body)
				} else {
					dec := json.NewDecoder(resp.Body)
					if derr := dec.Decode(out); derr != nil {
						err = derr
						return
					}
				}
			}
			c.logToFile(true, req, resp.StatusCode, start, "", reqPrev, "")
			c.logToSlog(true, req, resp.StatusCode, start, "", "")
			err = nil
		}()
		if err == nil || !strings.Contains(err.Error(), "retryable") {
			return err
		}
		if err2 := c.sleepBackoffWithPolicy(ctx, attempt, nil, nil); err2 != nil {
			return err
		}
	}
}

// Validators
func (c *Client) pickValidator(path string) ResponseValidator {
	if c.ResponseValidators != nil {
		if v, ok := c.ResponseValidators[path]; ok {
			return v
		}
	}
	return c.DefaultResponseValidator
}

// Retry decision
func (c *Client) shouldRetryWithPolicy(attempt int, err error, resp *http.Response) bool {
	if c.RetryPolicy != nil {
		_, retry := c.RetryPolicy.Next(attempt, err, resp)
		return retry
	}
	return c.shouldRetry(err, statusOf(resp), attempt)
}
func statusOf(resp *http.Response) int {
	if resp == nil {
		return 0
	}
	return resp.StatusCode
}

// Sleep backoff according to policy or legacy config
func (c *Client) sleepBackoffWithPolicy(ctx context.Context, attempt int, err error, resp *http.Response) error {
	if c.RetryPolicy != nil {
		d, retry := c.RetryPolicy.Next(attempt, err, resp)
		if !retry {
			return nil
		}
		t := time.NewTimer(d)
		defer t.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-t.C:
			return nil
		}
	}
	return c.sleepBackoff(ctx, attempt)
}
func (c *Client) shouldRetry(netErr error, status int, attempt int) bool {
	if c.RetryAttempts <= 0 {
		return false
	}
	if attempt >= c.RetryAttempts {
		return false
	}
	if netErr != nil {
		return true
	}
	if status > 0 && c.shouldRetryStatus(status) {
		return true
	}
	return false
}
func (c *Client) shouldRetryStatus(status int) bool {
	if c.RetryStatuses == nil {
		return false
	}
	_, ok := c.RetryStatuses[status]
	return ok
}
func (c *Client) sleepBackoff(ctx context.Context, attempt int) error {
	var d time.Duration
	if c.BackoffInitial > 0 && c.BackoffMultiplier > 0 {
		d = c.BackoffInitial
		for i := 1; i < attempt; i++ {
			d = time.Duration(float64(d) * c.BackoffMultiplier)
			if c.BackoffMax > 0 && d > c.BackoffMax {
				d = c.BackoffMax
				break
			}
		}
		if c.BackoffJitter > 0 {
			nanos := float64(d.Nanoseconds())
			delta := nanos * clamp01(c.BackoffJitter)
			if attempt%2 == 0 {
				nanos += delta
			} else {
				nanos -= delta
			}
			if nanos < float64(time.Millisecond) {
				nanos = float64(time.Millisecond)
			}
			d = time.Duration(nanos)
		}
	} else if c.RetryDelay > 0 {
		d = c.RetryDelay
	} else {
		d = 100 * time.Millisecond
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// Log helpers
func (c *Client) logToSlog(success bool, req *http.Request, status int, start time.Time, errPreview string, bodyPreview string) {
	if c.Logger == nil {
		return
	}
	if success && !c.LogToSlogSuccess {
		return
	}
	if !success && !c.LogToSlogError {
		return
	}
	urlStr, host := "", ""
	if req != nil && req.URL != nil {
		urlStr = req.URL.String()
		host = req.URL.Host
	}
	dur := time.Since(start)
	h := c.CorrelationHeader
	if h == "" {
		h = "X-Correlation-ID"
	}
	cid := ""
	if req != nil {
		cid = req.Header.Get(h)
	}
	attrs := []any{"method", methodSafe(req), "url", urlStr, "status", status, "dur", dur.String(), "host", host}
	if cid != "" {
		attrs = append(attrs, "cid", cid)
	}
	if !success && errPreview != "" {
		attrs = append(attrs, "err", truncate(errPreview, 200))
	}
	if !success && bodyPreview != "" {
		attrs = append(attrs, "respb", truncate(bodyPreview, 200))
	}
	if success {
		c.Logger.Info("httpx", attrs...)
	} else {
		c.Logger.Warn("httpx", attrs...)
	}
}
func (c *Client) logToFile(success bool, req *http.Request, status int, start time.Time, errPreview, reqBodyPrev, respBodyPrev string) {
	if c.LogDir == "" {
		return
	}
	if success && !c.LogSuccess {
		return
	}
	if !success && !c.LogError {
		return
	}
	sub := "error"
	if success {
		sub = "success"
	}
	dir := filepath.Join(c.LogDir, sub)
	_ = helpers.EnsureDir(dir)
	fname := time.Now().Format("2006-01-02") + ".txt"
	p := filepath.Join(dir, fname)
	urlStr, host := "", ""
	if req != nil && req.URL != nil {
		urlStr = req.URL.String()
		host = req.URL.Host
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	dur := time.Since(start).Truncate(time.Millisecond)
	line := fmt.Sprintf("%s | %s %s | status=%d | host=%s | dur=%s", ts, methodSafe(req), urlStr, status, host, dur)
	h := c.CorrelationHeader
	if h == "" {
		h = "X-Correlation-ID"
	}
	if req != nil {
		if cid := req.Header.Get(h); cid != "" {
			line += " | cid=" + cid
		}
	}
	if c.LogHeaders {
		if rh := c.redactHeaders(req.Header); rh != nil {
			b, _ := json.Marshal(rh)
			hs := string(b)
			if len(hs) > 1000 {
				hs = hs[:1000]
			}
			line += " | reqh=" + hs
		}
	}
	if c.LogRequestBody && reqBodyPrev != "" {
		line += " | reqb=" + strings.ReplaceAll(reqBodyPrev, "\n", " ")
	}
	if !success && c.LogResponseBodyOnError && respBodyPrev != "" {
		line += " | respb=" + strings.ReplaceAll(respBodyPrev, "\n", " ")
	}
	if !success && errPreview != "" {
		if len(errPreview) > 1000 {
			errPreview = errPreview[:1000]
		}
		line += " | err=" + strings.ReplaceAll(errPreview, "\n", " ")
	}
	// rotation
	if c.MaxLogFileSize > 0 {
		if fi, err := os.Stat(p); err == nil {
			projected := fi.Size() + int64(len(line)) + 1
			if projected > c.MaxLogFileSize {
				_ = os.Rename(p, p+"."+time.Now().Format("20060102_150405"))
			}
		}
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line + "\n")
}

func (c *Client) redactHeaders(h http.Header) http.Header {
	if !c.LogHeaders || h == nil {
		return nil
	}
	if len(c.RedactHeaders) == 0 {
		return h.Clone()
	}
	cp := make(http.Header, len(h))
	for k, vs := range h {
		if _, ok := c.RedactHeaders[strings.ToLower(k)]; ok {
			cp[k] = []string{"***"}
		} else {
			cp[k] = append([]string(nil), vs...)
		}
	}
	return cp
}
func (c *Client) redactBodyPreview(s string) string {
	if s == "" {
		return s
	}
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return s
	}
	red := c.redactJSON(v)
	b, err := json.Marshal(red)
	if err != nil {
		return s
	}
	return string(b)
}
func (c *Client) redactJSON(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			if _, ok := c.RedactBodyFields[strings.ToLower(k)]; ok {
				out[k] = "***"
			} else {
				out[k] = c.redactJSON(val)
			}
		}
		return out
	case []any:
		for i := range t {
			t[i] = c.redactJSON(t[i])
		}
		return t
	default:
		return v
	}
}

// Builders & Options
type RequestOptions struct {
	Params  url.Values
	Headers http.Header
	Timeout time.Duration
}

type Query struct{ v url.Values }

func Q() Query { return Query{v: make(url.Values)} }
func (q Query) Add(k string, vals ...string) Query {
	for _, v := range vals {
		q.v.Add(k, v)
	}
	return q
}
func (q Query) Set(k, v string) Query { q.v.Set(k, v); return q }
func (q Query) Del(k string) Query    { q.v.Del(k); return q }
func (q Query) Values() url.Values    { return q.v }
func (q Query) Encode() string        { return q.v.Encode() }

type Hdr struct{ h http.Header }

func H() Hdr                      { return Hdr{h: make(http.Header)} }
func (h Hdr) Add(k, v string) Hdr { h.h.Add(k, v); return h }
func (h Hdr) Set(k, v string) Hdr { h.h.Set(k, v); return h }
func (h Hdr) Bearer(token string) Hdr {
	if token != "" {
		h.h.Set("Authorization", "Bearer "+token)
	}
	return h
}
func (h Hdr) Values() http.Header { return h.h }

func QM(m map[string]string) Query {
	q := Q()
	for k, v := range m {
		q = q.Set(k, v)
	}
	return q
}
func HM(m map[string]string) Hdr {
	h := H()
	for k, v := range m {
		h = h.Set(k, v)
	}
	return h
}

// JSON shortcuts
func (c *Client) GetJSON(ctx context.Context, path string, out any) error {
	return c.exchangeJSON(ctx, http.MethodGet, path, nil, nil, nil, out)
}
func (c *Client) PostJSON(ctx context.Context, path string, in, out any) error {
	return c.exchangeJSON(ctx, http.MethodPost, path, nil, nil, in, out)
}
func (c *Client) PutJSON(ctx context.Context, path string, in, out any) error {
	return c.exchangeJSON(ctx, http.MethodPut, path, nil, nil, in, out)
}
func (c *Client) PatchJSON(ctx context.Context, path string, in, out any) error {
	return c.exchangeJSON(ctx, http.MethodPatch, path, nil, nil, in, out)
}
func (c *Client) DeleteJSON(ctx context.Context, path string, out any) error {
	return c.exchangeJSON(ctx, http.MethodDelete, path, nil, nil, nil, out)
}

func (c *Client) GetJSONWith(ctx context.Context, path string, params url.Values, headers http.Header, out any) error {
	return c.exchangeJSON(ctx, http.MethodGet, path, params, headers, nil, out)
}
func (c *Client) PostJSONWith(ctx context.Context, path string, params url.Values, headers http.Header, in, out any) error {
	return c.exchangeJSON(ctx, http.MethodPost, path, params, headers, in, out)
}
func (c *Client) PutJSONWith(ctx context.Context, path string, params url.Values, headers http.Header, in, out any) error {
	return c.exchangeJSON(ctx, http.MethodPut, path, params, headers, in, out)
}
func (c *Client) PatchJSONWith(ctx context.Context, path string, params url.Values, headers http.Header, in, out any) error {
	return c.exchangeJSON(ctx, http.MethodPatch, path, params, headers, in, out)
}
func (c *Client) DeleteJSONWith(ctx context.Context, path string, params url.Values, headers http.Header, out any) error {
	return c.exchangeJSON(ctx, http.MethodDelete, path, params, headers, nil, out)
}

func withTimeoutOpt(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d > 0 {
		return context.WithTimeout(ctx, d)
	}
	return ctx, nil
}

func (c *Client) GetJSONOpts(ctx context.Context, path string, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	if cancel != nil {
		defer cancel()
	}
	return c.exchangeJSON(ctx2, http.MethodGet, path, opt.Params, opt.Headers, nil, out)
}
func (c *Client) PostJSONOpts(ctx context.Context, path string, in, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	if cancel != nil {
		defer cancel()
	}
	return c.exchangeJSON(ctx2, http.MethodPost, path, opt.Params, opt.Headers, in, out)
}
func (c *Client) PutJSONOpts(ctx context.Context, path string, in, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	if cancel != nil {
		defer cancel()
	}
	return c.exchangeJSON(ctx2, http.MethodPut, path, opt.Params, opt.Headers, in, out)
}
func (c *Client) PatchJSONOpts(ctx context.Context, path string, in, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	if cancel != nil {
		defer cancel()
	}
	return c.exchangeJSON(ctx2, http.MethodPatch, path, opt.Params, opt.Headers, in, out)
}
func (c *Client) DeleteJSONOpts(ctx context.Context, path string, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	if cancel != nil {
		defer cancel()
	}
	return c.exchangeJSON(ctx2, http.MethodDelete, path, opt.Params, opt.Headers, nil, out)
}

func (c *Client) GetJSONQ(ctx context.Context, path string, q Query, out any) error {
	return c.GetJSONWith(ctx, path, q.Values(), nil, out)
}
func (c *Client) PostJSONQ(ctx context.Context, path string, q Query, in, out any) error {
	return c.PostJSONWith(ctx, path, q.Values(), nil, in, out)
}
func (c *Client) PutJSONQ(ctx context.Context, path string, q Query, in, out any) error {
	return c.PutJSONWith(ctx, path, q.Values(), nil, in, out)
}
func (c *Client) PatchJSONQ(ctx context.Context, path string, q Query, in, out any) error {
	return c.PatchJSONWith(ctx, path, q.Values(), nil, in, out)
}
func (c *Client) DeleteJSONQ(ctx context.Context, path string, q Query, out any) error {
	return c.DeleteJSONWith(ctx, path, q.Values(), nil, out)
}

func (c *Client) GetJSONQH(ctx context.Context, path string, q Query, h Hdr, out any) error {
	return c.GetJSONWith(ctx, path, q.Values(), h.Values(), out)
}
func (c *Client) PostJSONQH(ctx context.Context, path string, q Query, h Hdr, in, out any) error {
	return c.PostJSONWith(ctx, path, q.Values(), h.Values(), in, out)
}
func (c *Client) PutJSONQH(ctx context.Context, path string, q Query, h Hdr, in, out any) error {
	return c.PutJSONWith(ctx, path, q.Values(), h.Values(), in, out)
}
func (c *Client) PatchJSONQH(ctx context.Context, path string, q Query, h Hdr, in, out any) error {
	return c.PatchJSONWith(ctx, path, q.Values(), h.Values(), in, out)
}
func (c *Client) DeleteJSONQH(ctx context.Context, path string, q Query, h Hdr, out any) error {
	return c.DeleteJSONWith(ctx, path, q.Values(), h.Values(), out)
}

// Streaming helpers
func (c *Client) StreamNDJSON(ctx context.Context, path string, params url.Values, headers http.Header, handle func(json.RawMessage) error) error {
	req, resp, _, err := c.doJSONWith(ctx, http.MethodGet, path, params, nil, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := readSmallBody(resp.Body)
		c.logToFile(false, req, resp.StatusCode, time.Now(), b, "", "")
		return &HTTPError{StatusCode: resp.StatusCode, Body: b}
	}
	dec := json.NewDecoder(resp.Body)
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		if handle != nil {
			if err := handle(raw); err != nil {
				return err
			}
		}
	}
}
func (c *Client) DownloadToFile(ctx context.Context, path string, params url.Values, headers http.Header, dst string) error {
	req, resp, _, err := c.doJSONWith(ctx, http.MethodGet, path, params, nil, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := readSmallBody(resp.Body)
		c.logToFile(false, req, resp.StatusCode, time.Now(), b, "", "")
		return &HTTPError{StatusCode: resp.StatusCode, Body: b}
	}
	if err := helpers.EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}
func (c *Client) UploadStream(ctx context.Context, method, path string, params url.Values, headers http.Header, r io.Reader, contentType string, out any) error {
	u, err := c.resolveURLWithParams(path, params)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, u, r)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if headers != nil {
		for k, vs := range headers {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
	}
	if c.AutoCorrelation {
		h := c.CorrelationHeader
		if h == "" {
			h = "X-Correlation-ID"
		}
		if req.Header.Get(h) == "" {
			req.Header.Set(h, helpers.MustUUIDv4())
		}
	}
	resp, err := c.Do(ctx, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := readSmallBody(resp.Body)
		return &HTTPError{StatusCode: resp.StatusCode, Body: b}
	}
	if out == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
}

// misc
func methodSafe(req *http.Request) string {
	if req == nil {
		return ""
	}
	return req.Method
}
func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
func clamp01(f float64) float64 {
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}
