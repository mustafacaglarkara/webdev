# fs

Dosya sistemi, arşiv (zip / rar / 7z), görsel işleme, FTP ve SMTP e-posta
gönderimi için yardımcılar.

```go
import "github.com/mustafacaglarkara/webdev/pkg/fs"
```

## Dosya ve dizin

| Fonksiyon | Açıklama |
|---|---|
| `FileExists(path) bool`, `DirExists(path) bool` | Varlık kontrolü |
| `EnsureDir(path) error` | Dizin yoksa oluşturur (`0755`) |
| `ReadFileString(path) (string, error)` | Dosyayı metin olarak okur |
| `WriteFileString(path, data) error` | Üst dizini oluşturup yazar (`0644`) |
| `CopyFile(src, dst) error` | Kopyalar; kaynak ve hedef aynı dosyaysa `ErrSameFile` döner, dosyaya dokunmaz |
| `Remove(path) error` | Dosya siler |
| `ListDirFiles(path) ([]string, error)` | Dizindeki dosyalar (alt dizinler hariç) |
| `Walk(root, fn) error` | `filepath.WalkDir` sarmalayıcısı |
| `IsAllowedExtension(name, allowed) bool` | Uzantı beyaz listesi (`.jpg` veya `jpg`) |
| `DetectMIMEFromFile(path)`, `IsAllowedMIMEFromFile(path, allowed)` | İçerikten MIME tespiti |

```go
if err := fs.CopyFile("a.txt", "yedek/a.txt"); errors.Is(err, fs.ErrSameFile) {
    // kaynak == hedef
}
ok, mime, err := fs.IsAllowedMIMEFromFile("upload.bin", []string{"image/", "application/pdf"})
```

## Arşivler

### Oluşturma

```go
err := fs.ZipDir("./rapor", "./rapor.zip")
err = fs.CreateZipFromPaths([]string{"a.txt", "./klasor"}, "out.zip")
```

- Hedef zip kaynak dizinin içindeyse arşive kendisi eklenmez.
- Yalnızca normal dosyalar ve dizinler arşivlenir (symlink/cihaz dosyaları atlanır).
- `Close` hataları döndürülür; hata durumunda yarım arşiv silinir.

### Çıkarma (güvenli)

```go
// Varsayılan limitlerle
err := fs.ExtractZip("upload.zip", "./hedef")
err = fs.ExtractRARNative("upload.rar", "./hedef")

// Özel limitlerle
opts := fs.ExtractOptions{
    MaxFileBytes:  50 << 20,  // dosya başı 50 MiB
    MaxTotalBytes: 200 << 20, // toplam 200 MiB
    MaxEntries:    1000,
    SkipSymlinks:  true,      // symlink girdilerini hata yerine atla
    MaxRARDictionaryBytes: 64 << 20, // yalnızca RAR: çözücü sözlük sınırı (varsayılan 256 MiB)
}
err = fs.ExtractZipWithOptions("upload.zip", "./hedef", opts)

var limErr *fs.ExtractLimitError
switch {
case errors.Is(err, fs.ErrUnsafePath):
    // zip-slip, mutlak yol, symlink ...
case errors.As(err, &limErr):
    fmt.Println("limit aşıldı:", limErr.Limit, limErr.Max)
}
```

Varsayılanlar: `DefaultMaxExtractFileBytes` (1 GiB), `DefaultMaxExtractTotalBytes`
(4 GiB), `DefaultMaxExtractEntries` (10000), `DefaultMaxRARDictionaryBytes` (256 MiB).
`ExtractOptions` içinde sıfır bırakılan alan varsayılanı kullanır. RAR çözümü
`github.com/nwaples/rardecode/v2` ile yapılır; v1'deki sınırsız sözlük boyutu açığı
(GO-2025-4020) `MaxRARDictionaryBytes` ile kapatılmıştır.

### Harici araçlarla (unrar / 7z)

```go
err := fs.ExtractRarCLI("a.rar", "./out")          // 10 dk süre sınırı (DefaultCLITimeout)
err = fs.Extract7zCLI("a.7z", "./out")             // 7z, 7zz veya 7za aranır
ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
defer cancel()
err = fs.ExtractRarCLIContext(ctx, "a.rar", "./out")
err = fs.Extract7zCLIContext(ctx, "a.7z", "./out")
```

`ExtractRar` / `Extract7z` geriye dönük takma adlardır. `HasNativeRAR()` true,
`HasNative7z()` false döner; `Extract7zNative` uygulanmamıştır (hata döner).

## Görseller

```go
img, format, err := fs.LoadImage("foto.jpg")                       // DefaultMaxImagePixels (50 MP) limiti
img, format, err = fs.LoadImageWithLimit("foto.jpg", 20_000_000)
w, h, err := fs.GetDimensions("foto.webp")                         // yalnızca başlık okunur
err = fs.SaveImageWithQuality(img, "cikti.jpg", 85)                // .jpg .jpeg .png .gif .webp
err = fs.ResizeImage("foto.jpg", "kucuk.jpg", 800, 0, 85)          // 0: oran korunur
err = fs.GenerateThumbnail("foto.jpg", "thumb.png", 200, 80)
err = fs.ConvertToWebP("foto.png", "foto.webp", 80)
err = fs.OptimizeJPEG("foto.jpg", "foto-opt.jpg", 70)
b64, err := fs.EncodeImageBase64(img, "png", 0)                    // jpeg/jpg, png, webp
```

- Decode öncesi `DecodeConfig` ile boyut kontrol edilir; limit aşılırsa `ErrImageTooLarge`.
- Çıktı biçimi dosya oluşturulmadan doğrulanır (`ErrUnsupportedImageFormat`).
- Kayıt geçici dosyaya yapılıp `rename` edilir; hata durumunda yarım dosya kalmaz.

## FTP

```go
c, err := fs.NewFTPClientWithOptions("ftp.example.com:21", "user", "pass", fs.FTPOptions{
    Timeout:   10 * time.Second,
    TLSConfig: &tls.Config{ServerName: "ftp.example.com"}, // explicit FTPS (AUTH TLS)
})
if err != nil {
    return err
}
defer c.Close()

f, _ := os.Open("buyuk.iso")
defer f.Close()
err = c.UploadFrom("/uploads/buyuk.iso", f) // akış olarak

out, _ := os.Create("indirilen.iso")
defer out.Close()
n, err := c.DownloadTo("/uploads/buyuk.iso", out)
```

Diğer metotlar: `Upload`, `Download` (belleğe), `Delete`, `List`, `Rename`,
`MakeDir`, `RemoveDir`, `Close`. `NewFTPClient(addr, user, pass)` şifresiz FTP
kullanır. Giriş başarısız olursa bağlantı kapatılır. `ImplicitTLS: true` ile
implicit FTPS (genelde 990) kullanılır.

## E-posta (SMTP)

```go
cfg := fs.SMTPCfg{
    Host: "smtp.example.com", Port: 587, User: "u", Pass: "p",
    Timeout:       20 * time.Second, // deneme başına; 0 => DefaultSMTPTimeout (30 sn)
    RetryAttempts: 3, RetryDelay: 2 * time.Second,
    RateInterval:  time.Minute, RateMax: 60,
}
err := fs.SendEmail(ctx, cfg, "Ben <ben@example.com>",
    []string{"a@example.com"}, nil, nil,
    "Konu", "düz metin", "<p>html</p>", []string{"./rapor.pdf"})
if errors.Is(err, fs.ErrEmailRateLimited) { /* ... */ }
```

- `UseTLS: true` implicit TLS (465); aksi halde sunucu destekliyorsa STARTTLS.
- Yalnızca geçici hatalar (`421`, `450`, `451`, `452` yanıt kodu, ağ zaman aşımı) yeniden denenir.
- Mesaj gövdesi gönderilmeye başladıktan sonraki hatalar yeniden denenmez (çift gönderim önlenir).
- Süre sınırı bağlantıya uygulanır; `ctx` iptali bağlantıyı hemen keser, sızan goroutine kalmaz.
- Hız sınırı yapılandırma başınadır (host, port, user); başarısız gönderim hak tüketmez.
- Okunamayan, dizin olan veya `http(s)://` ekler hata döndürür.

`EnsureAttachmentDir(path)` ek dosyası için üst dizini oluşturur.

## Güvenlik notları

- **Arşiv çıkarma**: `ExtractZip*` ve `ExtractRARNative*` tek bir güvenli yol
  birleştirme yardımcısı kullanır: girdi adı temizlenir; mutlak yol, sürücü
  harfi, `..` ile dışarı çıkma, symlink girdisi ve hedefte zaten var olan bir
  symlink üzerinden yazma reddedilir. Önek kontrolü ayraçla yapılır
  (`/dest-evil`, `/dest` içinde sayılmaz). Dosya başı / toplam boyut ve girdi
  sayısı limitleri "zip bomb"a karşı korur; aşımda yarım dosya silinir.
  Setuid/setgid bitleri uygulanmaz.
- **Harici araçlar**: `ExtractRarCLI`/`Extract7zCLI` argümanları `--` ile verir
  ve süre sınırı uygular; ancak yol güvenliği ve boyut limitleri aracın
  kendisine kalır. Güvenilmeyen arşivler için yerleşik çıkarıcıları tercih edin.
- **Görseller**: piksel limiti "decompression bomb"a karşı korur.
- **FTP**: `NewFTPClient` parolayı açık metin gönderir; mümkünse `TLSConfig` kullanın.
