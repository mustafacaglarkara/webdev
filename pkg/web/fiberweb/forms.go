package fiberweb

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/forms"
)

// Form urlencoded veya multipart istek alanlarını forms.Form'a toplar. Birden fazla değeri
// olan alanlar []string olarak yazılır.
func Form(c *fiber.Ctx) *forms.Form {
	data := map[string]any{}
	for k, vals := range formValues(c) {
		if len(vals) == 1 {
			data[k] = vals[0]
		} else {
			data[k] = append([]string(nil), vals...)
		}
	}
	return forms.NewFromMap(data)
}

// JSONForm JSON gövdesini forms.Form'a okur (application/json). Gövde geçersiz JSON ise
// ham gövde "_raw" alanına konur.
func JSONForm(c *fiber.Ctx) *forms.Form {
	b := c.Body()
	if len(b) == 0 {
		return forms.NewFromMap(map[string]any{})
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return forms.NewFromMap(map[string]any{"_raw": string(b)})
	}
	return forms.NewFromMap(m)
}
