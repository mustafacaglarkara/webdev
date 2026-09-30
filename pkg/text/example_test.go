package text_test

import (
	"fmt"

	"github.com/mustafacaglarkara/webdev/pkg/text"
)

func ExampleToSlug() {
	fmt.Println(text.ToSlug("Çalışma Alanı - 2025!"))
	fmt.Println(text.ToSlug("İstanbul IŞIK"))
	fmt.Println(text.ToSlug("Kâğıt café crème"))
	fmt.Printf("%q\n", text.ToSlug("日本語"))
	// Output:
	// calisma-alani-2025
	// istanbul-isik
	// kagit-cafe-creme
	// ""
}

func ExampleToSlugForFile() {
	fmt.Println(text.ToSlugForFile("Çılgın Fotoğraf(1).JPG"))
	fmt.Println(text.ToSlugForFile("rapor.v1.2.PDF"))
	fmt.Println(text.ToSlugForFile("../../etc/passwd"))
	fmt.Println(text.ToSlugForFile(".jpg"))
	fmt.Println(text.ToSlugForFile(""))
	fmt.Println(text.ToSlugForFile("belge."))
	// Output:
	// cilgin-fotograf-1.jpg
	// rapor-v1-2.pdf
	// etc-passwd
	// file.jpg
	// file
	// belge
}

func ExampleToUpperTR() {
	fmt.Println(text.ToUpper("istanbul ılık"))
	fmt.Println(text.ToUpperTR("istanbul ılık"))
	fmt.Println(text.ToLowerTR("IŞIK İSTANBUL"))
	fmt.Println(text.TitleTR("iSTANBUL ılık"))
	// Output:
	// ISTANBUL ILIK
	// İSTANBUL ILIK
	// ışık istanbul
	// İstanbul Ilık
}

func ExampleFixTurkishMojibake() {
	fmt.Println(text.FixTurkishMojibake("GÃ¼nÃ¼n Ã¶zeti: Ã‡aÄŸdaÅŸlÄ±k"))
	fmt.Println(text.FixTurkishMojibake("ÅŸ Åž Ä± Ä°"))
	fmt.Println(text.FixTurkishMojibake("KÂR"))
	// Output:
	// Günün özeti: Çağdaşlık
	// ş Ş ı İ
	// KÂR
}

func ExampleSanitizeHTML() {
	unsafe := `<script>alert('x')</script><b>kalın</b> <a href="http://ex.com" onclick="x()">link</a>`
	fmt.Println(text.SanitizeHTML(unsafe))
	fmt.Println(text.SanitizeHTMLStrict(unsafe))

	p := text.HTMLPolicyUGC()
	p.AllowAttrs("class").OnElements("span")
	fmt.Println(text.SanitizeHTMLWith(p, `<span class="not">x</span>`))
	// Output:
	// <b>kalın</b> <a href="http://ex.com" rel="nofollow">link</a>
	// kalın link
	// <span class="not">x</span>
}

func ExampleTruncate() {
	fmt.Println(text.Truncate("merhaba dünya", 9))
	fmt.Println(text.NormalizeSpace("  çok   fazla\t boşluk \n"))
	fmt.Println(text.ReverseString("İstanbul"))
	fmt.Println(text.Coalesce("", "", "ilk"))
	fmt.Println(text.IsBlank(" \t\n"))
	fmt.Println(text.UnescapeHTML("&Ccedil;alışma &amp; Deneme"))
	fmt.Println(text.SplitAndTrim(" a , b ,, c "))
	// Output:
	// merhaba d
	// çok fazla boşluk
	// lubnatsİ
	// ilk
	// true
	// Çalışma & Deneme
	// [a b c]
}
