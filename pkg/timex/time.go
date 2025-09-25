package timex

import (
	"time"
)

// Now returns current time (wrapper for time.Now)
func Now() time.Time { return time.Now() }

func FormatDate(t time.Time, layout string) string { return t.Format(layout) }
func DateDiff(a, b time.Time) time.Duration        { return a.Sub(b) }
func Timestamp() int64                             { return time.Now().Unix() }

func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
func EndOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}
func ParseTime(layout, value string) (time.Time, error) { return time.Parse(layout, value) }
func MustParseTime(layout, value string) time.Time {
	t, err := time.Parse(layout, value)
	if err != nil {
		panic(err)
	}
	return t
}

// SleepCtx waits for duration d or ctxDone channel; returns true if timer fired, false if cancelled.
func SleepCtx(ctxDone <-chan struct{}, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctxDone:
		return false
	case <-t.C:
		return true
	}
}
