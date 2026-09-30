package config_test

import (
	"fmt"
	"os"
	"time"

	"github.com/mustafacaglarkara/webdev/pkg/config"
)

func Example_env() {
	os.Setenv("APP_PORT", "8080")
	os.Setenv("APP_WORKERS", "dört")
	os.Setenv("APP_TIMEOUT", "5s")
	defer os.Unsetenv("APP_PORT")
	defer os.Unsetenv("APP_WORKERS")
	defer os.Unsetenv("APP_TIMEOUT")

	fmt.Println(config.GetEnv("APP_PORT", "3000"))
	fmt.Println(config.GetEnvInt("APP_WORKERS", 4)) // geçersiz -> sessizce fallback
	fmt.Println(config.GetEnvDuration("APP_TIMEOUT", time.Second))

	if _, _, err := config.LookupEnvInt("APP_WORKERS"); err != nil {
		fmt.Println(err)
	}
	if _, ok, err := config.LookupEnvInt("APP_YOK"); !ok && err == nil {
		fmt.Println("APP_YOK tanımlı değil")
	}
	if _, err := config.RequireEnv("APP_SECRET"); err != nil {
		fmt.Println(err)
	}
	// Output:
	// 8080
	// 4
	// 5s
	// config: APP_WORKERS="dört" geçersiz: strconv.Atoi: parsing "dört": invalid syntax
	// APP_YOK tanımlı değil
	// config: ortam değişkeni tanımlı değil: APP_SECRET
}

type Base struct {
	ID int `json:"id" validate:"required"`
}

type User struct {
	Base
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email"`
}

func ExampleGetTags() {
	fmt.Println(config.GetTags(User{}, "json"))
	fmt.Println(config.GetTag(User{}, "Name", "validate"))
	fmt.Println(config.GetTag(&User{}, "ID", "json"))
	// Output:
	// map[Email:email ID:id Name:name]
	// required,min=2 true
	// id true
}

func Example_yaml() {
	type Svc struct {
		Name string `yaml:"name"`
		Port int    `yaml:"port"`
	}
	s, _ := config.ToYAML(Svc{"çağrı-servisi", 8080})
	fmt.Print(s)
	v, err := config.FromYAML[Svc](s)
	fmt.Println(v, err)
	// Output:
	// name: çağrı-servisi
	// port: 8080
	// {çağrı-servisi 8080} <nil>
}
