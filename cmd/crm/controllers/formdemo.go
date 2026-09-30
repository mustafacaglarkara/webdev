package controllers

import (
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/pkg/forms"
	"github.com/mustafacaglarkara/webdev/pkg/text"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// formErrorsFlashKey doğrulama hatalarını yönlendirme boyunca taşıyan flash anahtarı.
const formErrorsFlashKey = "formdemo_errors"

// formDemoRules pkg/validation kural dizgileri.
var formDemoRules = map[string]string{
	"name":    "required|min:2|max:80",
	"email":   "required|email",
	"age":     "integer|between:18,120",
	"topic":   "required|in:sales,support,other",
	"message": "max:500",
	"agree":   "required",
}

// formDemoMessages alan.kural → locale anahtarı. Mesajlar isteğin dilinde üretilir.
var formDemoMessages = map[string]string{
	"name.required":  "formdemo.err.name_required",
	"name.min":       "formdemo.err.name_min",
	"name.max":       "formdemo.err.name_max",
	"email.required": "formdemo.err.email_required",
	"email.email":    "formdemo.err.email_invalid",
	"age.integer":    "formdemo.err.age_integer",
	"age.between":    "formdemo.err.age_between",
	"topic.required": "formdemo.err.topic_required",
	"topic.in":       "formdemo.err.topic_in",
	"message.max":    "formdemo.err.message_max",
	"agree.required": "formdemo.err.agree_required",
}

// FormDemo GET /forms/demo. Önceki gönderimin hataları flash'tan, değerleri old
// input'tan (şablonda old(ctx, "alan")) okunur.
func (h *Handlers) FormDemo(c *fiber.Ctx) error {
	f := forms.NewFromMap(nil)
	if raw := fiberweb.Flash(c)(formErrorsFlashKey); raw != "" {
		errs := map[string]string{}
		if err := json.Unmarshal([]byte(raw), &errs); err != nil {
			slog.Debug("formdemo: bad errors flash", "err", err)
		}
		for k, v := range errs {
			f.AddError(k, v)
		}
		f.Validated = true
	}
	return h.render(c, "formdemo", fiber.Map{"Title": h.T(c, "formdemo.title"), "Form": f})
}

// FormDemoSubmit POST /forms/demo (CSRF korumalı). Post/Redirect/Get: hata varsa değerler
// old input'a, hatalar flash'a yazılır ve forma dönülür; başarıda yalnızca flash yazılır.
func (h *Handlers) FormDemoSubmit(c *fiber.Ctx) error {
	f := fiberweb.Form(c)
	if s, ok := f.Data["name"].(string); ok {
		f.Data["name"] = text.NormalizeSpace(s)
	}
	msgs := make(map[string]string, len(formDemoMessages))
	for k, key := range formDemoMessages {
		msgs[k] = h.T(c, key)
	}
	if !f.ValidateMapWithMessages(formDemoRules, msgs) {
		_ = fiberweb.SetOldInputs(c, formValues(f))
		b, err := json.Marshal(f.Errors)
		if err != nil {
			return err
		}
		if err := fiberweb.AddFlash(c, formErrorsFlashKey, string(b)); err != nil {
			return err
		}
		return fiberweb.SetFlash(c, "error", h.T(c, "formdemo.failed"), "/forms/demo", fiber.StatusSeeOther)
	}
	name, _ := f.CleanedData()["name"].(string)
	msg := h.T(c, "formdemo.success", map[string]any{
		"Name": text.TitleTR(strings.ToLower(name)),
		"Slug": text.ToSlug(name),
	})
	return fiberweb.SetFlash(c, "success", msg, "/forms/demo", fiber.StatusSeeOther)
}
