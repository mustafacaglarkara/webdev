package timex

import (
	"context"
	"errors"
	"testing"
	"time"
	_ "time/tzdata" // testlerin sistem tz veritabanından bağımsız çalışması için
)

func TestStartEndOfDay(t *testing.T) {
	ist := Istanbul()
	in := time.Date(2025, 9, 25, 14, 30, 0, 0, ist)
	if got, want := StartOfDay(in), time.Date(2025, 9, 25, 0, 0, 0, 0, ist); !got.Equal(want) {
		t.Errorf("StartOfDay = %v", got)
	}
	if got, want := EndOfDay(in), time.Date(2025, 9, 25, 23, 59, 59, 999999999, ist); !got.Equal(want) {
		t.Errorf("EndOfDay = %v", got)
	}
	if StartOfDay(in).Location() != ist {
		t.Error("konum korunmalı")
	}
}

// Regresyon (TMX-2): gece yarısının var olmadığı günlerde StartOfDay önceki güne düşüyordu.
func TestStartOfDayDSTGap(t *testing.T) {
	loc, err := time.LoadLocation("America/Santiago")
	if err != nil {
		t.Skip(err)
	}
	day := time.Date(2024, 9, 8, 15, 0, 0, 0, loc) // 2024-09-08 00:00 Santiago'da yok
	s := StartOfDay(day)
	if y, m, d := s.Date(); y != 2024 || m != 9 || d != 8 {
		t.Fatalf("StartOfDay başka güne düştü: %v", s)
	}
	if s.Hour() != 1 {
		t.Errorf("günün ilk anı 01:00 olmalı: %v", s)
	}
	prevEnd := EndOfDay(time.Date(2024, 9, 7, 12, 0, 0, 0, loc))
	if !prevEnd.Add(time.Nanosecond).Equal(s) {
		t.Errorf("EndOfDay(önceki gün)+1ns = %v, StartOfDay = %v", prevEnd.Add(time.Nanosecond), s)
	}
	// Normal günlerde fark 24 saat değildir (23 saatlik gün).
	if got := EndOfDay(day).Sub(s) + time.Nanosecond; got != 23*time.Hour {
		t.Errorf("gün uzunluğu = %v", got)
	}
}

func TestParseTime(t *testing.T) {
	tm, err := ParseTime("2006-01-02 15:04", "2025-09-25 10:00")
	if err != nil || tm.Location() != time.UTC || tm.Hour() != 10 {
		t.Fatalf("ParseTime UTC dönmeli: %v %v", tm, err)
	}
	if _, err := ParseTime("2006-01-02", "25.09.2025"); err == nil {
		t.Error("hata beklenirdi")
	}
	defer func() {
		if recover() == nil {
			t.Error("MustParseTime panik atmalı")
		}
	}()
	if MustParseTime("2006-01-02", "2025-09-25").Day() != 25 {
		t.Error("MustParseTime")
	}
	MustParseTime("2006-01-02", "x")
}

func TestParseTimeIn(t *testing.T) {
	tm, err := ParseTimeIstanbul("02.01.2006 15:04", "25.09.2025 10:00")
	if err != nil {
		t.Fatal(err)
	}
	if got := tm.UTC().Format("15:04"); got != "07:00" {
		t.Errorf("İstanbul 10:00 = UTC %s, want 07:00", got)
	}
	tm2, err := ParseTimeIn("2006-01-02 15:04", "2025-09-25 10:00", Istanbul())
	if err != nil || !tm2.Equal(tm) {
		t.Errorf("ParseTimeIn = %v %v", tm2, err)
	}
	tm3, err := ParseTimeIn("2006-01-02", "2025-09-25", nil)
	if err != nil || tm3.Location() != time.UTC {
		t.Errorf("nil loc UTC olmalı: %v %v", tm3, err)
	}
	if _, off := time.Date(2025, 1, 1, 0, 0, 0, 0, Istanbul()).Zone(); off != 3*3600 {
		t.Errorf("İstanbul ofseti = %d", off)
	}
}

func TestSleepCtx(t *testing.T) {
	if !SleepCtx(nil, time.Millisecond) {
		t.Error("süre dolmalıydı")
	}
	done := make(chan struct{})
	close(done)
	// Regresyon: iptal edilmiş kanal, d=0 olsa bile önceliklidir.
	for range 50 {
		if SleepCtx(done, 0) {
			t.Fatal("iptal önceliği yok")
		}
	}
	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if SleepCtx(ctx.Done(), 5*time.Second) {
		t.Error("iptal beklenirdi")
	}
	if time.Since(start) > 2*time.Second {
		t.Error("iptal geç algılandı")
	}
}

func TestSleepContext(t *testing.T) {
	if err := SleepContext(context.Background(), time.Millisecond); err != nil {
		t.Error(err)
	}
	var nilCtx context.Context // nil ctx bilinçli olarak test ediliyor
	if err := SleepContext(nilCtx, 0); err != nil {
		t.Error(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := SleepContext(ctx, 0); !errors.Is(err, context.Canceled) {
		t.Errorf("önceden iptal: %v", err)
	}
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel2()
	if err := SleepContext(ctx2, 5*time.Second); !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("timeout: %v", err)
	}
}

func TestMisc(t *testing.T) {
	a := time.Date(2025, 9, 25, 0, 0, 0, 0, time.UTC)
	if DateDiff(a, a.Add(-48*time.Hour)) != 48*time.Hour {
		t.Error("DateDiff")
	}
	if FormatDate(a, "02.01.2006") != "25.09.2025" {
		t.Error("FormatDate")
	}
	if d := Timestamp() - Now().Unix(); d < -1 || d > 1 {
		t.Error("Timestamp")
	}
}
