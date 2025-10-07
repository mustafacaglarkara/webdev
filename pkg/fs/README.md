# fs paketi

Bu paket, dosya sistemi işlemleri, arşivleme (zip/rar/7z), FTP, resim işleme ve e-posta gönderimi gibi çok amaçlı yardımcı fonksiyonlar içerir.

## Kurulum

Modülünüze ekleyin:

```
go get your/module/path/pkg/fs
```

## Fonksiyonlar ve Kullanım Örnekleri

### Dosya ve Klasör İşlemleri

#### Dosya/Klasör Kontrolü
```go
fs.FileExists("dosya.txt") // bool
dirVar := fs.DirExists("klasor")
```

#### Klasör Oluşturma
```go
err := fs.EnsureDir("yeni/klasor")
```

#### Dosya Okuma/Yazma
```go
s, _ := fs.ReadFileString("dosya.txt")
_ = fs.WriteFileString("yeni.txt", "icerik")
```

#### Dosya Kopyalama/Silme
```go
_ = fs.CopyFile("kaynak.txt", "hedef.txt")
_ = fs.Remove("dosya.txt")
```

#### Klasördeki Dosyaları Listeleme
```go
files, _ := fs.ListDirFiles("klasor")
```

### Upload Doğrulama (Uzantı ve MIME)

#### Uzantı Kontrolü
```go
ok := fs.IsAllowedExtension("resim.JPG", []string{".jpg", "png", "webp"}) // true
```

#### MIME Tipi Tespiti ve Doğrulama
```go
mime, _ := fs.DetectMIMEFromFile("resim.jpg") // örn: "image/jpeg"
allowed, detected, _ := fs.IsAllowedMIMEFromFile("resim.jpg", []string{"image/", "application/pdf"})
// allowed, detected => true, "image/jpeg"
```

Notlar:
- Uzantı kontrolünde noktalı veya noktasız değerler kullanılabilir (case-insensitive).
- MIME doğrulamada baştan eşleşme kabul edilir (ör: "image/" tüm image türlerini kapsar).

### Arşivleme (ZIP, RAR, 7z)

#### Klasörü ZIP'e Arşivleme
```go
_ = fs.ZipDir("klasor", "arsiv.zip")
```

#### Birden Fazla Dosya/Klasörü ZIP'e Arşivleme
```go
_ = fs.CreateZipFromPaths([]string{"a.txt", "b/"}, "hepsi.zip")
```

#### ZIP Çıkarma
```go
_ = fs.ExtractZip("arsiv.zip", "hedef_klasor")
```

#### RAR/7z Çıkarma (CLI veya native)
```go
_ = fs.ExtractRar("arsiv.rar", "hedef")
_ = fs.Extract7z("arsiv.7z", "hedef")
// Native RAR desteği:
_ = fs.ExtractRARNative("arsiv.rar", "hedef")
```

### FTP İşlemleri
```go
ftp, _ := fs.NewFTPClient("localhost:21", "kullanici", "sifre")
defer ftp.Close()
_ = ftp.Upload("uzak.txt", []byte("veri"))
data, _ := ftp.Download("uzak.txt")
_ = ftp.Delete("uzak.txt")
_ = ftp.MakeDir("yeni_klasor")
```

### Resim İşlemleri

#### Resim Yükleme ve Boyut Alma
```go
img, format, _ := fs.LoadImage("resim.jpg")
w, h, _ := fs.GetDimensions("resim.jpg")
```

#### Resim Kaydetme/Kalite Ayarı
```go
_ = fs.SaveImageWithQuality(img, "yeni.jpg", 80)
```

#### WebP Dönüştürme
```go
_ = fs.ConvertToWebP("resim.jpg", "resim.webp", 80)
```

#### Yeniden Boyutlandırma ve Thumbnail
```go
_ = fs.ResizeImage("resim.jpg", "kucuk.jpg", 200, 0, 80)
_ = fs.GenerateThumbnail("resim.jpg", "thumb.jpg", 128, 80)
```

#### Base64 Encode
```go
b64, _ := fs.EncodeImageBase64(img, "jpeg", 80)
```

### E-posta Gönderimi

#### SMTP Ayarları ve Gönderim
```go
cfg := fs.SMTPCfg{
    Host: "smtp.example.com", Port: 587, User: "user", Pass: "pass", UseTLS: true,
    Timeout: 10 * time.Second, RateInterval: time.Minute, RateMax: 10, RetryAttempts: 3,
}
err := fs.SendEmail(ctx, cfg, "from@example.com", []string{"to@example.com"}, nil, nil, "Konu", "Metin", "<b>HTML</b>", nil)
```

## Notlar
- Arşivleme işlemleri için bazı formatlarda harici CLI araçları gerekebilir (unrar, 7z).
- Resim işlemleri için ek kütüphaneler (imaging, webp) gereklidir.
- FTP işlemleri için github.com/jlaffaye/ftp, e-posta için github.com/jordan-wright/email kullanılır.
- Tüm fonksiyonlar hata durumunda error döner, kontrol edilmelidir.

Daha fazla detay için kodu inceleyebilirsiniz.
