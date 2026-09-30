package conv_test

import (
	"fmt"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/conv"
)

func Example() {
	fmt.Println(conv.ToInt("42", 0), conv.ToInt(" 42 ", 0), conv.ToInt("abc", 5))
	fmt.Println(conv.ToInt64("123456789012", 0))
	fmt.Println(conv.ToFloat64("3.14", 0), conv.ToFloat64("3,14", -1))
	fmt.Println(conv.ToBool("true", false), conv.ToBool("evet", false))
	fmt.Println(conv.ToDuration("2h45m", time.Minute), conv.ToDuration("x", time.Minute))
	v, err := conv.ParseInt("12a")
	fmt.Println(v, err)
	fmt.Println(conv.RoundFloat(3.14159, 2), conv.RoundFloat(1234.5, -2), conv.RoundFloat(-2.5, 0))
	// Output:
	// 42 42 5
	// 123456789012
	// 3.14 -1
	// true false
	// 2h45m0s 1m0s
	// 0 strconv.ParseInt: parsing "12a": invalid syntax
	// 3.14 1200 -3
}

func ExampleSecureRandomInt() {
	n, err := conv.SecureRandomInt(100000, 999999) // 6 haneli OTP
	if err != nil {
		panic(err)
	}
	fmt.Println(n >= 100000 && n <= 999999)
	fmt.Println(conv.RandomInt(20, 10) >= 10) // uçlar yer değiştirir, panik yok
	// Output:
	// true
	// true
}
