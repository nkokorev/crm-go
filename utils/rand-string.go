package utils

import (
	"crypto/aes"
	"math/rand"
	"time"
	"unicode/utf8"
	"unsafe"
)

//const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyz"
const letterBytesChar = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const letterBytesNum = "abcdefghijklmnopqrstuvwxyz1234567890"
const LENGTH_HASH_ID  = 8


func RandStringBytes(n int) string {
	return StringWithCharset(n, letterBytesNum)
}

func StringWithCharset(length int, charset string) string {

	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UTC().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}


func RandStringBytesMaskImprSrcUnsafe(n int, withNum bool) string {

	letterBytes := letterBytesChar

	if withNum {
		letterBytes = letterBytesNum
	}

	var src = rand.NewSource((time.Now().Add(time.Minute*268)).UnixNano())

	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// Считаем длину ключа 128 битным
func CreateAes128Key() (string, error) {
	str := RandStringBytesMaskImprSrcUnsafe(16, false)
	return str, ValidationAesKey(str)
}
func ValidationAesKey(key string) error {
	if len(key) != 16 {
		return Error{Message:"Некорректная длина ключа AES-128", Errors: map[string]interface{}{"aesKey":"Длина 128-битного ключа должна быть равна 16 символам в UTF-8"}}
	}

	// check utf-8
	r, size := utf8.DecodeRune([]byte(key))
	if r == utf8.RuneError || size < 1 {
		return Error{Message:"Некорректный ключ AES-128", Errors: map[string]interface{}{"aesKey":"Символы должны быть в кодировке UTF-8"}}
	}

	// пробуем создать новый AES Cipher
	_, err := aes.NewCipher([]byte(key))
	if err != nil {
		return Error{Message:"Некорректный ключ AES-128", Errors: map[string]interface{}{"aesKey":"Проверьте символы и кодировку UTF-8"}}
	}

	return nil
}
func CreateHS256Key() string {
	return RandStringBytesMaskImprSrcUnsafe(32, false)
}