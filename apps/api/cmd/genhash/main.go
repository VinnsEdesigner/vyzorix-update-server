package main

import (
	"fmt"
	password "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/password"
)

func main() {
	h, err := password.HashPassword("TestBot@2026")
	if err != nil {
		panic(err)
	}
	fmt.Println(h)
}
