package web

import (
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// FiberLogCookieSizes: geliştirme ortamında Set-Cookie header toplam boyutunu loglar.
// thresholdBytes > 0 ise aşıldığında ayrıca uyarı loglar.
func FiberLogCookieSizes(thresholdBytes int) fiber.Handler {
	return func(c *fiber.Ctx) error {
		err := c.Next()
		if err != nil {
			return err
		}
		// sadece HTML veya redirect cevaplarında kontrol ederiz
		ct := c.Get("Content-Type")
		if ct == "" || strings.Contains(ct, "text/html") || (c.Response().StatusCode() >= 300 && c.Response().StatusCode() < 400) {
			var cookies [][]byte
			c.Response().Header.VisitAll(func(k, v []byte) {
				if string(k) == "Set-Cookie" {
					// kopya al (v slice reuse edilmesin)
					cookies = append(cookies, append([]byte(nil), v...))
				}
			})
			total := 0
			for _, ck := range cookies {
				total += len(ck)
			}
			if total > 0 {
				log.Printf("[dev] Set-Cookie headers count=%d totalBytes=%d", len(cookies), total)
				if thresholdBytes > 0 && total > thresholdBytes {
					log.Printf("[warn] Set-Cookie total %d exceeds threshold %d", total, thresholdBytes)
				}
			}
		}
		return nil
	}
}
