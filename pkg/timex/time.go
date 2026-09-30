// Package timex, zaman ve tarih yardımcıları sunar.
package timex

import (
	"context"
	"sync"
	"time"
)

// Now time.Now sarmalayıcısıdır.
func Now() time.Time { return time.Now() }

// FormatDate t'yi layout ile biçimlendirir (t.Format).
func FormatDate(t time.Time, layout string) string { return t.Format(layout) }

// DateDiff a - b farkını döner.
func DateDiff(a, b time.Time) time.Duration { return a.Sub(b) }

// Timestamp şu anki Unix zamanını (saniye) döner.
func Timestamp() int64 { return time.Now().Unix() }

// StartOfDay t'nin kendi konumundaki (t.Location()) gününün ilk anını döner.
// Gece yarısının yaz saati geçişi nedeniyle var olmadığı bölgelerde
// (ör. America/Santiago) günün var olan ilk anı döner; sonuç her zaman t ile
// aynı takvim gününe düşer.
func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return startOfDate(y, m, d, t.Location())
}

// EndOfDay t'nin gününün son nanosaniyesini döner: ertesi günün başlangıcından
// 1ns önce. Yaz saati geçişi olan günlerde de doğru sonuç verir.
func EndOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return startOfDate(y, m, d+1, t.Location()).Add(-time.Nanosecond)
}

// startOfDate verilen günün loc'taki ilk anını döner. time.Date, var olmayan
// bir yerel saat için önceki güne düşen bir an döndürebilir; bu durumda gün
// başlayana kadar ileri gidilir.
func startOfDate(y int, m time.Month, d int, loc *time.Location) time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, loc)
	wy, wm, wd := time.Date(y, m, d, 12, 0, 0, 0, loc).Date() // normalize edilmiş hedef gün
	for range 4 {
		ty, tm, td := t.Date()
		if ty == wy && tm == wm && td == wd {
			return t
		}
		// t önceki güne düştü: t'nin saat dilimi döneminin bittiği an (yaz
		// saati geçişi) hedef günün ilk anıdır.
		_, end := t.ZoneBounds()
		if end.IsZero() || !end.After(t) {
			break
		}
		t = end
	}
	return t
}

// ParseTime value'yu layout ile ayrıştırır (time.Parse). Layout/değer saat
// dilimi bilgisi içermiyorsa sonuç UTC'dir; yerel saat olarak yorumlamak için
// ParseTimeIn veya ParseTimeIstanbul kullanın.
func ParseTime(layout, value string) (time.Time, error) { return time.Parse(layout, value) }

// MustParseTime ParseTime gibidir ancak hata durumunda panik atar
// (yalnızca sabit/test verileri için kullanın). Sonuç UTC'dir.
func MustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

// ParseTimeIn value'yu layout ile loc konumunda ayrıştırır
// (time.ParseInLocation). Saat dilimi bilgisi yoksa değer loc'un yerel saati
// kabul edilir. loc nil ise UTC kullanılır.
func ParseTimeIn(layout, value string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = time.UTC
	}
	return time.ParseInLocation(layout, value, loc)
}

var (
	istanbulOnce sync.Once
	istanbulLoc  *time.Location
)

// Istanbul "Europe/Istanbul" konumunu döner. Sistemde saat dilimi
// veritabanı yoksa sabit UTC+03:00 ("+03") konumuna düşer (Türkiye 2016'dan
// beri yaz saati uygulamıyor; daha eski tarihler için tzdata gerekir —
// programınıza `import _ "time/tzdata"` ekleyebilirsiniz).
func Istanbul() *time.Location {
	istanbulOnce.Do(func() {
		loc, err := time.LoadLocation("Europe/Istanbul")
		if err != nil {
			loc = time.FixedZone("+03", 3*60*60)
		}
		istanbulLoc = loc
	})
	return istanbulLoc
}

// ParseTimeIstanbul value'yu Europe/Istanbul yerel saati olarak ayrıştırır.
func ParseTimeIstanbul(layout, value string) (time.Time, error) {
	return time.ParseInLocation(layout, value, Istanbul())
}

// SleepCtx d süresi kadar bekler; ctxDone kapanırsa hemen döner. Süre
// dolarsa true, iptal edildiyse false döner. ctxDone zaten kapalıysa (d ne
// olursa olsun) false döner; iptal, süre dolumuna göre önceliklidir.
func SleepCtx(ctxDone <-chan struct{}, d time.Duration) bool {
	select {
	case <-ctxDone:
		return false
	default:
	}
	if d <= 0 {
		return true
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctxDone:
		return false
	case <-t.C:
		return true
	}
}

// SleepContext d süresi kadar bekler veya ctx iptal edilene kadar bekler.
// Süre dolarsa nil, ctx iptal edildiyse (önceden iptal edilmiş olsa bile)
// ctx.Err() döner. ctx nil ise context.Background kabul edilir.
func SleepContext(ctx context.Context, d time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
