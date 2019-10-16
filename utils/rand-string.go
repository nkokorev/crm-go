package utils

import (
	"math/rand"
)

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyz"
const LENGTH_HASH_ID  = 8


func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	//fmt.Println(string(b))
	return string(b)
}
