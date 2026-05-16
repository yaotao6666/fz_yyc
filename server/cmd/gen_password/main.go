package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 生成 bcrypt 哈希
	passwords := []string{"tm666666", "merchant123"}

	for _, pwd := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Error hashing password %s: %v\n", pwd, err)
			continue
		}
		fmt.Printf("Password: %s\n", pwd)
		fmt.Printf("Hash: %s\n\n", string(hash))
	}

	// 验证密码
	testCases := []struct {
		password string
		hash     string
	}{
		{"tm666666", "$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqJ7Q9A1vU7.F7.F7.F7.F7.F7.F7"},
		{"merchant123", "$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqJ7Q9A1vU7.F7.F7.F7.F7.F7.F7"},
	}

	fmt.Println("\nVerification test:")
	for _, tc := range testCases {
		err := bcrypt.CompareHashAndPassword([]byte(tc.hash), []byte(tc.password))
		if err != nil {
			fmt.Printf("Password %s: INVALID (err: %v)\n", tc.password, err)
		} else {
			fmt.Printf("Password %s: VALID\n", tc.password)
		}
	}
}
