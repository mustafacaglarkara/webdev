package conv

import (
	"math"
	"testing"
	"time"
)

func TestToConversions(t *testing.T) {
	if ToInt("42", 0) != 42 || ToInt("abc", 5) != 5 || ToInt(" 42 \n", 0) != 42 {
		t.Error("ToInt")
	}
	if ToInt64("123456789012", 0) != 123456789012 || ToInt64(" 7 ", 0) != 7 || ToInt64("x", -1) != -1 {
		t.Error("ToInt64")
	}
	if ToFloat64("3.14", 0) != 3.14 || ToFloat64(" 2.5 ", 0) != 2.5 || ToFloat64("3,14", 1) != 1 {
		t.Error("ToFloat64")
	}
	if !ToBool("true", false) || !ToBool(" 1 ", false) || ToBool("evet", false) {
		t.Error("ToBool")
	}
	if ToDuration("2h45m", 0) != 2*time.Hour+45*time.Minute || ToDuration(" 1s ", 0) != time.Second || ToDuration("x", time.Minute) != time.Minute {
		t.Error("ToDuration")
	}
	if v, err := ParseInt(" 12 "); err != nil || v != 12 {
		t.Error("ParseInt")
	}
	if _, err := ParseInt("12a"); err == nil {
		t.Error("ParseInt hata dönmeli")
	}
	if v, err := ParseFloat("1.5"); err != nil || v != 1.5 {
		t.Error("ParseFloat")
	}
}

// Regresyon (CNV-1): eski RandomInt max < min veya tam int aralığında panik atıyordu.
func TestRandomIntNoPanic(t *testing.T) {
	cases := [][2]int{{10, 20}, {20, 10}, {5, 5}, {-3, 3}, {math.MinInt, math.MaxInt}, {math.MaxInt, math.MinInt}, {math.MinInt, 0}, {0, math.MaxInt}, {math.MaxInt - 1, math.MaxInt}}
	for _, c := range cases {
		lo, hi := min(c[0], c[1]), max(c[0], c[1])
		for range 200 {
			v := RandomInt(c[0], c[1])
			if v < lo || v > hi {
				t.Fatalf("RandomInt(%d,%d) = %d aralık dışı", c[0], c[1], v)
			}
			s, err := SecureRandomInt(c[0], c[1])
			if err != nil {
				t.Fatalf("SecureRandomInt: %v", err)
			}
			if s < lo || s > hi {
				t.Fatalf("SecureRandomInt(%d,%d) = %d aralık dışı", c[0], c[1], s)
			}
		}
	}
}

func TestRandomIntCoversRange(t *testing.T) {
	seen := map[int]bool{}
	seenSecure := map[int]bool{}
	for range 2000 {
		seen[RandomInt(1, 6)] = true
		v, _ := SecureRandomInt(6, 1)
		seenSecure[v] = true
	}
	for i := 1; i <= 6; i++ {
		if !seen[i] || !seenSecure[i] {
			t.Errorf("%d hiç üretilmedi", i)
		}
	}
}

func TestRoundFloat(t *testing.T) {
	cases := []struct {
		v    float64
		p    int
		want float64
	}{
		{3.14159, 2, 3.14},
		{2.5, 0, 3},
		{-2.5, 0, -3},
		{1234.5, -2, 1200},
		{1.5, 400, 1.5},
		{123, -400, 0},
	}
	for _, c := range cases {
		if got := RoundFloat(c.v, c.p); got != c.want {
			t.Errorf("RoundFloat(%v,%d) = %v, want %v", c.v, c.p, got, c.want)
		}
	}
	if !math.IsNaN(RoundFloat(math.NaN(), 2)) {
		t.Error("NaN korunmalı")
	}
	if !math.IsInf(RoundFloat(math.Inf(1), 2), 1) || !math.IsInf(RoundFloat(math.Inf(-1), 2), -1) {
		t.Error("Inf korunmalı")
	}
	if got := RoundFloat(math.MaxFloat64, 2); got != math.MaxFloat64 {
		t.Errorf("taşma: %v", got)
	}
}
