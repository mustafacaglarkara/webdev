package conv

import (
	"math"
	"math/rand"
	"strconv"
	"time"
)

func ToInt(s string, def int) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return def
}
func ToInt64(s string, def int64) int64 {
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	return def
}
func ToFloat64(s string, def float64) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return def
}
func ToBool(s string, def bool) bool {
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	return def
}
func ToDuration(s string, def time.Duration) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	return def
}

func ParseInt(s string) (int64, error)     { return strconv.ParseInt(s, 10, 64) }
func ParseFloat(s string) (float64, error) { return strconv.ParseFloat(s, 64) }
func RandomInt(min, max int) int           { return rand.Intn(max-min+1) + min }
func RoundFloat(val float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(val*factor) / factor
}
