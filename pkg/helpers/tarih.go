package helpers

import "time"

func Now() time.Time                               { return time.Now() }
func FormatDate(t time.Time, layout string) string { return t.Format(layout) }
func DateDiff(a, b time.Time) time.Duration        { return a.Sub(b) }
func Timestamp() int64                             { return time.Now().Unix() }
