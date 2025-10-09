package web

import (
	"net/url"
	"strings"
)

var redirectWhitelist []string

// SetRedirectWhitelist sets allowed hostnames for absolute redirects (e.g., ["example.com","app.local"]).
// If empty, only relative redirects are allowed.
func SetRedirectWhitelist(hosts []string) {
	redirectWhitelist = hosts
}

// IsSafeRedirect returns true if the provided next is a safe redirect target.
// - Relative paths are considered safe (starting with '/').
// - Absolute URLs are allowed only if host matches whitelist entries (case-insensitive match or suffix match).
func IsSafeRedirect(next string) bool {
	next = strings.TrimSpace(next)
	if next == "" {
		return true
	}
	// relative
	if strings.HasPrefix(next, "/") {
		return true
	}
	u, err := url.Parse(next)
	if err != nil {
		return false
	}
	if !u.IsAbs() {
		return strings.HasPrefix(u.Path, "/")
	}
	if len(redirectWhitelist) == 0 {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, allow := range redirectWhitelist {
		allow = strings.ToLower(strings.TrimSpace(allow))
		if allow == "" {
			continue
		}
		if host == allow || strings.HasSuffix(host, "."+allow) {
			return true
		}
	}
	return false
}

// NormalizeSafeRedirect returns next if safe; otherwise returns fallback (default "/").
func NormalizeSafeRedirect(next string, fallback ...string) string {
	fb := "/"
	if len(fallback) > 0 && fallback[0] != "" {
		fb = fallback[0]
	}
	if next == "" {
		return fb
	}
	if IsSafeRedirect(next) {
		return next
	}
	return fb
}
