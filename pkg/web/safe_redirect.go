package web

import (
	"net/url"
	"strings"
	"sync"
)

var (
	redirectMu        sync.RWMutex
	redirectWhitelist []string
)

// SetRedirectWhitelist mutlak yönlendirmeler için izinli host adlarını ayarlar
// (ör. ["example.com","app.local"]). Boşsa yalnızca göreli yönlendirmelere izin verilir.
// Eşzamanlı kullanım için güvenlidir; girdi kopyalanır.
func SetRedirectWhitelist(hosts []string) {
	cp := make([]string, 0, len(hosts))
	for _, h := range hosts {
		h = strings.ToLower(strings.TrimSpace(h))
		if h != "" {
			cp = append(cp, h)
		}
	}
	redirectMu.Lock()
	redirectWhitelist = cp
	redirectMu.Unlock()
}

func whitelist() []string {
	redirectMu.RLock()
	defer redirectMu.RUnlock()
	return redirectWhitelist
}

func hasControlChar(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < 0x20 || s[i] == 0x7f {
			return true
		}
	}
	return false
}

// IsSafeRedirect next güvenli bir yönlendirme hedefiyse true döner.
//   - "/" ile başlayan göreli yollar güvenlidir; ancak "//host", "/\host" (tarayıcılar
//     bunları başka siteye giden şema-göreli URL olarak yorumlar) ve kontrol karakteri
//     (tab, CR, LF ...) içerenler reddedilir (WEB-1).
//   - Mutlak URL'ler yalnızca http/https şemasıyla ve host whitelist'te ise (tam veya
//     alt alan adı eşleşmesi) kabul edilir (WEB-2).
func IsSafeRedirect(next string) bool {
	next = strings.TrimSpace(next)
	if next == "" {
		return true
	}
	if hasControlChar(next) || strings.Contains(next, `\`) {
		return false
	}
	if strings.HasPrefix(next, "/") {
		return !strings.HasPrefix(next, "//")
	}
	u, err := url.Parse(next)
	if err != nil {
		return false
	}
	if !u.IsAbs() {
		// "foo/bar" veya "?x" gibi hedefler: yalnızca "/" ile başlayan yollara izin var.
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
	default:
		return false
	}
	if u.User != nil {
		return false
	}
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false
	}
	for _, allow := range whitelist() {
		if host == allow || strings.HasSuffix(host, "."+allow) {
			return true
		}
	}
	return false
}

// NormalizeSafeRedirect next güvenliyse (boşlukları kırpılmış hâlini), değilse fallback'i
// döner. fallback verilmezse "/" kullanılır; açıkça verilen fallback boş dizge olsa bile
// aynen döner (A5-4): NormalizeSafeRedirect(next, "") ile "next yok/güvensiz" durumu
// "next=/" durumundan ayırt edilebilir.
func NormalizeSafeRedirect(next string, fallback ...string) string {
	fb := "/"
	if len(fallback) > 0 {
		fb = fallback[0]
	}
	next = strings.TrimSpace(next)
	if next == "" {
		return fb
	}
	if IsSafeRedirect(next) {
		return next
	}
	return fb
}
