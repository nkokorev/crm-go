package models

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	db "github.com/nkokorev/crm/database"
	"time"
)

//a struct to rep user account
type Account struct {
	DBModel
	Name string `json:"name" gorm:"unique_index"` // СтанПроф / ООО ПК ВТВ-Инжинеринг / Ратус Медия / X6-Band (должно ли быть уникальным??)
	//Label string `json:"label"` // accountId : stan-prof / vtvent / ratus-media / x6-band
	Users         []User `json:"users" gorm:"many2many:account_users;"`
	IPv4 string `json:"ipv4"`
}

// создает новый токен
var CreateAccountToken = func (userId, accountId uint) (cryptToken string, error error) {

	expiresAt := time.Now().Add(time.Minute * 45).Unix()
	claims := Token{
		userId,
		accountId,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AuthServer",
		},
	}
	cryptToken, error = claims.CreateToken()



	return
}

func GetAccount(u uint) *Account {

	account := &Account{}
	//acc := &Account{}
	db := db.GetDB()
	//db.GetDB().Model(&acc).Association("Accounts")
	//db.GetDB().Preload( "Accounts" ).First (&acc)

	//db.Model(&user).Related("Accounts")

	//db.Preload("Accounts").Where("id = ?", u).First(&user)

	db.Where("id = ?", u).First(account)

	return account
}


func AddUserToAccount(user *User, account *Account) {

	fmt.Println("Add User to Account: ", account.Name, "User name: ", user.Username)

	db := db.GetDB()
	db.Model(&user).Association("Accounts").Append(account)
}

func RemoveUserFromAccount(user *User, account *Account) {

	fmt.Println("Remove User from Account: ", account.Name, "User name: ", user.Username)

	db := db.GetDB()
	db.Model(&user).Association("Accounts").Delete(account)
}

