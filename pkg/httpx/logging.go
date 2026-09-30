package httpx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// defaultRedactHeaders her zaman maskelenen başlıklar (küçük harf).
var defaultRedactHeaders = map[string]struct{}{
	"authorization":        {},
	"proxy-authorization":  {},
	"cookie":               {},
	"set-cookie":           {},
	"x-api-key":            {},
	"api-key":              {},
	"x-auth-token":         {},
	"x-access-token":       {},
	"x-csrf-token":         {},
	"x-xsrf-token":         {},
	"x-amz-security-token": {},
}

// credentialHeaders farklı host'lara varsayılan olarak gönderilmeyen başlıklar.
var credentialHeaders = map[string]struct{}{
	"authorization":       {},
	"proxy-authorization": {},
	"cookie":              {},
	"x-api-key":           {},
	"api-key":             {},
	"x-auth-token":        {},
	"x-access-token":      {},
}

func isCredentialHeader(k string) bool {
	_, ok := credentialHeaders[strings.ToLower(k)]
	return ok
}

// defaultRedactBodyFields JSON gövdelerinde her zaman maskelenen alanlar.
var defaultRedactBodyFields = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"secret":        {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"id_token":      {},
	"client_secret": {},
	"api_key":       {},
	"apikey":        {},
}

// secretQueryExact tam eşleşen gizli sorgu parametreleri (küçük harf).
var secretQueryExact = map[string]struct{}{
	"key": {}, "sig": {}, "code": {}, "auth": {}, "pwd": {}, "pass": {},
	"session": {}, "sessionid": {}, "jwt": {}, "authorization": {},
}

// secretQuerySubstr alt dize olarak geçtiğinde gizli sayılan parametre parçaları.
var secretQuerySubstr = []string{"token", "secret", "password", "passwd", "signature", "api_key", "apikey", "api-key", "credential", "private"}

func isSecretQueryParam(k string) bool {
	l := strings.ToLower(k)
	if _, ok := secretQueryExact[l]; ok {
		return true
	}
	for _, s := range secretQuerySubstr {
		if strings.Contains(l, s) {
			return true
		}
	}
	// x-amz-credential, x-goog-signature ... yukarıda yakalanır; *_key biçimi:
	return strings.HasSuffix(l, "_key") || strings.HasSuffix(l, "-key")
}

// redactURL loglanacak URL'de kullanıcı parolasını ve gizli sorgu
// parametrelerini "***" ile değiştirir.
func redactURL(u *url.URL) string {
	if u == nil {
		return ""
	}
	cp := *u
	if cp.User != nil {
		if _, has := cp.User.Password(); has {
			cp.User = url.UserPassword(cp.User.Username(), "***")
		}
	}
	if cp.RawQuery != "" {
		q, err := url.ParseQuery(cp.RawQuery)
		if err != nil {
			cp.RawQuery = "***"
		} else {
			changed := false
			for k, vs := range q {
				if isSecretQueryParam(k) {
					for i := range vs {
						vs[i] = "***"
					}
					changed = true
				}
			}
			if changed {
				cp.RawQuery = q.Encode()
			}
		}
	}
	return cp.String()
}

func (c *Client) isRedactedHeader(k string) bool {
	l := strings.ToLower(k)
	if _, ok := defaultRedactHeaders[l]; ok {
		return true
	}
	_, ok := c.RedactHeaders[l]
	return ok
}

func (c *Client) isRedactedField(k string) bool {
	l := strings.ToLower(k)
	if _, ok := defaultRedactBodyFields[l]; ok {
		return true
	}
	_, ok := c.RedactBodyFields[l]
	return ok
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
		urlStr = redactURL(req.URL)
		host = req.URL.Host
	}
	dur := time.Since(start)
	cid := ""
	if req != nil {
		cid = req.Header.Get(c.correlationHeaderName())
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
	urlStr, host := "", ""
	if req != nil && req.URL != nil {
		urlStr = redactURL(req.URL)
		host = req.URL.Host
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	dur := time.Since(start).Truncate(time.Millisecond)
	line := fmt.Sprintf("%s | %s %s | status=%d | host=%s | dur=%s", ts, methodSafe(req), urlStr, status, host, dur)
	if req != nil {
		if cid := req.Header.Get(c.correlationHeaderName()); cid != "" {
			line += " | cid=" + cid
		}
	}
	if c.LogHeaders && req != nil {
		if rh := c.redactHeaders(req.Header); rh != nil {
			b, _ := json.Marshal(rh)
			line += " | reqh=" + truncate(string(b), 1000)
		}
	}
	if c.LogRequestBody && reqBodyPrev != "" {
		line += " | reqb=" + strings.ReplaceAll(reqBodyPrev, "\n", " ")
	}
	if !success && c.LogResponseBodyOnError && respBodyPrev != "" {
		line += " | respb=" + strings.ReplaceAll(respBodyPrev, "\n", " ")
	}
	if !success && errPreview != "" {
		line += " | err=" + strings.ReplaceAll(truncate(errPreview, 1000), "\n", " ")
	}

	// Yazma ve rotasyon aynı kilit altında: eşzamanlı istekler satırları
	// karıştırmaz, rotasyon yarışmaz.
	c.logMu.Lock()
	defer c.logMu.Unlock()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return
	}
	p := filepath.Join(dir, time.Now().Format("2006-01-02")+".txt")
	if c.MaxLogFileSize > 0 {
		if fi, err := os.Stat(p); err == nil {
			projected := fi.Size() + int64(len(line)) + 1
			if projected > c.MaxLogFileSize {
				_ = os.Rename(p, p+"."+time.Now().Format("20060102_150405.000000000"))
			}
		}
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(line + "\n")
}

// redactHeaders loglama için başlıkların maskelenmiş kopyasını döner.
// Varsayılan hassas başlıklar (Authorization, Cookie, Set-Cookie, X-Api-Key ...)
// her zaman maskelenir.
func (c *Client) redactHeaders(h http.Header) http.Header {
	if !c.LogHeaders || h == nil {
		return nil
	}
	cp := make(http.Header, len(h))
	for k, vs := range h {
		if c.isRedactedHeader(k) {
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
			if c.isRedactedField(k) {
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
