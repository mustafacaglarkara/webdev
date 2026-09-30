package timex_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/timex"
)

func ExampleParseTime() {
	t, _ := timex.ParseTime("2006-01-02 15:04", "2025-09-25 14:30")
	fmt.Println(t) // saat dilimi yoksa UTC

	ist, _ := timex.ParseTimeIstanbul("02.01.2006 15:04", "25.09.2025 14:30")
	fmt.Println(ist)
	fmt.Println(ist.UTC())

	loc, _ := time.LoadLocation("Europe/Berlin")
	b, _ := timex.ParseTimeIn("2006-01-02 15:04", "2025-09-25 14:30", loc)
	fmt.Println(b)
	// Output:
	// 2025-09-25 14:30:00 +0000 UTC
	// 2025-09-25 14:30:00 +0300 +03
	// 2025-09-25 11:30:00 +0000 UTC
	// 2025-09-25 14:30:00 +0200 CEST
}

func ExampleStartOfDay() {
	t := time.Date(2025, 9, 25, 14, 30, 0, 0, timex.Istanbul())
	fmt.Println(timex.StartOfDay(t))
	fmt.Println(timex.EndOfDay(t))
	fmt.Println(timex.FormatDate(t, "02.01.2006 15:04"))
	fmt.Println(timex.DateDiff(t, t.Add(-48*time.Hour)))
	// Output:
	// 2025-09-25 00:00:00 +0300 +03
	// 2025-09-25 23:59:59.999999999 +0300 +03
	// 25.09.2025 14:30
	// 48h0m0s
}

func ExampleSleepContext() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	err := timex.SleepContext(ctx, 5*time.Second)
	fmt.Println(errors.Is(err, context.DeadlineExceeded))

	done := make(chan struct{})
	close(done)
	fmt.Println(timex.SleepCtx(done, time.Second))
	fmt.Println(timex.SleepCtx(nil, time.Millisecond))
	// Output:
	// true
	// false
	// true
}
