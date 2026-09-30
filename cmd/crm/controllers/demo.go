package controllers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mustafacaglarkara/webdev/cmd/crm/menusvc"
	"github.com/mustafacaglarkara/webdev/pkg/web"
	"github.com/mustafacaglarkara/webdev/pkg/web/fiberweb"
)

// MenuDemo GET /demo/menu — depodan yüklenen menünün, geçerli kullanıcı için süzülmüş
// hâlini tablo olarak gösterir.
func (h *Handlers) MenuDemo(c *fiber.Ctx) error {
	all := menusvc.Flatten(h.Menu.Items(c.UserContext()))
	visible := menusvc.Flatten(h.Menu.For(c))
	u, _ := fiberweb.CurrentUser(c)
	return h.render(c, "demo_menu", fiber.Map{
		"Title":       h.T(c, "demo_menu.title"),
		"Rows":        visible,
		"Role":        web.ExtractUserRole(u, web.GuestRole),
		"RecordCount": len(all),
		"HiddenCount": len(all) - len(visible),
	})
}
