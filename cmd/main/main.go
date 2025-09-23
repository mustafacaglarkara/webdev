package main

import (
	"fmt"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/helpers"
	"github.com/mustafacaglarkara/webdev/pkg/logx"
)

func main() {
	var isim string = "Mustafa ÇAğlar KARA"
	fmt.Println(isim)

	// Logger örneği
	logx.Info("uygulama başladı", "ts", helpers.FormatDate(time.Now(), "2006-01-02 15:04:05"))

	// Küçük bir smoke test
	fmt.Println("Slug:", helpers.ToSlug("Merhaba Dünya! Çalışıyor mu?"))
	fmt.Println("Şu an:", helpers.FormatDate(time.Now(), "2006-01-02 15:04:05"))
	fmt.Println("MD5:", helpers.MD5Hash("parola"))
	fmt.Println("Random(1-3):", helpers.RandomInt(1, 3))

	/*
		var tarayici = bufio.NewScanner(os.Stdin)
		fmt.Println("Bir yazı giriniz:")
		tarayici.Scan()
		yazi := tarayici.Text()
		var sayi, _ = strconv.ParseInt(yazi, 10, 64)
		fmt.Println("Girdiğiniz yazı:", yazi, sayi)

		// --- Helper fonksiyon örnekleri ---
		fmt.Println("Şu anki zaman:", helpers.Now())
		fmt.Println("Tarih formatlı:", helpers.FormatDate(time.Now(), "2006-01-02"))
		fmt.Println("Slug örneği:", helpers.ToSlug("Merhaba Dünya! Çalışıyor mu?"))
		fmt.Println("MD5 hash:", helpers.MD5Hash("parola"))
		fmt.Println("SHA256 hash:", helpers.SHA256Hash("parola"))
		fmt.Println("Base64 encode:", helpers.Base64Encode("test123"))
		decoded, _ := helpers.Base64Decode(helpers.Base64Encode("test123"))
		fmt.Println("Base64 decode:", decoded)
		fmt.Println("String ters çevir:", helpers.ReverseString("GoLang"))
		fmt.Println("Random sayı (1-100):", helpers.RandomInt(1, 100))
		fmt.Println("Ondalık yuvarla:", helpers.RoundFloat(3.14159, 2))
	*/
}
