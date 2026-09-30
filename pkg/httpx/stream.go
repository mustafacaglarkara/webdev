package httpx

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// streamClient toplam süre sınırı (http.Client.Timeout) OLMAYAN bir istemci
// döner; aynı transport, yönlendirme politikası ve cookie jar kullanılır.
// Uzun süren akış/indirme/yükleme çağrıları istemci timeout'u ile kesilmez;
// süre kontrolü ctx, transport'un ResponseHeaderTimeout değeri ve
// StreamIdleTimeout ile yapılır.
func (c *Client) streamClient() *http.Client {
	return &http.Client{
		Transport:     c.HC.Transport,
		CheckRedirect: c.HC.CheckRedirect,
		Jar:           c.HC.Jar,
	}
}

// idleReader her Read'de zamanlayıcıyı yeniler; süre dolarsa cancel çağrılır.
type idleReader struct {
	r     io.Reader
	timer *time.Timer
	d     time.Duration
}

func (ir *idleReader) Read(p []byte) (int, error) {
	n, err := ir.r.Read(p)
	if n > 0 {
		ir.timer.Reset(ir.d)
	}
	return n, err
}

// errStreamIdle akışta StreamIdleTimeout boyunca veri gelmediğinde döner.
var errStreamIdle = errors.New("httpx: stream idle timeout")

// streamCtx isteğe özel iptal edilebilir context ve (varsa) boşta kalma
// zamanlayıcısı hazırlar.
type streamCtx struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	timer  *time.Timer
	once   sync.Once
}

func (c *Client) newStreamCtx(parent context.Context) *streamCtx {
	if parent == nil {
		parent = context.Background()
	}
	ctx, cancel := context.WithCancelCause(parent)
	s := &streamCtx{ctx: ctx, cancel: cancel}
	if c.StreamIdleTimeout > 0 {
		s.timer = time.AfterFunc(c.StreamIdleTimeout, func() { cancel(errStreamIdle) })
	}
	return s
}

func (s *streamCtx) wrap(r io.Reader, d time.Duration) io.Reader {
	if s.timer == nil {
		return r
	}
	s.timer.Reset(d)
	return &idleReader{r: r, timer: s.timer, d: d}
}

func (s *streamCtx) close() {
	s.once.Do(func() {
		if s.timer != nil {
			s.timer.Stop()
		}
		s.cancel(nil)
	})
}

// err ctx iptal nedeni idle timeout ise onu döner.
func (s *streamCtx) err(orig error) error {
	if cause := context.Cause(s.ctx); errors.Is(cause, errStreamIdle) {
		return errors.Join(errStreamIdle, orig)
	}
	return orig
}

// openStream GET isteği gönderir ve 2xx yanıtı döner (yeniden deneme yapmaz).
func (c *Client) openStream(sc *streamCtx, path string, params url.Values, headers http.Header) (*http.Response, error) {
	start := time.Now()
	req, _, err := c.buildJSONRequest(sc.ctx, http.MethodGet, path, params, nil, headers)
	if err != nil {
		return nil, err
	}
	resp, err := c.send(sc.ctx, c.streamClient(), req)
	if err != nil {
		c.logToFile(false, req, 0, start, err.Error(), "", "")
		c.logToSlog(false, req, 0, start, err.Error(), "")
		return nil, sc.err(err)
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		b, _ := readSmallBody(resp.Body)
		c.logToFile(false, req, resp.StatusCode, start, c.redactBodyPreview(b), "", "")
		c.logToSlog(false, req, resp.StatusCode, start, c.redactBodyPreview(b), "")
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: b}
	}
	return resp, nil
}

// StreamNDJSON satır satır JSON (NDJSON) akışını okur ve her kayıt için handle'ı çağırır.
// İstemcinin toplam timeout'u uygulanmaz; iptal için ctx, boşta kalma için
// WithStreamIdleTimeout kullanın. Yeniden deneme yapılmaz.
func (c *Client) StreamNDJSON(ctx context.Context, path string, params url.Values, headers http.Header, handle func(json.RawMessage) error) error {
	sc := c.newStreamCtx(ctx)
	defer sc.close()
	resp, err := c.openStream(sc, path, params, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(sc.wrap(resp.Body, c.StreamIdleTimeout))
	for {
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return sc.err(err)
		}
		if handle != nil {
			if err := handle(raw); err != nil {
				return err
			}
		}
	}
}

// DownloadToFile yanıt gövdesini dst dosyasına yazar. İçerik önce aynı dizinde
// geçici bir dosyaya yazılır ve yalnızca başarıyla tamamlanırsa dst adına
// taşınır; hata/iptal durumunda yarım dosya kalmaz ve mevcut dst korunur.
// İstemcinin toplam timeout'u uygulanmaz (bkz. StreamNDJSON).
func (c *Client) DownloadToFile(ctx context.Context, path string, params url.Values, headers http.Header, dst string) (err error) {
	sc := c.newStreamCtx(ctx)
	defer sc.close()
	resp, err := c.openStream(sc, path, params, headers)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, "."+filepath.Base(dst)+".*.part")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() {
		if err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
		}
	}()
	if _, err = io.Copy(tmp, sc.wrap(resp.Body, c.StreamIdleTimeout)); err != nil {
		return sc.err(err)
	}
	if err = tmp.Sync(); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmpName, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpName, dst)
}

// UploadStream r içeriğini akış olarak gönderir. Gövde tekrar okunamayacağından
// yeniden deneme yapılmaz ve istemcinin toplam timeout'u uygulanmaz; süre
// kontrolü için ctx kullanın. JSON yanıt MaxResponseBytes ile sınırlıdır.
func (c *Client) UploadStream(ctx context.Context, method, path string, params url.Values, headers http.Header, r io.Reader, contentType string, out any) error {
	if ctx == nil {
		ctx = context.Background()
	}
	u, err := c.resolveURLWithParams(path, params)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, u, r)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for k, vs := range headers {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}
	c.applyCorrelation(req)
	start := time.Now()
	resp, err := c.send(ctx, c.streamClient(), req)
	if err != nil {
		c.logToFile(false, req, 0, start, err.Error(), "", "")
		c.logToSlog(false, req, 0, start, err.Error(), "")
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := readSmallBody(resp.Body)
		c.logToFile(false, req, resp.StatusCode, start, c.redactBodyPreview(b), "", "")
		c.logToSlog(false, req, resp.StatusCode, start, c.redactBodyPreview(b), "")
		return &HTTPError{StatusCode: resp.StatusCode, Body: b}
	}
	c.logToFile(true, req, resp.StatusCode, start, "", "", "")
	c.logToSlog(true, req, resp.StatusCode, start, "", "")
	if out == nil {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
		return nil
	}
	body, err := c.readLimited(resp.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, out)
}
