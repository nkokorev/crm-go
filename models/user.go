package models

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"time"
	//"unicode/utf8"
)

// global User
type User struct {
	ID        	uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID 		string `json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Username 	string `json:"username" gorm:"unique_index;not null;"`
	Email 		string `json:"email" gorm:"unique_index;not null;"`
	Name 		string `json:"name"`
	Surname 	string `json:"surname"`
	Patronymic 	string `json:"patronymic"`
	Password 	string `json:"-" gorm:"not null"` // json:"-"
	//Token string               `json:"token";sql:"-"`
	Accounts    []Account `json:"accounts" gorm:"many2many:account_users;"`
	AUsers 		[]AccountUser `json:"-"` // ??
	//Permissions         []Permission `json:"permissions"`
	//Permissions []Permission `json:"permissions" gorm:"many2many:user_permissions;"`

	//Permissions   []Permission `gorm:"polymorphic:Owner;"`
	CreatedAt *time.Time `json:"created_at;omitempty"`
	UpdatedAt *time.Time `json:"updated_at;omitempty"`

	DeletedAt *time.Time `sql:"index" json:"-"`
}

// specific name fixed user in the account
type AccountUser struct {
	ID        	uint `gorm:"primary_key;unique_index;" json:"-"`
	UserID 		uint // belong to user
	AccountID 	uint // belong to account
	Permissions []Permission `json:"permissions" gorm:"many2many:account_user_permissions;"`
	Roles   	[]Role `json:"roles" gorm:"many2many:account_user_roles;"`
	ApiKeys		[]ApiKey `json:"-"`
}

/*func (AccountUser) TableName() string {
	return "account_users"
}*/

//Create new User by &User{}. Dont hash pwd. HashID will be created.
func (user *User) Create() (error u.Error) {

	error = user.ValidateCreate()
	if error.HasErrors() {
		return
	}

	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		error.Message = t.Trans(t.UserFailedToCreate)
		return
	}
	user.Password = string(password)

	//user.HashID = u.RandStringBytes(u.LENGTH_HASH_ID)
	user.HashID, error = u.CreateHashID(user)
	if error.HasErrors() {
		error.Message = t.Trans(t.UserFailedToCreate)
		return
	}

	if err = base.GetDB().Create(user).Error; err != nil {
		fmt.Println(user)
		error.Message = t.Trans(t.UserFailedToCreate)
		return
	}

	base.GetDB().Save(user)


	return
}

// Full Validate incoming user details for create new user...
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
func (user *User) SoftDelete() (error u.Error) {

	if reflect.TypeOf(user.ID).String() != "uint" {
		error.Message = t.Trans(t.UserDeletionError)
		return
	}

	if err := base.GetDB().Delete(&user).Error; err != nil {
		error.Message = t.Trans(t.UserDeletionError)
		return
	}
	return
}

// Delete User from BD !!!. При удалении пользователя удаляются все связанными с ним данные.
func (user *User) Delete() (error u.Error) {

	if reflect.TypeOf(user.ID).String() != "uint" {
		error.Message = t.Trans(t.UserDeletionError)
		return
	}

	if err := base.GetDB().Unscoped().Delete(&user).Error; err != nil {
		error.Message = t.Trans(t.UserDeletionError)
		return
	}

	return
}

// добавляет роль пользователю
func (aUser AccountUser) AppendRole(role Role) (error u.Error) {
	err := base.GetDB().Model(&aUser).Association("Roles").Append(&role).Error
	if err != nil {
		error.Message = t.Trans(t.UserFailedAddRole)
		return
	}
	return
}

// удаляет роль у пользователя
func (aUser AccountUser) RemoveRole(role Role) (error u.Error) {
	err := base.GetDB().Model(&aUser).Association("Roles").Delete(&role).Error
	if err != nil {
		error.Message = t.Trans(t.UserFailedRemoveRole)
		return
	}
	return
}

// Вспомогательная функция получения пользователя ассоциированного с пользователем
func (aUser *AccountUser) LoadFromDB() (error u.Error) {
	//database.GetDB().Model(&auser).Where("account_id = ? AND user_id = ?", acc_id, user_id).First(&auser, "account_id = ? AND user_id = ?", acc_id, user_id)
	err := base.GetDB().Model(&aUser).First(&aUser, "account_id = ? AND user_id = ?", aUser.AccountID, aUser.UserID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			error.Message = "Record not found"
		} else {
			error.Message = "Connection error. Please retry"
		}
		return
	}
	return
}





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

// создание нового аккаунт см.: (account *Account) Create(user *User)
func (user *User) CreateAccount(account *Account) (error u.Error) {
	return account.Create(user)
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
