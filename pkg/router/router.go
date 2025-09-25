package router

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

var (
	tmplMu    sync.RWMutex
	templates = make(map[string]string)
	// regex to find {name} tokens (after normalization)
	tokenRe = regexp.MustCompile(`\{([^/}]+)\}`)
	// regex to capture {name:...} or {name} for normalization
	normRe = regexp.MustCompile(`\{([^/}:]+)(?:[^}]*)\}`)
)

// registerTemplate stores the route pattern template for reverse lookups.
func registerTemplate(name, pattern string) {
	tmplMu.Lock()
	defer tmplMu.Unlock()
	templates[name] = pattern
}

// RegisterTemplate exposes registerTemplate to other packages.
func RegisterTemplate(name, pattern string) {
	registerTemplate(name, pattern)
}

// RegisterRoute stores a route template for reverse URL lookups. Same as RegisterTemplate.
func RegisterRoute(name, pattern string) {
	registerTemplate(name, pattern)
}

// GetTemplate returns the registered template for a route name.
func GetTemplate(name string) (string, bool) {
	tmplMu.RLock()
	defer tmplMu.RUnlock()
	p, ok := templates[name]
	return p, ok
}

// normalizePattern strips param regexes from {name:regex} -> {name}
func normalizePattern(pattern string) string {
	return normRe.ReplaceAllString(pattern, `{$1}`)
}

// ReverseURL builds a URL from a named route template.
// params supports either:
// - a single map[string]string or map[string]any for named substitutions
// - positional parameters (fmt.Sprintf style or to fill {name} tokens in order)
// Examples:
// ReverseURL("user.show", map[string]string{"id":"123"}) => /user/123
// ReverseURL("user.show", "123") => /user/123 (if template is /user/{id})
// ReverseURL("search", "golang") => fmt.Sprintf on printf-like templates
func ReverseURL(name string, params ...any) (string, error) {
	tmplMu.RLock()
	pattern, ok := templates[name]
	tmplMu.RUnlock()
	if !ok {
		return "", fmt.Errorf("route template not found: %s", name)
	}

	// If pattern contains printf verbs (%s, %d etc.) and positional params provided, use fmt.Sprintf
	if strings.Contains(pattern, "%") && len(params) > 0 {
		// escape each arg for path safety
		esc := make([]any, len(params))
		for i := range params {
			esc[i] = url.PathEscape(fmt.Sprint(params[i]))
		}
		// guard fmt.Sprintf panics by recovering
		defer func() { _ = recover() }()
		return fmt.Sprintf(pattern, esc...), nil
	}

	// Normalize pattern to strip token regex parts (e.g. {id:[0-9]+} -> {id})
	norm := normalizePattern(pattern)

	// Find brace tokens {name}
	tokens := tokenRe.FindAllStringSubmatch(norm, -1)
	if len(tokens) == 0 {
		// no tokens, return pattern as-is
		return pattern, nil
	}

	// If single param and it's a map => named substitution
	if len(params) == 1 {
		switch m := params[0].(type) {
		case map[string]string:
			return substituteNamedEscaped(norm, m)
		case map[string]any:
			// convert to string map
			sm := make(map[string]string, len(m))
			for k, v := range m {
				sm[k] = fmt.Sprint(v)
			}
			return substituteNamedEscaped(norm, sm)
		}
	}

	// Positional substitution: replace tokens in order with params (escaped)
	if len(params) < len(tokens) {
		return "", fmt.Errorf("not enough params for route %s: need %d got %d", name, len(tokens), len(params))
	}
	out := norm
	for i, tk := range tokens {
		val := url.PathEscape(fmt.Sprint(params[i]))
		out = strings.Replace(out, tk[0], val, 1)
	}
	return out, nil
}

func substituteNamedEscaped(pattern string, m map[string]string) (string, error) {
	out := pattern
	// For each {name} token, replace from map
	matches := tokenRe.FindAllStringSubmatch(pattern, -1)
	for _, mm := range matches {
		whole := mm[0]
		key := mm[1]
		v, ok := m[key]
		if !ok {
			return "", fmt.Errorf("missing param '%s' for route", key)
		}
		out = strings.Replace(out, whole, url.PathEscape(v), 1)
	}
	return out, nil
}

// ReverseURLMust is like ReverseURL but panics on error (convenience)
func ReverseURLMust(name string, params ...any) string {
	s, err := ReverseURL(name, params...)
	if err != nil {
		panic(err)
	}
	return s
}

// ReverseURLWithQuery: ReverseURL + query ekleme (query map[string]any; slice değerler çoklu param olur)
func ReverseURLWithQuery(name string, params []any, query map[string]any) (string, error) {
	base, err := ReverseURL(name, params...)
	if err != nil {
		return "", err
	}
	if len(query) == 0 {
		return base, nil
	}
	vals := url.Values{}
	for k, v := range query {
		if v == nil {
			continue
		}
		switch t := v.(type) {
		case []string:
			for _, s := range t {
				vals.Add(k, s)
			}
		case []any:
			for _, iv := range t {
				vals.Add(k, fmt.Sprint(iv))
			}
		default:
			vals.Add(k, fmt.Sprint(t))
		}
	}
	enc := vals.Encode()
	if enc == "" {
		return base, nil
	}
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	return base + sep + enc, nil
}

// Convenience error values
var (
	ErrRouteNotFound = errors.New("route not found")
)
