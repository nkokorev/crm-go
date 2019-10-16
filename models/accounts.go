package models

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/nkokorev/crm-go/database"
	u "github.com/nkokorev/crm-go/utils"
	"reflect"
	"time"
)

//a struct to rep user account
type Account struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID string `json:"hash_id" gorm:"unique_index"`
	Name string `json:"name"` // СтанПроф / ООО ПК ВТВ-Инжинеринг / Ратус Медия / X6-Band (должно ли быть уникальным??)
	//Label string `json:"label"` // accountId : stan-prof / vtvent / ratus-media / x6-band
	UserID uint `json:"-"` // владелец
	Users         []User `json:"users" gorm:"many2many:account_users;"`
	Stores         []Store `json:"stores"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`
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
	db := database.GetDB()
	//db.GetDB().Model(&acc).Association("Accounts")
	//db.GetDB().Preload( "Accounts" ).First (&acc)

	//db.Model(&user).Related("Accounts")

	//db.Preload("Accounts").Where("id = ?", u).First(&user)

	db.Where("id = ?", u).First(account)

	return account
}

func GetAccountByHashID(hash string) *Account {

	account := &Account{}
	//acc := &Account{}
	db := database.GetDB()
	//db.GetDB().Model(&acc).Association("Accounts")
	//db.GetDB().Preload( "Accounts" ).First (&acc)

	//db.Model(&user).Related("Accounts")

	//db.Preload("Accounts").Where("id = ?", u).First(&user)

	db.Where("hash_id = ?", hash).First(account)

	return account
}

// Создает новый аккаунт с указанным именем и призывает его к владельцу
func (user *User) CreateAccount(name string) (hash string) {
	db := database.GetDB()
	hash = u.RandStringBytes(u.LENGTH_HASH_ID)
	fmt.Println( reflect.TypeOf(db.Create(&Account{Name: name, HashID: hash, UserID: user.ID})).String() )
	return
}


func AddUserToAccount(user *User, account *Account) {

	fmt.Println("Add User to Account: ", account.Name, "User name: ", user.Username)

	db := database.GetDB()
	db.Model(&user).Association("Accounts").Append(account)
}

func RemoveUserFromAccount(user *User, account *Account) {

	fmt.Println("Remove User from Account: ", account.Name, "User name: ", user.Username)

	db := database.GetDB()
	db.Model(&user).Association("Accounts").Delete(account)
}

