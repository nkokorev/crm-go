package models

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/nkokorev/crm-go/database"
	u "github.com/nkokorev/crm-go/utils"
	_ "github.com/nkokorev/auth-server/locales"
	t "github.com/nkokorev/crm-go/locales"
	"log"
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

// Создает новый аккаунт по структуре account. Аккаунт привязывается к владельцу по user_id, добавляя его в список пользователей аккаунта.
func (account *Account) Create(owner *User) (error u.Error) {

	if reflect.TypeOf(owner.ID).String() != "uint" {
		log.Println("Невозможно создать аккаунт, у пользователя не задан ID")
		error.Message = "Can't create new account: user ID not specified"
		return
	}

	account.HashID, error = CreateHashID(account)
	if error.HasErrors() {
		error.Message = t.Trans(t.AccountFailedToCreate)
		return
	}

	// account Owner
	account.UserID = owner.ID

	err := database.GetDB().Create(account).Error
	if err != nil {
		error.Message = t.Trans(t.AccountFailedToCreate)
		return
	}

	// user.AddToAccount()
	roles := []Role{} // owner
	account.AddUser(owner, roles)

	return
}

// Мягкое удаление аккаунта (до 30 дней)
func (account *Account) SoftDelete() (error u.Error) {

	if reflect.TypeOf(account.ID).String() != "uint" {
		error.Message = t.Trans(t.AccountDeletionError)
		return
	}

	database.GetDB().Delete(&account)
	return
}

// Удаление аккаунта и всех связанных данных!!
func (account *Account) Delete() (error u.Error) {

	if reflect.TypeOf(account.ID).String() != "uint" {
		error.Message = t.Trans(t.AccountDeletionError)
		return
	}

	database.GetDB().Unscoped().Delete(&account)
	return
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


// Add user to account with roles
func (acc *Account) AddUser(user *User, roles... []Role) (error u.Error) {

	database.GetDB().Model(&user).Association("Accounts").Append(acc)
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



