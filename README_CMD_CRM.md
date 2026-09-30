# cmd/crm — örnek uygulama

`cmd/crm`, `github.com/mustafacaglarkara/webdev` kütüphanesini gerçek bir Fiber v2 + Jet
uygulamasında gösteren küçük bir CRM'dir: CSRF korumalı giriş/çıkış, flash mesajları,
`?next=` ile güvenli yönlendirme, rol + casbin yetkilendirmesi, veriden yüklenen menü,
Türkçe/İngilizce arayüz, form doğrulama ve JSON API.

Ayrıntılı belge (ortam değişkenleri, route tablosu, paket eşlemesi):
[`cmd/crm/README.md`](cmd/crm/README.md). Dil dosyaları:
[`cmd/crm/locales/README.md`](cmd/crm/locales/README.md).

## Hızlı başlangıç

```bash
go run ./cmd/crm            # http://localhost:8080
```

Geliştirme hesapları: `admin` / `admin123` ve `user` / `user123`
(`CRM_ADMIN_PASSWORD`, `CRM_USER_PASSWORD` ile değiştirilir; varsayılanlar
kullanıldığında başlangıçta uyarı loglanır).

Üretim benzeri:

```bash
CRM_ENV=production \
CRM_SESSION_KEY="$(openssl rand -base64 48)" \
CRM_ADMIN_PASSWORD='...' CRM_USER_PASSWORD='...' \
CRM_PORT=8080 go run ./cmd/crm
```

`production` modunda `CRM_SESSION_KEY` (en az 32 bayt) ve demo parolaları zorunludur;
eksikse uygulama başlamaz.

## Denetim

```bash
go test -race ./cmd/crm/...
go run ./cmd/i18ncheck
curl -s localhost:8080/api/health
```
