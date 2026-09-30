// Package httpx JSON odaklı, yeniden deneme, loglama ve doğrulama destekli bir
// HTTP istemci sarmalayıcısıdır.
package httpx

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/xeipuuv/gojsonschema"
)

// Varsayılanlar
const (
	DefaultTimeout               = 30 * time.Second
	DefaultResponseHeaderTimeout = 30 * time.Second
	DefaultMaxResponseBytes      = 10 << 20 // 10 MiB
	DefaultMaxAttempts           = 10       // her yolda geçerli mutlak deneme sınırı
	DefaultMaxRetryAfter         = 60 * time.Second
)

// ErrResponseTooLarge yanıt gövdesi MaxResponseBytes'ı aştığında döner.
var ErrResponseTooLarge = errors.New("httpx: response body too large")

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

	// Retry güvenliği
	// RetryNonIdempotent true ise POST/PATCH da yeniden denenir. false iken bu
	// metotlar yalnızca istekte Idempotency-Key başlığı varsa yeniden denenir.
	RetryNonIdempotent bool
	// MaxAttempts her yolda (RetryPolicy dahil) uygulanan mutlak deneme sınırı;
	// 0 ise max(DefaultMaxAttempts, RetryAttempts).
	MaxAttempts int
	// MaxRetryAfter: sunucunun Retry-After değeri bundan büyükse yeniden denenmez; 0 ise DefaultMaxRetryAfter.
	MaxRetryAfter time.Duration

	// MaxResponseBytes JSON yanıt gövdesi sınırı; 0 ise DefaultMaxResponseBytes, <0 ise sınırsız.
	MaxResponseBytes int64
	// StreamIdleTimeout > 0 ise stream/indirme gövdesinde bu süre boyunca veri
	// gelmezse istek iptal edilir.
	StreamIdleTimeout time.Duration

	// Log detayları
	LogHeaders             bool
	LogRequestBody         bool
	LogResponseBodyOnError bool
	// RedactHeaders / RedactBodyFields varsayılan listelere EK olarak maskelenir.
	RedactHeaders    map[string]struct{}
	RedactBodyFields map[string]struct{}
	MaxLogFileSize   int64

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

	logMu sync.Mutex
}

type Option func(*Client)

// Options
func WithTimeout(d time.Duration) Option        { return func(c *Client) { c.HC.Timeout = d } }
func WithBaseURL(u string) Option               { return func(c *Client) { c.BaseURL = strings.TrimRight(u, "/") } }
func WithHeader(k, v string) Option             { return func(c *Client) { c.Headers.Add(k, v) } }
func WithTransport(rt http.RoundTripper) Option { return func(c *Client) { c.HC.Transport = rt } }

// WithClient verilen istemcinin bir KOPYASINI kullanır; sonraki seçenekler
// (WithTimeout, WithTransport) çağıranın istemcisini değiştirmez.
func WithClient(hc *http.Client) Option {
	return func(c *Client) {
		if hc != nil {
			cp := *hc
			c.HC = &cp
		}
	}
}
func WithFileLogging(dir string, logSuccess, logError bool) Option {
	return func(c *Client) { c.LogDir, c.LogSuccess, c.LogError = dir, logSuccess, logError }
}

// WithRetry toplam deneme sayısı (attempts), sabit bekleme ve yeniden denenecek
// durum kodlarını ayarlar. Ağ hataları her zaman yeniden denenir (idempotent
// metotlarda). statuses verilmezse hiçbir HTTP durum kodu yeniden denenmez.
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

// WithRetryNonIdempotent POST/PATCH isteklerinin de yeniden denenmesine izin verir.
// Sunucu tarafı idempotent değilse çift kayıt oluşabilir.
func WithRetryNonIdempotent(allow bool) Option {
	return func(c *Client) { c.RetryNonIdempotent = allow }
}

// WithMaxAttempts mutlak deneme sınırını ayarlar (RetryPolicy dahil her yolda).
func WithMaxAttempts(n int) Option { return func(c *Client) { c.MaxAttempts = n } }

// WithMaxRetryAfter kabul edilecek en uzun Retry-After süresini ayarlar.
func WithMaxRetryAfter(d time.Duration) Option { return func(c *Client) { c.MaxRetryAfter = d } }

// WithMaxResponseBytes JSON yanıt gövdesi sınırını ayarlar (<0: sınırsız).
func WithMaxResponseBytes(n int64) Option { return func(c *Client) { c.MaxResponseBytes = n } }

// WithStreamIdleTimeout stream/indirme sırasında veri gelmeme süresi sınırı.
func WithStreamIdleTimeout(d time.Duration) Option {
	return func(c *Client) { c.StreamIdleTimeout = d }
}

func WithLoggingDetails(logHeaders, logReqBody, logRespBodyOnError bool) Option {
	return func(c *Client) {
		c.LogHeaders, c.LogRequestBody, c.LogResponseBodyOnError = logHeaders, logReqBody, logRespBodyOnError
	}
}

// WithRedactions varsayılan maskeleme listelerine ek başlık ve JSON alan adları ekler.
func WithRedactions(headerKeys []string, bodyFields []string) Option {
	return func(c *Client) {
		if len(headerKeys) > 0 {
			if c.RedactHeaders == nil {
				c.RedactHeaders = make(map[string]struct{}, len(headerKeys))
			}
			for _, k := range headerKeys {
				c.RedactHeaders[strings.ToLower(k)] = struct{}{}
			}
		}
		if len(bodyFields) > 0 {
			if c.RedactBodyFields == nil {
				c.RedactBodyFields = make(map[string]struct{}, len(bodyFields))
			}
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

var (
	defaultTransportOnce sync.Once
	defaultTransport     http.RoundTripper
)

// sharedTransport ResponseHeaderTimeout ayarlı paylaşımlı bir transport döner.
// Stream/indirme çağrıları toplam timeout'suz istemci kullandığından başlık
// beklemesi bu değerle sınırlanır.
func sharedTransport() http.RoundTripper {
	defaultTransportOnce.Do(func() {
		if t, ok := http.DefaultTransport.(*http.Transport); ok {
			cp := t.Clone()
			cp.ResponseHeaderTimeout = DefaultResponseHeaderTimeout
			defaultTransport = cp
			return
		}
		defaultTransport = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           (&net.Dialer{Timeout: 30 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
			ResponseHeaderTimeout: DefaultResponseHeaderTimeout,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
		}
	})
	return defaultTransport
}

// New yeni bir istemci oluşturur. Varsayılan: 30 sn toplam timeout (JSON
// çağrıları için), 30 sn yanıt başlığı timeout'u, 10 MiB yanıt sınırı.
func New(opts ...Option) *Client {
	c := &Client{
		HC:      &http.Client{Timeout: DefaultTimeout, Transport: sharedTransport()},
		Headers: make(http.Header),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
	if c.Headers == nil {
		c.Headers = make(http.Header)
	}
	return c
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
	if len(params) == 0 {
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

// HTTPError 4xx/5xx yanıtları temsil eder. Body en fazla 4 KiB içerir.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("http error: status=%d body=%s", e.StatusCode, e.Body)
}
func readSmallBody(body io.ReadCloser) (string, error) {
	b, err := io.ReadAll(io.LimitReader(body, 4<<10))
	return string(b), err
}

// baseHost BaseURL'nin host'unu (küçük harf) döner; BaseURL yoksa "".
func (c *Client) baseHost() string {
	if c.BaseURL == "" {
		return ""
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

// applyDefaultHeaders istemci düzeyindeki başlıkları isteğe ekler.
//   - İstekte zaten bulunan başlıklar ezilmez/çoğaltılmaz.
//   - BaseURL tanımlıyken farklı bir host'a giden isteklere kimlik başlıkları
//     (Authorization, Cookie, X-Api-Key ...) eklenmez.
func (c *Client) applyDefaultHeaders(req *http.Request) {
	if len(c.Headers) == 0 {
		return
	}
	crossHost := false
	if bh := c.baseHost(); bh != "" && req.URL != nil && !strings.EqualFold(req.URL.Host, bh) {
		crossHost = true
	}
	for k, vs := range c.Headers {
		if len(req.Header.Values(k)) > 0 {
			continue
		}
		if crossHost && isCredentialHeader(k) {
			continue
		}
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
}

// Do isteği istemci başlıkları ile gönderir (yeniden deneme yapmaz).
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return c.send(ctx, c.HC, req)
}

func (c *Client) send(ctx context.Context, hc *http.Client, req *http.Request) (*http.Response, error) {
	if ctx != nil {
		req = req.WithContext(ctx)
	}
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	c.applyDefaultHeaders(req)
	return hc.Do(req)
}

// buildJSONRequest istek nesnesini ve (maskelenmiş) gövde önizlemesini hazırlar.
func (c *Client) buildJSONRequest(ctx context.Context, method, path string, params url.Values, in any, headers http.Header) (*http.Request, string, error) {
	u, err := c.resolveURLWithParams(path, params)
	if err != nil {
		return nil, "", err
	}
	var body io.Reader
	var reqBodyPreview string
	if in != nil {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		if err := enc.Encode(in); err != nil {
			return nil, "", err
		}
		if c.LogRequestBody {
			reqBodyPreview = truncate(c.redactBodyPreview(buf.String()), 1000)
		}
		body = buf
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Accept", "application/json")
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	c.applyCorrelation(req)
	return req, reqBodyPreview, nil
}

func (c *Client) correlationHeaderName() string {
	if c.CorrelationHeader == "" {
		return "X-Correlation-ID"
	}
	return c.CorrelationHeader
}

func (c *Client) applyCorrelation(req *http.Request) {
	if !c.AutoCorrelation {
		return
	}
	h := c.correlationHeaderName()
	if req.Header.Get(h) == "" {
		req.Header.Set(h, newUUIDv4())
	}
}

// newUUIDv4 rastgele (v4) UUID üretir.
func newUUIDv4() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func (c *Client) maxResponseBytes() int64 {
	if c.MaxResponseBytes == 0 {
		return DefaultMaxResponseBytes
	}
	return c.MaxResponseBytes
}

// readLimited gövdeyi MaxResponseBytes sınırıyla okur.
func (c *Client) readLimited(r io.Reader) ([]byte, error) {
	limit := c.maxResponseBytes()
	if limit < 0 {
		return io.ReadAll(r)
	}
	b, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("%w: limit %d bytes", ErrResponseTooLarge, limit)
	}
	return b, nil
}

// exchangeJSON çekirdek akış: istek, yeniden deneme politikası, doğrulayıcı ve loglama.
func (c *Client) exchangeJSON(ctx context.Context, method, path string, params url.Values, headers http.Header, in, out any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	canRetry := c.retryAllowedFor(method, headers)
	maxAttempts := c.attemptCap()
	var lastErr error
	for attempt := 1; ; attempt++ {
		if err := ctx.Err(); err != nil {
			return joinCtxErr(err, lastErr)
		}
		start := time.Now()
		req, reqPrev, err := c.buildJSONRequest(ctx, method, path, params, in, headers)
		if err != nil {
			return err
		}
		resp, err := c.send(ctx, c.HC, req)
		if err != nil {
			c.logToFile(false, req, 0, start, err.Error(), reqPrev, "")
			c.logToSlog(false, req, 0, start, err.Error(), "")
			if ctx.Err() != nil {
				return joinCtxErr(ctx.Err(), err)
			}
			delay, ok := c.nextRetry(attempt, maxAttempts, canRetry, err, nil)
			if !ok {
				return err
			}
			lastErr = err
			if serr := sleepCtx(ctx, delay); serr != nil {
				return joinCtxErr(serr, lastErr)
			}
			continue
		}
		err = c.handleResponse(req, resp, start, reqPrev, path, out, attempt, maxAttempts, canRetry)
		var re *retryableError
		if !errors.As(err, &re) {
			return err
		}
		lastErr = re.cause
		if serr := sleepCtx(ctx, re.delay); serr != nil {
			return joinCtxErr(serr, lastErr)
		}
	}
}

// handleResponse yanıtı işler; yeniden denenebilir bir durumda *retryableError döner.
func (c *Client) handleResponse(req *http.Request, resp *http.Response, start time.Time, reqPrev, path string, out any, attempt, maxAttempts int, canRetry bool) error {
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := readSmallBody(resp.Body)
		// Bağlantının yeniden kullanılabilmesi için kalan gövdeyi sınırlı boşalt.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		var respPrev string
		if c.LogResponseBodyOnError {
			respPrev = c.redactBodyPreview(b)
		}
		// Hata önizlemesi de maskelenir (gövdede token vb. olabilir).
		errPrev := c.redactBodyPreview(b)
		c.logToFile(false, req, resp.StatusCode, start, errPrev, reqPrev, respPrev)
		c.logToSlog(false, req, resp.StatusCode, start, errPrev, respPrev)
		httpErr := &HTTPError{StatusCode: resp.StatusCode, Body: b}
		delay, ok := c.nextRetry(attempt, maxAttempts, canRetry, nil, resp)
		if !ok {
			return httpErr
		}
		if ra, has := parseRetryAfter(resp.Header.Get("Retry-After"), time.Now()); has {
			if ra > c.maxRetryAfter() {
				return httpErr
			}
			if ra > delay {
				delay = ra
			}
		}
		return &retryableError{delay: delay, cause: httpErr}
	}
	validator := c.pickValidator(path)
	if validator != nil || out != nil {
		body, err := c.readLimited(resp.Body)
		if err != nil {
			c.logToFile(false, req, resp.StatusCode, start, err.Error(), reqPrev, "")
			c.logToSlog(false, req, resp.StatusCode, start, err.Error(), "")
			return err
		}
		if validator != nil {
			if err := validator(resp, body); err != nil {
				return err
			}
		}
		if out != nil && len(bytes.TrimSpace(body)) > 0 {
			if err := json.Unmarshal(body, out); err != nil {
				return err
			}
		}
	} else {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	}
	c.logToFile(true, req, resp.StatusCode, start, "", reqPrev, "")
	c.logToSlog(true, req, resp.StatusCode, start, "", "")
	return nil
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

// IdempotencyKey Idempotency-Key başlığını ayarlar; bu başlık varken POST/PATCH
// istekleri de yeniden denenebilir.
func (h Hdr) IdempotencyKey(key string) Hdr {
	if key != "" {
		h.h.Set("Idempotency-Key", key)
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
	if ctx == nil {
		ctx = context.Background()
	}
	if d > 0 {
		return context.WithTimeout(ctx, d)
	}
	return ctx, func() {}
}

func (c *Client) GetJSONOpts(ctx context.Context, path string, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	defer cancel()
	return c.exchangeJSON(ctx2, http.MethodGet, path, opt.Params, opt.Headers, nil, out)
}
func (c *Client) PostJSONOpts(ctx context.Context, path string, in, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	defer cancel()
	return c.exchangeJSON(ctx2, http.MethodPost, path, opt.Params, opt.Headers, in, out)
}
func (c *Client) PutJSONOpts(ctx context.Context, path string, in, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	defer cancel()
	return c.exchangeJSON(ctx2, http.MethodPut, path, opt.Params, opt.Headers, in, out)
}
func (c *Client) PatchJSONOpts(ctx context.Context, path string, in, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	defer cancel()
	return c.exchangeJSON(ctx2, http.MethodPatch, path, opt.Params, opt.Headers, in, out)
}
func (c *Client) DeleteJSONOpts(ctx context.Context, path string, out any, opt RequestOptions) error {
	ctx2, cancel := withTimeoutOpt(ctx, opt.Timeout)
	defer cancel()
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
