package helpers

import (
	"math"
	"math/rand"
	"strconv"
)

func ParseInt(s string) (int64, error)     { return strconv.ParseInt(s, 10, 64) }
func ParseFloat(s string) (float64, error) { return strconv.ParseFloat(s, 64) }
func RandomInt(min, max int) int           { return rand.Intn(max-min+1) + min }
func RoundFloat(val float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(val*factor) / factor
}
