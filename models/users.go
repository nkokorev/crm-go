package models

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	u "github.com/nkokorev/crm/http/utils"
	t "github.com/nkokorev/crm/locales"
	db "github.com/nkokorev/crm/database"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type DBModel struct {
	ID        uint `gorm:"primary_key" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

type User struct {
	DBModel
	Username string `json:"username" gorm:"unique_index"`
	Email string `json:"email" gorm:"unique_index"`
	Name string `json:"name"`
	Surname string `json:"surname"`
	Patronymic string `json:"patronymic"`
	Password string `json:"-"` // json:"-"
	//Token string               `json:"token";sql:"-"`
	Accounts         []Account `json:"accounts" gorm:"many2many:account_users;"`
}

// Авторизует пользователя, в случае успеха возвращает jwt-token
func AuthLogin(username, password string) (cryptToken string, error Error) {

	user := &User{}

	err := db.GetDB().Where("username = ?", username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			error.AddErrors("username", t.Trans(t.UserNotFound) )
		} else {
			error.Message = "Connection error. Please retry"
		}
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

	db.GetDB().Preload("Accounts").Where("id = ?", u).First(&user)

	//db.Preload("Accounts", func(db *gorm.DB) *gorm.DB {
	//	return db.Table("accounts").Select("accounts.id, accounts.name, account_users.account_id, account_users.user_id")
	//}).Where("id = ?", u).First(&user)


	return user
}

func GetUser(u uint) *User {
	user := &User{}
	db.GetDB().First(&user, u)
	return user
}

//Validate incoming user details...
func (user *User) Validate() (map[string] interface{}, bool) {

	errors := make(map[string]string)

	if !strings.Contains(user.Email, "@") {
		return u.Message(false, "Email address is required"), false
	}

	if len(user.Password) < 6 {
		return u.Message(false, "Password is required"), false
	}

	//Email must be unique
	temp := &User{}

	//check for errors and duplicate emails
	//err := db.GetDB().Table("users").Where("email = ?", user.Email).First(temp).Error

	err := db.GetDB().Table("users").Where("email = ?", user.Email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error. Please retry"), false
	}
	if temp.Email != "" {
		//t.GetBundle()
		errors["email"] = "Username already in use by another user"
		//return u.Message(false, "Email address already in use by another user."), false
	}

	//check for errors and duplicate username
	err = db.GetDB().Table("users").Where("username = ?", user.Username).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return u.Message(false, "Connection error. Please retry"), false
	}
	if temp.Username != "" {
		errors["username"] = "Username already in use by another user"
	}

	if len(errors) > 0 {
		return u.MessageWithErrors("Ошибки заполнения формы", errors), false
	}
	return u.Message(false, "Requirement passed"), true
}

func (user *User) Create() (map[string] interface{}) {

	if resp, ok := user.Validate(); !ok {
		return resp
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	user.Password = string(hashedPassword)

	db.GetDB().Create(user)

	if user.ID <= 0 {
		return u.Message(false, "Failed to create user, connection error.")
	}

	fmt.Println("User create") // заменить потом на логи

	user.Password = "" //delete password

	response := u.Message(true, "User has been created")
	response["user"] = user
	return response
}



func UserLoginOld(username, password string) (map[string]interface{}) {

	myErrors := make(map[string]string)

	user := &User{}

	err := db.GetDB().Where("username = ?", username).First(&user).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			myErrors["username"] = t.Trans(t.UserNotFound)
		} else {
			return u.Message(false, "Connection error. Please retry")
		}
	}

	// если пользователь не найден temp.Username == nil, то пароль не будет искаться, т.к. он будет равен нулю (не с чем сравнивать)
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
		myErrors["password"] = t.Trans(t.UserPasswordIncorrect)
	}

	if len(myErrors) > 0 {
		return u.MessageWithErrors(t.Trans(t.LoginInvalidCredentials), myErrors)
	}

	expiresAt := time.Now().Add(time.Minute * 100000).Unix()
	claims := Token{
		user.ID,
		0,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AuthServer",
		},
	}
	cryptToken, err := claims.CreateToken()

	resp := u.Message(true, "Logged In")
	resp["token"] = cryptToken
	return resp
}



func GetUsersAccount(u uint) (accounts []Account) {

	//var accounts []Account
	user := GetUser(u)

	db.GetDB().Model(&user).Related(&accounts,  "Accounts")
	return
}

// add User To Account
// remove User from Account
func AssociateUserToAccount(user *User, account *Account) error{
	db.GetDB().Model(&user).Association("Accounts").Replace(account)
	return nil
}
