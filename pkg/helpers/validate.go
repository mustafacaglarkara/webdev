package helpers

import (
	"net/mail"
	"net/url"
)

// --- Doğrulamalar ---
func IsEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}

func IsURL(s string) bool {
	u, err := url.ParseRequestURI(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func NotEmpty(s string) bool { return s != "" }
