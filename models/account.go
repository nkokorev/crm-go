package models

import (
	"errors"
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

	if reflect.TypeOf(owner.ID).String() != "uint" {
		return errors.New("Can't create new account: user ID not specified")
	}

	account.HashID, err = u.CreateHashID(account)
	if err != nil {
		return err
	}

	// set ID account Owner. Нужна обработка удаления пользователя и его аккаунтов
	account.UserID = owner.ID
	if err = base.GetDB().Create(account).Error; err != nil {
		return err
	}
	if err = base.GetDB().Save(account).Error; err != nil {
		return err
	}

	// Добавляем пользователя в созданный им аккаунт
	if account.AppendUser(owner) != nil {
		// вообще это фаталити, т.к. аккаунт создан, но пользователь не добавился в аккаунт, хотя права доступа к нему он имеет
		// пользователь не увидит созданный аккаунт, в списке доступных аккаунтов, следовательно не сможет зайти или удалить этот аккаунт
		if err = account.Delete(); err != nil {
			return errors.New("Can't append user in his account. And account user not deleted!")
		}
		return errors.New("Can't append user to your account")
	}

	// назначаем роль владельца владельцу аккаунта
	aUser := AccountUser{}
	if err = aUser.GetAccountUser(owner.ID, account.ID); err != nil {
		// это фаталити, бесхозный акк, нужно его удалить
		if err = account.Delete(); err != nil {
			return err
		}
		return err
	}
	if err = aUser.SetRoleOwner(); err != nil {
		// это фаталити, бесхозный акк, нужно его удалить
		if err = account.Delete(); err != nil {
			return err
		}
		return err
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

// Удаление аккаунта и всех связанных данных
func (account *Account) Delete() error {
	if reflect.TypeOf(account.ID).String() != "uint" {
		log.Println("Переданный аккаунт для удаления не содержит ID: ", account)
		return e.AccountDeletionError
	}

	// удаляем вручную все не системные роли
	if err := account.RemoveAllRoles(); err != nil {
		return err
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

// удаление всех ролей, не являющимися системными
func (account *Account) RemoveAllRoles() error {
	roles := []Role{}
	if err := base.GetDB().Delete(&roles,"account_id = ? AND system = FALSE", account.ID).Error; err != nil {
		return err
	}
	return nil
}


// ######### ниже не проверенные фукнции ###########
// Создает голую роль в аккаунте с разрешениями (Permission / []Permissions)
func (account *Account) CreateRole(role *Role) error {
	if reflect.TypeOf(account.ID).String() != "uint" {
		return errors.New("Cant create role, not found account ID")
	}
	role.AccountID = account.ID
	return role.Create()
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



