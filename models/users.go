package models

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type User struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	Username string `json:"username" gorm:"unique_index"`
	Email string `json:"email" gorm:"unique_index"`
	Name string `json:"name"`
	Surname string `json:"surname"`
	Patronymic string `json:"patronymic"`
	Password string `json:"-"` // json:"-"
	//Token string               `json:"token";sql:"-"`
	Accounts         []Account `json:"accounts" gorm:"many2many:account_users;"`
	//Permissions         []Permission `json:"permissions"`
	//Permissions []Permission `json:"permissions" gorm:"many2many:user_permissions;"`
	Permissions   []Permission `gorm:"polymorphic:Owner;"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

// Авторизует пользователя, в случае успеха возвращает jwt-token
func AuthLogin(username, password string) (cryptToken string, error u.Error) {

	user := &User{}

	err := database.GetDB().Where("username = ?", username).First(&user).Error
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

	database.GetDB().Preload("Accounts").Where("id = ?", u).First(&user)

	//db.Preload("Accounts", func(db *gorm.DB) *gorm.DB {
	//	return db.Table("accounts").Select("accounts.id, accounts.name, account_users.account_id, account_users.user_id")
	//}).Where("id = ?", u).First(&user)


	return user
}

func GetUser(u uint) *User {
	user := &User{}
	database.GetDB().First(&user, u)
	return user
}

// создает пользователя на основе переданной структуры *User. Пароль должен быть в явном виде, чтобы его можно было провалидировать.
func (user *User) Create() (error u.Error) {

	error = user.Validate()
	if error.HasErrors() {
		return
	}

	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		error.Message = t.Trans(t.UserFailedToCreate)
		return
	}
	user.Password = string(password)

	err = database.GetDB().Create(user).Error
	if err != nil {
		error.Message = t.Trans(t.UserFailedToCreate)
	}

	return
}

//Validate incoming user details...
func (user *User) Validate() (error u.Error) {

	error = u.VerifyEmail(user.Email, true)
	if error.HasErrors() {
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	error = u.VerifyPassword(user.Password)
	if error.HasErrors() {
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	if len(user.Username) > 25 {
		error.AddErrors("username", t.Trans(t.UserInputIsTooLong) )
	}
	if len(user.Name) > 25 {
		error.AddErrors("name", t.Trans(t.UserInputIsTooLong) )
	}
	if len(user.Surname) > 25 {
		error.AddErrors("surname", t.Trans(t.UserInputIsTooLong) )
	}
	if len(user.Patronymic) > 25 {
		error.AddErrors("patronymic", t.Trans(t.UserInputIsTooLong) )
	}

	if error.HasErrors() {
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	//Email must be unique
	temp := &User{}

	//check for errors and duplicate emails
	err := database.GetDB().Table("users").Where("email = ?", user.Email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		error.Message = t.Trans(t.UserFailedToCreate)
		return
	}
	if temp.Email != "" {
		error.AddErrors("email", t.Trans(t.EmailAlreadyUse) )
	}

	temp = &User{} // set to empty

	err = database.GetDB().Table("users").Where("username = ?", user.Username).First(temp).Error
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





func GetUsersAccount(u uint) (accounts []Account) {
	user := GetUser(u)
	database.GetDB().Model(&user).Related(&accounts,  "Accounts")
	return
}

// add User To Account
// remove User from Account
func AssociateUserToAccount(user *User, account *Account) error{
	database.GetDB().Model(&user).Association("Accounts").Replace(account)
	return nil
}
