package utils

import (
	"math/rand"
	"time"
)

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyz"
const LENGTH_HASH_ID  = 8


func RandStringBytes(n int) string {
	return StringWithCharset(n, letterBytes)
}

func StringWithCharset(length int, charset string) string {

	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}