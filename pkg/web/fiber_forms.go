package web

import (
	"encoding/json"
	"net/url"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/forms"
)

// FiberForm collects request form fields (x-www-form-urlencoded or multipart) into a forms.Form.
func FiberForm(c *fiber.Ctx) *forms.Form {
	data := map[string]any{}
	// urlencoded and multipart fields via fasthttp PostArgs
	args := c.Request().PostArgs()
	if args != nil {
		args.VisitAll(func(k, v []byte) {
			key := string(k)
			val := string(v)
			if old, ok := data[key]; ok {
				switch ov := old.(type) {
				case []string:
					data[key] = append(ov, val)
				case string:
					data[key] = []string{ov, val}
				default:
					data[key] = val
				}
			} else {
				data[key] = val
			}
		})
	}
	// also include multipart form fields if any (Fiber helper)
	if mf, err := c.MultipartForm(); err == nil && mf != nil {
		for k, vals := range mf.Value {
			if len(vals) == 1 {
				data[k] = vals[0]
			} else {
				copyVals := make([]string, len(vals))
				copy(copyVals, vals)
				data[k] = copyVals
			}
		}
	}
	return forms.NewFromMap(data)
}

// FiberJSONForm reads JSON body into a forms.Form (use for application/json).
func FiberJSONForm(c *fiber.Ctx) *forms.Form {
	b := c.Body()
	if len(b) == 0 {
		return forms.NewFromMap(map[string]any{})
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		// fallback: put whole body as _raw
		return forms.NewFromMap(map[string]any{"_raw": string(b)})
	}
	return forms.NewFromMap(m)
}

// helper: convert url.Values to map[string]any (not exported)
func valuesToMap(v url.Values) map[string]any {
	m := make(map[string]any, len(v))
	for k, vals := range v {
		if len(vals) == 1 {
			m[k] = vals[0]
		} else {
			arr := make([]string, len(vals))
			copy(arr, vals)
			m[k] = arr
		}
	}
	return m
}
