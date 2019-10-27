package models

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	e "github.com/nkokorev/crm-go/errors"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"time"
)

type User struct {
	ID        	uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string `json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Username 	string `json:"username" gorm:"unique_index;not null;"`
	Email 		string `json:"email" gorm:"unique_index;not null;"`
	Name 		string `json:"name"`
	Surname 	string `json:"surname"`
	Patronymic 	string `json:"patronymic"`
	Password 	string `json:"-" gorm:"not null"` // json:"-"
	Accounts    []Account `json:"accounts" gorm:"many2many:account_users;"`
	AUsers 		[]AccountUser `json:"-"` // ??
	CreatedAt *time.Time `json:"created_at;omitempty"`
	UpdatedAt *time.Time `json:"updated_at;omitempty"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

func (user *User) Create() (err error) {

	// проверка на попытку создать дубль пользователя, который уже был создан
	if reflect.TypeOf(user.ID).String() == "uint" {
		if user.ID > 0 && !base.GetDB().First(&User{}, user.ID).RecordNotFound() {
			// todo need to translation
			return errors.New("Can't create new user: user with this ID already crated!")
		}
	}

	myErr := user.ValidateCreate()
	if myErr.HasErrors() {
		return e.UserFailedToCreate
	}

	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(password)

	user.HashID, err = u.CreateHashID(user)
	if err != nil {
		return err
	}

	if err = base.GetDB().Create(user).Error; err != nil {
		return err
	}

	if err = base.GetDB().Save(user).Error; err != nil {
		return  err
	}

	return nil
}

// Full Validate incoming user fields
func (user *User) ValidateCreate() (error u.Error) {

	err := u.VerifyEmail(user.Email, false)
	if err != nil {
		error.AddErrors("email", err.Error())
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	err = u.VerifyPassword(user.Password)
	if err != nil {
		error.AddErrors("email", err.Error())
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	err = u.VerifyUsername(user.Username)
	if err != nil {
		error.AddErrors("username", err.Error() )
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	if len([]rune(user.Name)) > 25 {
		error.AddErrors("name", t.Trans(t.UserInputIsTooLong) )
	}
	if len([]rune(user.Surname)) > 25 {
		error.AddErrors("surname", t.Trans(t.UserInputIsTooLong) )
	}
	if len([]rune(user.Patronymic)) > 25 {
		error.AddErrors("patronymic", t.Trans(t.UserInputIsTooLong) )
	}

	if error.HasErrors() {
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	//Email must be unique
	temp := &User{}

	//check for errors and duplicate emails
	err = base.GetDB().Unscoped().Model(&User{}).Where("email = ?", user.Email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		error.Message = t.Trans(t.UserFailedToCreate)
		return
	}
	if temp.Email != "" {
		error.AddErrors("email", t.Trans(t.EmailAlreadyUse) )
	}

	temp = &User{} // set to empty

	err = base.GetDB().Unscoped().Model(&User{}).Where("username = ?", user.Username).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		error.Message = t.Trans(t.UserFailedToCreate)
		return
	}
	if temp.Username != "" {
		error.AddErrors("username", t.Trans(t.UsernameAlreadyUse) )
	}

	if error.HasErrors() {
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
	}
	return
}

// SoftDelete user by user.ID.
func (user *User) SoftDelete() error {

	if reflect.TypeOf(user.ID).String() != "uint" {
		return e.UserDeletionErrorNotID
	}

	if err := base.GetDB().Delete(&user).Error; err != nil {
		return err
	}
	return nil
}

// Delete User from BD. При удалении пользователя удаляются все связанными с ним данные, кроме аккаунта(ов)!!!
func (user *User) Delete() error {

	if reflect.TypeOf(user.ID).String() != "uint" {
		return e.UserDeletionErrorNotID
	}

	// проверка на наличие связанных аккаунтов (стоит ограничение внешнего ключа)
	count := 0
	if err := base.GetDB().Model(&Account{}).Where("user_id = ?", user.ID).Count(&count).Error; err != nil {
		return e.UserDeletionErrorHasAccount
	}
	if count > 0 {
		return e.UserDeletionErrorHasAccount
	}

	if err := base.GetDB().Unscoped().Delete(&user).Error; err != nil {
		return err
	}

	return nil
}

// создание нового аккаунт в контексте пользователя
func (user *User) CreateAccount(account *Account) error {
	return account.Create(user)
}


/// #### ниже функции надо доработать

// Авторизует пользователя, в случае успеха возвращает jwt-token
func AuthLogin(username, password string) (cryptToken string, error u.Error) {

	user := &User{}

	err := base.GetDB().Where("username = ?", username).First(user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			error.AddErrors("username", t.Trans(t.UserNotFound) )
		} else {
			error.Message = "Connection error. Please retry"
		}
		return "", error
	}

	// если пользователь не найден temp.Username == nil, то пароль не будет искаться, т.к. он будет равен нулю (не с чем сравнивать)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
		error.AddErrors("password", t.Trans(t.UserPasswordIncorrect) )
	}

	if error.HasErrors() {
		error.Message = t.Trans(t.LoginInvalidCredentials)
		fmt.Println(error.Errors)
		return "", error
	}

	expiresAt := time.Now().Add(time.Minute * 10).Unix()
	claims := Token{
		user.ID,
		0,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AuthServer",
		},
	}
	cryptToken, err = claims.CreateToken()

	return cryptToken, error
}

func GetUserProfile(u uint) *User {

	user := &User{}

	base.GetDB().Preload("Accounts").Where("id = ?", u).First(&user)

	//db.Preload("Accounts", func(db *gorm.DB) *gorm.DB {
	//	return db.Table("accounts").Select("accounts.id, accounts.name, account_users.account_id, account_users.user_id")
	//}).Where("id = ?", u).First(&user)


	return user
}

func GetUser(u uint) *User {
	user := &User{}
	base.GetDB().First(&user, u)
	return user
}



func GetUsersAccount(u uint) (accounts []Account) {
	user := GetUser(u)
	base.GetDB().Model(&user).Related(&accounts,  "Accounts")
	return
}

// add User To Account
// remove User from Account
func AssociateUserToAccount(user *User, account *Account) error{
	base.GetDB().Model(&user).Association("Accounts").Replace(account)
	return nil
}
