package models

import (
	"github.com/dgrijalva/jwt-go"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
	e "github.com/nkokorev/crm-go/errors"
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

	base.GetDB().Save(account)

	// Права не назначаем, т.к. он владелец аккаунта
	if account.AppendUser(owner) != nil {
		// вообще это фаталити, т.к. аккаунт создан, но пользователь не добавился в аккаунт, хотя права доступа к нему он имеет
		// пользователь не увидит созданный аккаунт, в списке доступных аккаунтов, следовательно не сможет зайти или удалить этот аккаунт
		error.Message = "Can't append user to your account"
		if err := account.Delete(); err != nil {
			error.Message = "Can't append user in his account. And account user not deleted!"
		}
		return
	}
	return
}

// Мягкое удаление аккаунта (до 30 дней)
func (account *Account) SoftDelete() error {

	if reflect.TypeOf(account.ID).String() != "uint" {
		return e.AccountDeletionError
	}

	if err := base.GetDB().Delete(&account).Error; err != nil {
		return err
	}
	return nil
}

// Удаление аккаунта и всех связанных данных!!
func (account *Account) Delete() error {
	if reflect.TypeOf(account.ID).String() != "uint" {
		log.Println("Переданный аккаунт для удаления не содержит ID: ", account)
		return e.AccountDeletionError
	}

	if err := base.GetDB().Unscoped().Delete(&account).Error; err != nil {
		return err
	}
	return nil
}

// Add user to account
func (account *Account) AppendUser(user *User) error {
	if err := base.GetDB().Model(&account).Association("Users").Append(user).Error; err != nil {
		return err
	}
	return nil
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



