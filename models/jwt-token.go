package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io"
	"log"
)

//Token struct declaration
type JWT struct {
	UserID uint // if found > 0
	AccountID uint // if activated > 0
	SignedAccountID uint // ID of main account
	//Username string
	//Email string
	jwt.StandardClaims

	User User
	Account Account
}

func (claims JWT) CreateCryptoToken() (cryptToken string, err error) {

	if err := claims.UploadRelatedData();err != nil {
		return "", err
	}

	//Create JWT token
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)
	//tokenString, err := token.SignedString([]byte(os.Getenv("jwt_key")))
	tokenString, err := token.SignedString([]byte(claims.Account.UiApiJwtKey))
	if err != nil {
		return
	}

	// Encode jwt-token
	//cryptToken, err = JWT{}.encrypt([]byte(os.Getenv("aes_key")), tokenString)
	cryptToken, err = JWT{}.encrypt([]byte(claims.Account.UiApiAesKey), tokenString)
	if err != nil {
		return
	}
	return
}

// AES кодирование по ключу key[]
func (JWT) encrypt(key []byte, message string) (encmess string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	//returns to base64 encoded string
	encmess = base64.URLEncoding.EncodeToString(cipherText)
	return
}

// AES декодирование по ключу key[]
func (JWT) decrypt(key []byte, securemess string) (decodedmess string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short!")
		return
	}

	//IV needs to be unique, but doesn't have to be secure.
	//It's common to put it at the beginning of the ciphertext.
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedmess = string(cipherText)
	return
}

// декодирует token по внутреннему ключу
func (claims JWT) DecryptToken(token string) (tk string, err error) {
	if err := claims.UploadRelatedData();err != nil {
		return "", err
	}
	//tk, err = JWT{}.decrypt( []byte(os.Getenv("aes_key")), token)
	tk, err = JWT{}.decrypt( []byte(claims.Account.UiApiAesKey), token)
	return
}

func (tk *JWT) ParseToken(decryptedToken string) (err error) {

	if err := tk.UploadRelatedData();err != nil {
		return err
	}

	// получаем библиотечный токен
	token, err := jwt.ParseWithClaims(decryptedToken, tk, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("JWT: Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		//return []byte(os.Getenv("jwt_key")), nil
		return []byte(tk.Account.UiApiJwtKey), nil
	})
	if err != nil {
		log.Println("JWT error: ", err)
		return
	}

	if !token.Valid {
		log.Println("JWT error: Token is not valid")
		return
	}

	return
}

func (tk *JWT) ParseAndDecryptToken(cryptToken string) error {

	decryptedToken, err := JWT{}.DecryptToken(cryptToken);
	if err != nil {
		return err
	}

	err = tk.ParseToken(decryptedToken)
	if err != nil {
		return err
	}
	return err

}

func (tk *JWT) UploadRelatedData() error {
	// Получаем настройки аккаунта
	if tk.AccountID < 1 {
		return errors.New("Не верно указан аккаунт")
	}

	if tk.UserID < 1 {
		return errors.New("Не верно указан пользователь")
	}

	if err := db.First(&tk.Account, tk.AccountID).Error; err != nil {
		return errors.New("Не удалось найти аккаунт для создания крипто ключа")
	}

	if err := db.First(&tk.User, tk.UserID).Error; err != nil {
		return errors.New("Не удалось найти пользователя для создания крипто ключа")
	}

	return nil

}
