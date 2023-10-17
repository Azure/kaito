// Can be used to
// Generate crypto-secured
// Auth token
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func generateToken() string {
	byteToken := make([]byte, 32) // Size can be adjusted
	rand.Read(byteToken)
	return hex.EncodeToString(byteToken)
}

func main() {
	fmt.Println(generateToken())
}
