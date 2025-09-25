# pkg/helpers/config

This package contains configuration-related helpers used by the project. The functions are exposed via the top-level `pkg/helpers` wrappers for backwards compatibility.

Main features

- Environment helpers: `helpers.GetEnv`, `helpers.MustGetEnv`, `helpers.GetEnvInt`, `helpers.GetEnvBool`, `helpers.GetEnvDuration`
- YAML helpers: `helpers.ToYAML`, `helpers.FromYAML`, `helpers.WriteYAMLFile`, `helpers.ReadYAMLFile`

Quick examples

1) Reading environment variables with fallback

```go
port := helpers.GetEnv("PORT", "8080")
// parse if needed
```

2) Reading duration

```go
timeout := helpers.GetEnvDuration("TIMEOUT", 5*time.Second)
```

3) YAML serialize/deserialize

```go
cfg := struct{ Name string `yaml:"name"` }{Name: "svc"}
str, _ := helpers.ToYAML(cfg)
var out struct{ Name string `yaml:"name"` }
_ = helpers.FromYAML[struct{ Name string `yaml:"name"` }](str)
```

Notes

- The functions are thin wrappers pointing to `pkg/helpers/config` implementation; keep using `pkg/helpers` imports in existing code.

# config paketi

Ortam değişkenleri (ENV) ve YAML dosya işlemleri için yardımcı fonksiyonlar içerir. Uygulama yapılandırmasını kolayca yönetmek için kullanılır.

## Fonksiyonlar ve Kullanım Örnekleri

### Ortam Değişkenleri

#### GetEnv
Bir ortam değişkenini okur, yoksa varsayılan değeri döner.
```go
import "your/module/path/pkg/config"
port := config.GetEnv("PORT", "8080")
```

#### MustGetEnv
Bir ortam değişkenini okur, yoksa panic atar.
```go
val := config.MustGetEnv("SECRET_KEY")
```

#### GetEnvInt
Bir ortam değişkenini int olarak okur, yoksa varsayılan değeri döner.
```go
max := config.GetEnvInt("MAX_CONN", 10)
```

#### GetEnvBool
Bir ortam değişkenini bool olarak okur, yoksa varsayılan değeri döner.
```go
debug := config.GetEnvBool("DEBUG", false)
```

#### GetEnvDuration
Bir ortam değişkenini time.Duration olarak okur, yoksa varsayılan değeri döner.
```go
timeout := config.GetEnvDuration("TIMEOUT", 5*time.Second)
```

### YAML İşlemleri

#### ToYAML
Bir değeri YAML string'e çevirir.
```go
cfg := struct{ Name string `yaml:"name"` }{Name: "svc"}
yamlStr, _ := config.ToYAML(cfg)
```

#### FromYAML
Bir YAML string'i struct'a çevirir.
```go
var out struct{ Name string `yaml:"name"` }
out, _ = config.FromYAML[struct{ Name string `yaml:"name"` }](yamlStr)
```

#### WriteYAMLFile
Bir değeri YAML dosyasına yazar.
```go
err := config.WriteYAMLFile("config.yaml", cfg)
```

#### ReadYAMLFile
Bir YAML dosyasını struct olarak okur.
```go
var cfg struct{ Name string `yaml:"name"` }
cfg, err := config.ReadYAMLFile[struct{ Name string `yaml:"name"` }]("config.yaml")
```

## Notlar
- Ortam değişkeni fonksiyonları, hatalı veya eksik değerlerde güvenli şekilde varsayılan döner.
- YAML fonksiyonları generic olarak çalışır.
- Detaylar için kodu inceleyebilirsiniz.
