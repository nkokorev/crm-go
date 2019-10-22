package models

import (
	"github.com/dgrijalva/jwt-go"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
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
	AUsers 		[]AccountUser `json:"-"`
	Users       []User `json:"users" gorm:"many2many:account_users;"`
	ApiKeys		[]ApiKey `json:"-"`
	Roles 		[]Role	`json:"roles"`
	Stores      []Store `json:"stores"`
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

	account.HashID, error = u.CreateHashID(account)
	if error.HasErrors() {
		error.Message = t.Trans(t.AccountFailedToCreate)
		return
	}

	// account Owner
	account.UserID = owner.ID

	err := base.GetDB().Create(account).Error
	if err != nil {
		error.Message = t.Trans(t.AccountFailedToCreate)
		return
	}

	// user.AddToAccount()
	//roles := []Role{} // owner
	// todo неплохо бы давать владельцу права администратора по-умолчанию
	error = account.AppendUser(owner)

	return
}

// Мягкое удаление аккаунта (до 30 дней)
func (account *Account) SoftDelete() (error u.Error) {

	if reflect.TypeOf(account.ID).String() != "uint" {
		error.Message = t.Trans(t.AccountDeletionError)
		return
	}

	if err := base.GetDB().Delete(&account).Error; err != nil {
		error.Message = "Cant Soft Delete account"
		return
	}
	return
}

// Удаление аккаунта и всех связанных данных!!
func (account *Account) Delete() (error u.Error) {

	if reflect.TypeOf(account.ID).String() != "uint" {
		error.Message = t.Trans(t.AccountDeletionError)
		return
	}

	if err := base.GetDB().Unscoped().Delete(&account).Error; err != nil {
		error.Message = "Cant Delete account"
		return
	}
	return
}

// Add user to account
func (account *Account) AppendUser(user *User) (error u.Error) {

	if err := base.GetDB().Model(&account).Association("Users").Append(user).Error; err != nil {
		error.Message = "Cant add User to this Account"
		return
	}
	return
}

// исключение пользователя из аккаунта
func (account *Account) RemoveUser(user *User) (error u.Error) {

	if err := base.GetDB().Model(&account).Association("Users").Delete(user).Error; err != nil {
		error.Message = "Cant Remove User from Account"
		return
	}
	return
}

// ######### ниже не проверенные фукнции ###########
// todo надо доработать: передавать роль по ссылке или нет !!!
// Создает голую роль в аккаунте с разрешениями (Permission / []Permissions)
func (account Account) CreateRole(role *Role, values... interface{}) (error u.Error){
	return role.Create(account, values...)
}

// Создает роль администратора
func (account Account) CreateAdminRole(role *Role) (error u.Error) {

	// 1. Создаем роли и привязываем их к аккаунту
	role.Name = "Администратор аккаунта"
	role.Description = "Полный доступ ко всем элементам и настройкам системы."
	//role.AccountID = account.ID

	// todo добавить роль к аккаунту и назначить права администратора
	role.Create(account)
	role.AppendAdminPermissions()

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
	db := base.GetDB()
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
	db := base.GetDB()
	//db.GetDB().Model(&acc).Association("Accounts")
	//db.GetDB().Preload( "Accounts" ).First (&acc)

	//db.Model(&user).Related("Accounts")

	//db.Preload("Accounts").Where("id = ?", u).First(&user)

	db.Where("hash_id = ?", hash).First(account)

	return account
}



