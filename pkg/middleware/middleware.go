package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

// ---- RequestID ----

// RequestIDHeader istek kimliğinin okunduğu/yazıldığı başlık.
const RequestIDHeader = "X-Request-ID"

type requestIDKey struct{}

// RequestID her isteğe bir kimlik atar. İstemci geçerli bir X-Request-ID gönderdiyse
// (en fazla 128 karakter, yalnızca harf/rakam ve - _ . :) o kullanılır; aksi halde
// rastgele üretilir. Kimlik yanıt başlığına ve context'e yazılır (GetRequestID).
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if !validRequestID(id) {
			id = newRequestID()
		}
		w.Header().Set(RequestIDHeader, id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, id)))
	})
}

// GetRequestID context'teki istek kimliğini döner ("" olabilir).
func GetRequestID(ctx context.Context) string {
	s, _ := ctx.Value(requestIDKey{}).(string)
	return s
}

func validRequestID(s string) bool {
	if s == "" || len(s) > 128 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9':
		case c == '-' || c == '_' || c == '.' || c == ':':
		default:
			return false
		}
	}
	return true
}

func newRequestID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(b)
}

// ---- Logger ----

// Logger her isteği slog ile loglar: method, path, status, bytes, duration, remote, request_id.
// 5xx yanıtlar Error, 4xx Warn, diğerleri Info seviyesinde yazılır. l nil ise slog.Default().
func Logger(l *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := wrap(w)
			next.ServeHTTP(sw, r)
			logger := l
			if logger == nil {
				logger = slog.Default()
			}
			status := sw.statusCode()
			level := slog.LevelInfo
			switch {
			case status >= 500:
				level = slog.LevelError
			case status >= 400:
				level = slog.LevelWarn
			}
			logger.LogAttrs(r.Context(), level, "http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", status),
				slog.Int64("bytes", sw.bytes),
				slog.Duration("duration", time.Since(start)),
				slog.String("remote", r.RemoteAddr),
				slog.String("request_id", GetRequestID(r.Context())),
			)
		})
	}
}

// ---- Timeout ----

// Timeout handler'ı d süresiyle sınırlar (http.TimeoutHandler). Süre aşılırsa 503 döner
// ve r.Context() iptal edilir. Not: yanıt tamponlanır; Flush/Hijack (SSE, websocket)
// gerektiren rotalarda kullanmayın. d <= 0 ise ara katman etkisizdir.
func Timeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		if d <= 0 {
			return next
		}
		return http.TimeoutHandler(next, d, http.StatusText(http.StatusServiceUnavailable))
	}
}

// ---- BodyLimit ----

// BodyLimit istek gövdesini n baytla sınırlar. Content-Length n'i aşıyorsa handler
// çağrılmadan 413 döner; bildirilmeyen/yanlış uzunlukta gövdeler http.MaxBytesReader
// ile kesilir (handler okuma hatası alır). n <= 0 ise ara katman etkisizdir.
func BodyLimit(n int64) Middleware {
	return func(next http.Handler) http.Handler {
		if n <= 0 {
			return next
		}
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > n {
				http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
				return
			}
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, n)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ---- RealIP ----

// ParseTrustedProxies "10.0.0.0/8" veya "127.0.0.1" biçimindeki girdileri prefix listesine çevirir.
func ParseTrustedProxies(cidrs ...string) ([]netip.Prefix, error) {
	out := make([]netip.Prefix, 0, len(cidrs))
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		if strings.Contains(c, "/") {
			p, err := netip.ParsePrefix(c)
			if err != nil {
				return nil, err
			}
			out = append(out, p.Masked())
			continue
		}
		a, err := netip.ParseAddr(c)
		if err != nil {
			return nil, err
		}
		a = a.Unmap()
		out = append(out, netip.PrefixFrom(a, a.BitLen()))
	}
	return out, nil
}

// RealIP, istek güvenilir bir proxy'den geldiyse (RemoteAddr trusted listesinde)
// X-Forwarded-For (sağdan sola, güvenilir olmayan ilk adres) veya X-Real-IP başlığından
// istemci IP'sini çıkarıp r.RemoteAddr'a yazar. Güvenilir proxy listesi boşsa başlıklar
// ASLA dikkate alınmaz (IP sahteciliğine karşı).
func RealIP(trusted []netip.Prefix) Middleware {
	isTrusted := func(a netip.Addr) bool {
		a = a.Unmap()
		for _, p := range trusted {
			if p.Contains(a) {
				return true
			}
		}
		return false
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			remote, ok := parseAddr(r.RemoteAddr)
			if !ok || len(trusted) == 0 || !isTrusted(remote) {
				next.ServeHTTP(w, r)
				return
			}
			client := ""
			if xff := r.Header.Values("X-Forwarded-For"); len(xff) > 0 {
				parts := strings.Split(strings.Join(xff, ","), ",")
				for i := len(parts) - 1; i >= 0; i-- {
					a, ok := parseAddr(strings.TrimSpace(parts[i]))
					if !ok {
						break
					}
					client = a.String()
					if !isTrusted(a) {
						break
					}
				}
			}
			if client == "" {
				if a, ok := parseAddr(strings.TrimSpace(r.Header.Get("X-Real-IP"))); ok {
					client = a.String()
				}
			}
			if client != "" {
				r2 := r.Clone(r.Context())
				r2.RemoteAddr = client
				r = r2
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ClientIP r.RemoteAddr'dan IP'yi (port olmadan) döner.
func ClientIP(r *http.Request) string {
	if a, ok := parseAddr(r.RemoteAddr); ok {
		return a.String()
	}
	return r.RemoteAddr
}

func parseAddr(s string) (netip.Addr, bool) {
	if s == "" {
		return netip.Addr{}, false
	}
	if a, err := netip.ParseAddr(s); err == nil {
		return a.Unmap(), true
	}
	if host, _, err := net.SplitHostPort(s); err == nil {
		if a, err := netip.ParseAddr(host); err == nil {
			return a.Unmap(), true
		}
	}
	return netip.Addr{}, false
}

// ---- CORS ----

// CORSOptions CORS yapılandırması.
type CORSOptions struct {
	// AllowedOrigins izin verilen origin'ler (örn. "https://app.example.com"). "*" tüm
	// origin'lere izin verir ANCAK AllowCredentials true iken "*" yok sayılır (güvenlik).
	AllowedOrigins []string
	// AllowedMethods boşsa GET, POST, PUT, PATCH, DELETE, HEAD.
	AllowedMethods []string
	// AllowedHeaders boşsa Content-Type, Authorization, X-CSRF-Token, X-Requested-With.
	AllowedHeaders []string
	// ExposedHeaders tarayıcıya açılacak yanıt başlıkları.
	ExposedHeaders []string
	// AllowCredentials Access-Control-Allow-Credentials: true gönderir.
	AllowCredentials bool
	// MaxAge preflight önbellek süresi (saniye); 0 ise gönderilmez.
	MaxAge int
}

// CORS verilen seçeneklerle CORS başlıklarını uygular. İzin verilmeyen origin'lere
// CORS başlığı eklenmez (tarayıcı isteği engeller). Preflight (OPTIONS +
// Access-Control-Request-Method) istekleri 204 ile burada sonlandırılır.
func CORS(opts CORSOptions) Middleware {
	methods := opts.AllowedMethods
	if len(methods) == 0 {
		methods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodHead}
	}
	headers := opts.AllowedHeaders
	if len(headers) == 0 {
		headers = []string{"Content-Type", "Authorization", "X-CSRF-Token", "X-Requested-With"}
	}
	allowAll := false
	origins := map[string]struct{}{}
	for _, o := range opts.AllowedOrigins {
		o = strings.TrimSpace(o)
		if o == "*" {
			if !opts.AllowCredentials {
				allowAll = true
			}
			continue
		}
		if o != "" {
			origins[strings.ToLower(strings.TrimRight(o, "/"))] = struct{}{}
		}
	}
	allowMethods := strings.Join(methods, ", ")
	allowHeaders := strings.Join(headers, ", ")
	expose := strings.Join(opts.ExposedHeaders, ", ")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			h := w.Header()
			h.Add("Vary", "Origin")
			preflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
			if preflight {
				h.Add("Vary", "Access-Control-Request-Method")
				h.Add("Vary", "Access-Control-Request-Headers")
			}
			allowed := false
			if origin != "" {
				if allowAll {
					allowed = true
				} else {
					_, allowed = origins[strings.ToLower(origin)]
				}
			}
			if !allowed {
				if preflight {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			if allowAll {
				h.Set("Access-Control-Allow-Origin", "*")
			} else {
				h.Set("Access-Control-Allow-Origin", origin)
			}
			if opts.AllowCredentials {
				h.Set("Access-Control-Allow-Credentials", "true")
			}
			if preflight {
				h.Set("Access-Control-Allow-Methods", allowMethods)
				h.Set("Access-Control-Allow-Headers", allowHeaders)
				if opts.MaxAge > 0 {
					h.Set("Access-Control-Max-Age", strconv.Itoa(opts.MaxAge))
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}
			if expose != "" {
				h.Set("Access-Control-Expose-Headers", expose)
			}
			next.ServeHTTP(w, r)
		})
	}
}
