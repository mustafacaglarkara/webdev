package text

import "github.com/microcosm-cc/bluemonday"

// Paket düzeyindeki politikalar bir kez kurulur ve yeniden kullanılır.
// bluemonday.Policy kurulduktan sonra Sanitize çağrıları için eşzamanlı
// kullanıma uygundur; bu nesneler dışarıya verilmez, böylece değiştirilemezler.
var (
	ugcPolicy    = bluemonday.UGCPolicy()
	strictPolicy = bluemonday.StrictPolicy()
)

// HTMLPolicyUGC kullanıcı içeriği (UGC) için bluemonday politikasının YENİ bir
// kopyasını döner. Dönen politika özelleştirilip SanitizeHTMLWith ile
// kullanılabilir; özelleştirmeler SanitizeHTML'in varsayılanını etkilemez.
// Politikayı bir kez kurup saklayın; her istekte yeniden oluşturmayın.
func HTMLPolicyUGC() *bluemonday.Policy { return bluemonday.UGCPolicy() }

// HTMLPolicyStrict tüm etiketleri kaldıran bluemonday politikasının YENİ bir
// kopyasını döner.
func HTMLPolicyStrict() *bluemonday.Policy { return bluemonday.StrictPolicy() }

// SanitizeHTML kullanıcı girdisini UGC politikasıyla temizler: <script>,
// <style>, olay öznitelikleri (onclick ...) ve javascript: bağlantıları
// kaldırılır; <b>, <i>, <a href>, <p>, listeler vb. korunur.
func SanitizeHTML(s string) string { return ugcPolicy.Sanitize(s) }

// SanitizeHTMLStrict tüm HTML etiketlerini kaldırır, yalnızca metin (HTML
// kaçışlı olarak) kalır.
func SanitizeHTMLStrict(s string) string { return strictPolicy.Sanitize(s) }

// SanitizeHTMLWith verilen politika ile temizler. p nil ise SanitizeHTML
// ile aynı varsayılan UGC politikası kullanılır.
func SanitizeHTMLWith(p *bluemonday.Policy, s string) string {
	if p == nil {
		return ugcPolicy.Sanitize(s)
	}
	return p.Sanitize(s)
}
