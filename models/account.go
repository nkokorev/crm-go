package models

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	e "github.com/nkokorev/crm-go/errors"
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
func (account *Account) Create(owner *User) (err error) {

	fmt.Println("СОздаааем акккаунт!!!")
	if reflect.TypeOf(owner.ID).String() != "uint" {
		return errors.New("Can't create new account: user ID not specified")
	}

	account.HashID, err = u.CreateHashID(account)
	if err != nil {
		return err
	}

	// account Owner
	account.UserID = owner.ID

	err = base.GetDB().Create(account).Error
	if err != nil {
		return err
	}

	base.GetDB().Save(account)

	// Права не назначаем, т.к. он владелец аккаунта
	if account.AppendUser(owner) != nil {
		// вообще это фаталити, т.к. аккаунт создан, но пользователь не добавился в аккаунт, хотя права доступа к нему он имеет
		// пользователь не увидит созданный аккаунт, в списке доступных аккаунтов, следовательно не сможет зайти или удалить этот аккаунт
		if err := account.Delete(); err != nil {
			return errors.New("Can't append user in his account. And account user not deleted!")
		}
		return errors.New("Can't append user to your account")
	}

	if err := account.CreateAdminRole(); err != nil {
		return err
	}

	fmt.Println("Аккаунт создан!!!",account)
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

// Удаление аккаунта и всех связанных данных
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

// Добавление к аккаунту пользователя
func (account *Account) AppendUser(user *User) error {
	if err := base.GetDB().Model(&account).Association("Users").Append(user).Error; err != nil {
		return err
	}
	return nil
}

// исключение пользователя из аккаунта
func (account *Account) RemoveUser(user *User) error {

	if err := base.GetDB().Model(&account).Association("Users").Delete(user).Error; err != nil {
		return err
	}
	return nil
}




// ######### ниже не проверенные фукнции ###########
// Создает голую роль в аккаунте с разрешениями (Permission / []Permissions)
func (account *Account) CreateRole(role *Role,) error {
	return role.Create(account)
}

// создание ролей в контексте аккаунта
func (account *Account) CreateAdminRole(options_r... *Role) error {


	// 1. Создаем роль и привязываем ее к аккаунту
	role := &Role{}
	role.Name = "Администратор аккаунта"
	role.Description = "Полный доступ ко всем элементам и настройкам системы."
	if err := account.CreateRole(role); err != nil {
		return err
	}

	// 2. Назначаем роли все возможные права (full-assets)
	permissions := []Permission{}
	if err := base.GetDB().Find(&permissions).Error; err != nil {
		return err
	}

	role.AppendPermission(permissions)

	if len(options_r) > 0 {
		options_r[0] = role
	}
	return nil
}
func (account *Account) CreateManagerRole()  {

}
func (account *Account) CreateMarketerRole()  {

}

// Создает роль администратора
func (account Account) CreateAdminRoleOld(role *Role) (error u.Error) {

	// 1. Создаем роли и привязываем их к аккаунту
	role.Name = "Администратор аккаунта"
	role.Description = "Полный доступ ко всем элементам и настройкам системы."
	//role.AccountID = account.ID

	// todo добавить роль к аккаунту и назначить права администратора
	role.Create(&account)
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



