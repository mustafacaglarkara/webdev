package web

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const oldSessionKey = "old_form"

var (
	OldInputsMaxJSONSize      = 2048 // max JSON length before truncation strategies
	OldInputsTruncatePerValue = 256  // per value truncate length in strategy 1
	OldInputsTruncateFirstVal = 128  // first value truncate length in strategy 2
)

// SetOldInputLimits allows overriding global limits (use with caution in init/test code)
func SetOldInputLimits(maxJSON, perValue, firstValue int) {
	if maxJSON > 0 {
		OldInputsMaxJSONSize = maxJSON
	}
	if perValue > 0 {
		OldInputsTruncatePerValue = perValue
	}
	if firstValue > 0 {
		OldInputsTruncateFirstVal = firstValue
	}
}

// SetOldInputs stores ALL submitted form values (after size safeguards) into the flash session.
// It serializes url.Values as JSON. Size is bounded by truncation logic below to keep cookie < ~4KB.
func SetOldInputs(w http.ResponseWriter, r *http.Request, form url.Values) error {
	s := getStore()
	sess, err := s.Get(r, flashSessionName)
	if err != nil {
		return err
	}
	b, err := json.Marshal(form)
	if err != nil {
		return err
	}
	// Prevent extremely large session cookies: if JSON too big, truncate values
	maxSize := OldInputsMaxJSONSize
	if len(b) > maxSize {
		// Strategy 1: per-value truncation
		reduced := url.Values{}
		for k, vals := range form {
			for _, v := range vals {
				if len(v) > OldInputsTruncatePerValue {
					v = v[:OldInputsTruncatePerValue]
				}
				reduced.Add(k, v)
			}
		}
		b2, err2 := json.Marshal(reduced)
		if err2 == nil && len(b2) <= maxSize {
			sess.Values[oldSessionKey] = string(b2)
			return sess.Save(r, w)
		}
		// Strategy 2: first value only
		mini := url.Values{}
		for k, vals := range form {
			v := ""
			if len(vals) > 0 {
				v = vals[0]
				if len(v) > OldInputsTruncateFirstVal {
					v = v[:OldInputsTruncateFirstVal]
				}
			}
			mini.Set(k, v)
		}
		b3, _ := json.Marshal(mini)
		sess.Values[oldSessionKey] = string(b3)
		return sess.Save(r, w)
	}
	sess.Values[oldSessionKey] = string(b)
	return sess.Save(r, w)
}

// GetOldInputs retrieves old form values (consumes them) similar to flash messages.
func GetOldInputs(w http.ResponseWriter, r *http.Request) (url.Values, error) {
	s := getStore()
	sess, err := s.Get(r, flashSessionName)
	if err != nil {
		return nil, err
	}
	if v, ok := sess.Values[oldSessionKey].(string); ok {
		var vals url.Values
		if err := json.Unmarshal([]byte(v), &vals); err != nil {
			var mm map[string][]string
			if err2 := json.Unmarshal([]byte(v), &mm); err2 == nil {
				vals = url.Values(mm)
			} else {
				return nil, err
			}
		}
		delete(sess.Values, oldSessionKey)
		if err := sess.Save(r, w); err != nil {
			return vals, err
		}
		return vals, nil
	}
	return url.Values{}, nil
}

// Fiber wrappers -------------------------------------------------------------

// FiberSetOldInputs stores form values into Fiber context locals as pending old inputs.
// It no longer saves to session immediately. The pending inputs will be merged into the
// session by FiberSetFlash (when redirecting) or can be committed explicitly via FiberCommitOldInputs.
func FiberSetOldInputs(c *fiber.Ctx, form url.Values) error {
	c.Locals("pending_old_inputs", form)
	return nil
}

// FiberCommitOldInputs writes pending_old_inputs (if any) into the session and applies Set-Cookie headers.
// Returns nil if there were no pending inputs.
func FiberCommitOldInputs(c *fiber.Ctx) error {
	v := c.Locals("pending_old_inputs")
	if v == nil {
		return nil
	}
	form, ok := v.(url.Values)
	if !ok {
		return nil
	}
	w := &fiberResponseWriter{c: c}
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Set(string(k), string(v))
	})
	if err := SetOldInputs(w, r, form); err != nil {
		return err
	}
	w.applyHeaders()
	// clear pending
	c.Locals("pending_old_inputs", nil)
	return nil
}

// FiberGetOldInputs retrieves old form values (consumes them) using Fiber context and applies Set-Cookie headers.
func FiberGetOldInputs(c *fiber.Ctx) (url.Values, error) {
	w := &fiberResponseWriter{c: c}
	r, _ := http.NewRequest(c.Method(), c.OriginalURL(), nil)
	c.Request().Header.VisitAll(func(k, v []byte) {
		r.Header.Set(string(k), string(v))
	})
	vals, err := GetOldInputs(w, r)
	w.applyHeaders()
	return vals, err
}

// FiberOld returns the old value for a key. It caches the loaded old inputs into c.Locals so multiple
// calls in a template don't trigger repeated session reads.
func FiberOld(c *fiber.Ctx, key string) string {
	if v, ok := c.Locals("old_form_cache").(url.Values); ok {
		return v.Get(key)
	}
	vals, _ := FiberGetOldInputs(c)
	c.Locals("old_form_cache", vals)
	return vals.Get(key)
}

// FiberOldAll returns all old inputs (and caches them).
func FiberOldAll(c *fiber.Ctx) url.Values {
	if v, ok := c.Locals("old_form_cache").(url.Values); ok {
		return v
	}
	vals, _ := FiberGetOldInputs(c)
	c.Locals("old_form_cache", vals)
	return vals
}

// FiberAutoOldInputs is a middleware that automatically captures form inputs on unsafe methods
// and, if the handler returned a redirect (3xx), stores them into the session as old inputs
// for the next request (so templates can use old(ctx, "field")).
func FiberAutoOldInputs() fiber.Handler {
	return func(c *fiber.Ctx) error {
		m := c.Method()
		if m != fiber.MethodPost && m != fiber.MethodPut && m != "PATCH" {
			return c.Next()
		}
		ct := c.Get("Content-Type")
		if ct == "" || !(strings.Contains(ct, "application/x-www-form-urlencoded") || strings.Contains(ct, "multipart/form-data")) {
			return c.Next()
		}
		// collect form values
		vals := url.Values{}
		c.Request().PostArgs().VisitAll(func(k, v []byte) {
			key := string(k)
			vals.Add(key, string(v))
		})
		// run handler
		err := c.Next()
		// after handler, if redirect (3xx) then store old inputs (only if not already set by handler)
		status := c.Response().StatusCode()
		if status >= 300 && status < 400 {
			if c.Locals("pending_old_inputs") == nil { // avoid duplicate
				_ = FiberSetOldInputs(c, vals)
			}
		}
		return err
	}
}
