package models

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	u "github.com/nkokorev/crm-go/utils"
	"golang.org/x/crypto/bcrypt"
	"os"
	"regexp"
	"unicode"

	"time"
)

type User struct {
	ID        	uint `json:"id" gorm:"primary_key"`
	SignedAccountID uint `json:"signedAccountId" gorm:"index;not null"`
	
	Username 	string `json:"username" gorm:"type:varchar(255);unique_index;default:null;"`
	Email 		string `json:"email" gorm:"type:varchar(255);unique_index;default:null;"`
	MobilePhone	string `json:"mobilePhone" gorm:"type:varchar(255);unique_index;default:null;"` // нужно проработать формат данных
	Password 	string `json:"-" gorm:"type:varchar(255);default:null;"` // json:"-"

	Name 		string `json:"name" gorm:"type:varchar(255)"`
	Surname 	string `json:"surname" gorm:"type:varchar(255)"`
	Patronymic 	string `json:"patronymic" gorm:"type:varchar(255)"`

	//Role 		string `json:"role" gorm:"type:varchar(255);default:'client'"`

	DefaultAccountID uint `json:"defaultAccountId" gorm:"default:NULL"` // указывает какой аккаунт по дефолту загружать
	InvitedUserID uint `json:"-" gorm:"default:NULL"` // указывает какой аккаунт по дефолту загружать

	EmailVerifiedAt *time.Time `json:"emailVerifiedAt" gorm:"default:null"`
	PasswordReset bool `json:"passwordReset" gorm:"default:FALSE"`
	//EmailVerification bool `json:"email_verification" gorm:"default:false"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"-" sql:"index"`

	//Profile UserProfile `json:"profile" gorm:"preload"`

	Accounts []Account `json:"-" gorm:"many2many:account_users;preload"`
}

func (user User) create () (*User, error) {

	var outUser User

	// Проверим другие входящие данные пользователя
	if err := user.ValidateCreate(); err != nil {
		return nil, err
	}

	// 6. Создаем крипто пароль
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user.Password = string(password)

	// копируем разрешеныне данные
	outUser.SignedAccountID = user.SignedAccountID
	outUser.Username = user.Username
	outUser.Email = user.Email
	outUser.MobilePhone = user.MobilePhone

	outUser.Name = user.Name
	outUser.Surname = user.Surname
	outUser.Patronymic = user.Patronymic

	outUser.DefaultAccountID = user.DefaultAccountID
	outUser.InvitedUserID = user.InvitedUserID

	if err := db.Create(&outUser).Error; err != nil {
		return nil, err
	}

	return &outUser, nil
}




// осуществляет поиск по ID
func GetUserById (userId uint) (user *User, err error) {

	err = db.Model(&User{}).Find(user, userId).Error

	return user, err
}

func (user *User) Get () error {
	/*return db.Preload("Accounts", func(db *gorm.DB) *gorm.DB {
		return db.Order(("accaunts.id DESC"))
	}).Find(user).Error*/

	//return db.Preload("Account").First(user,user.ID).Error
	//db.Set("gorm:auto_preload", true)
	return db.Preload("Accounts").First(user).Error
}

// осуществляет поиск по email
func (user *User) GetByEmail () error {
	return db.First(user,"email = ?", user.Email).Error
}

// осуществляет поиск по имени пользователя
func (user *User) GetByUsername () error {
	return db.First(user,"username = ?", user.Username).Error
}

// сохраняет как новую модель... лучше вообще убрать этот метод
func (user *User) Save () error {
	//return db.Model(user).Omit("id", "deleted_at", "created_at", "updated_at").Save(user).Find(user, "id = ?", user.ID).Error
	return db.Model(user).Omit("id", "deleted_at", "created_at", "updated_at").Save(user).First(user, "id = ?", user.ID).Error
}

// обновляет указанные данные и сохраняет в текущую модель в БД
func (user *User) Update (input interface{}) error {
	return db.Model(user).Where("id = ?", user.ID).Omit("id", "username", "created_at", "updated_at", "deleted_at").Update(input).First(user).Error
}

// удаляет пользователя по ID
func (user *User) Delete () error {
	return db.Model(User{}).Where("id = ?", user.ID).Delete(user).Error
}


// ### HELPERS FUNC ###

// Существует ли пользователь с указанным ID
func (User) Exist(id uint) bool {
	return !db.Unscoped().First(&User{}, "ip = ?", id).RecordNotFound()
}

func (User) ExistEmail(email string) bool {
	return !db.Unscoped().First(&User{},"email = ?", email).RecordNotFound()
}

func (User) ExistUsername(username string) bool {
	return !db.Unscoped().First(&User{},"username = ?", username).RecordNotFound()
}


// ### Validate & Verify FUNC ###

func (User) VerifyPassword(pwd string) error {

	if len([]rune(pwd)) == 0 {
		return errors.New("Поле необходимо заполнить")
	}
	if len([]rune(pwd)) < 6 {
		return errors.New("Слишком короткий пароль")
	}

	if len([]rune(pwd)) > 25 {
		return errors.New("Слишком длинный пароль")
	}

	letters := 0
	var number, upper, special bool
	for _, c := range pwd {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
			letters++
		case unicode.IsPunct(c) || unicode.IsSymbol(c):
			special = true
		case unicode.IsLetter(c) || c == ' ':
			letters++
		default:
			//return false, false, false, false
		}
	}

	if ! (number && upper && special && letters >= 5) {
		return errors.New("Слишком простой пароль: проверьте наличие обязательных символов.")
	}

	return nil
}

// проверяет имя пользователя на соответствие правилам. Не проверяет уникальность
func (User) VerifyUsername(username string) error {

	if len(username) == 0 {
		return errors.New("Поле необходимо заполнить")
	}
	if len(username) < 3 {
		return errors.New("Имя пользователя слишком короткое")
	}
	if len(username) > 25 || len([]rune(username)) > 25 {
		return errors.New("Имя пользователя слишком длинное")
	}

	var rxUsername = regexp.MustCompile("^[a-zA-Z0-9,\\-_]+$")

	if !rxUsername.MatchString(username) {
		return errors.New("Используйте только a-z,A-Z,0-9 а также символ -")
	}

	if (User{}).ExistUsername(username) {
		return errors.New("Данный username уже используется")
	}

	return nil
}

// Проверка email для нового пользователя
func (User) VerifyEmail(email string) error {

	if err := u.VerifyEmail(email, !(os.Getenv("http_dev") == "true") ); err != nil {
		return err
	}

	if (User{}).ExistEmail(email) {
		return errors.New("Данный email-адрес уже используется")
	}

	return nil
}

// Проверка входящих полей
func (user User) ValidateCreate() error {

	var e u.Error

	// 1. Проверка email
	if err := user.VerifyEmail(user.Email); err != nil {
		e.AddErrors("email", err.Error())
		e.Message = "Проверьте правильность заполнения формы"
		return e
	}

	// 2. Проверка username пользователя
	if err := user.VerifyUsername(user.Username); err != nil {
		e.AddErrors("username", err.Error())
		e.Message = "Проверьте правильность заполнения формы"
		return e
	}

	// 3. Проверка password
	if err := user.VerifyPassword(user.Password); err != nil {
		e.AddErrors("password", err.Error())
		e.Message = "Проверьте правильность заполнения формы"
		return e
	}

	// 4. Проверка дополнительных полей на длину
	if len([]rune(user.Name)) > 25 {
		e.AddErrors("name", "Поле слишком длинное" )
	}
	if len([]rune(user.Surname)) > 25 {
		e.AddErrors("surname", "Поле слишком длинное")
	}
	if len([]rune(user.Patronymic)) > 25 {
		e.AddErrors("patronymic", "Поле слишком длинное" )
	}

	// проверка итоговых ошибок
	if e.HasErrors() {
		e.Message = "Проверьте правильность заполнения формы"
		return e
	}

	return nil
}

// Проверяет и отправляет email подтвеждение
func (user *User) SendEmailVerification() error {

	// 1. Создаем токен
	emailToken := &EmailAccessToken{}
	if err := emailToken.CreateInviteVerificationToken(user); err != nil {
		return err
	}

	// 2. Отправляем письмо
	return  emailToken.SendMail()
}

// Отправляет имя пользователя на его почту
func (user *User) SendEmailRecoveryUsername() error {

	// тут должна быть какая-то задержка...
	time.Sleep(time.Second * 1)
	// собственно тут простая отправка письма пользователю с его именем
	return  nil
}

// Проверяет пароль пользователя
func (user User) ComparePassword(password string) bool {
	// если пользователь не найден temp.Username == nil, то пароль не будет искаться, т.к. он будет равен нулю (не с чем сравнивать)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return false
	}

	return true
}

// Отправляет ссылку для сброса пароля пользователя на его почту и создает токен для сброса
func (user *User) RecoveryPasswordSendEmail() error {

	// 1. Создаем токен для сброса пароля
	emailToken := &EmailAccessToken{};
	if err := emailToken.CreateResetPasswordToken(user); err != nil {
		return err
	}

	return  emailToken.SendMail()
}

// сбрасывает пароль пользователя
func (user *User) ResetPassword() error {
	user.Password = ""
	user.PasswordReset = true
	return user.Save()
}

// устанавливает новый пароль
func (user *User) SetPassword(passwordNew, passwordOld string) error {

	// 1. Проверяем старый пароль, при условии, что пароль не сброшен
	if !user.PasswordReset && !user.ComparePassword(passwordOld) {
		return u.Error{Message:"Не верно указан старый пароль"}
	}


	// 2. Устанавливаем новый крипто пароль
	password, err := bcrypt.GenerateFromPassword([]byte(passwordNew), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(password)

	// 3. Ставим флаг passwordReset = false
	user.PasswordReset = false

	// 4. Сохраняем данные пользователя
	if err := user.Save();err!=nil {
		return err
	}

	// 5. Очищаем таблицу от токенов по сбросу пароля (если есть)
	(EmailAccessToken{}).UserDeletePasswordReset(user)

	return nil
}

// ### Account's FUNC ###

func (user *User) LoadAccounts() error {
	return db.Preload("Accounts").First(&user).Error
}

// проверяет доступ и возвращает данные аккаунта
func (user *User) GetAccount(account *Account) error {
	return db.Model(user).Where("user_id = ?", user.ID).Association("Accounts").Find(account).Error
}

// Только пользователь RatusCRM может создавать новые аккаунты
func (user User) CreateAccount(input Account) (*Account,error) {

	// Проверяем пользователя и его роль в RatusCRM.

	// 1. Создаем аккаунт
	a, err := input.create()
	if err != nil {
		return nil, err
	}

	// 2. Привязываем аккаунт к пользователю
	if err := a.AppendUser(&user); err != nil {
		return nil, err
	}

	// 3. Назначает роль owner


	return a, nil
}

// функция прокладка, обновление можно вызвать и из интерфейса аккаунта
func (user *User) UpdateAccount(a *Account, input interface{}) error {
	return a.Update(input)
}

// удаляет аккаунт, если пользователь имеет такие права
func (user *User) DeleteAccount(a *Account) error {

	// 1. Проверяем доступ "= проверяем права на удаление аккаунта"
	if db.Model(user).Where("account_id = ?", a.ID).Association("Accounts").Count() == 0 {
		return errors.New("Account not exist or your have't permissions for this account")
	}

	// 2. Привязываем аккаунт к пользователю. Реально удаляем через месяц.
	if err := a.SoftDelete(); err != nil {
		return err
	}

	return nil
}


/// ### Auth FUNC ###

// Авторизует пользователя: загружает пользователя с предзагрузкой аккаунтов и, в случае успеха возвращает jwt-token
func (user *User) AuthLogin(username, password string, onceLogin_opt... bool) (string, error) {

	var e u.Error

	// Делаем предзагрузку аккаунтов, чтобы потом их еще раз не подгружать
	if err := db.Preload("Accounts").Where("username = ?", username).First(user).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			e.AddErrors("username", "Пользователь не найден")
			e.Message = "Не верно указаны данные"
		} else {
			e.Message = "Внутренняя ошибка. Попробуйте позже"
		}
		return "", e
	}

	// если пользователь не найден temp.Username == nil, то пароль не будет искаться, т.к. он будет равен нулю (не с чем сравнивать)
	if !user.ComparePassword(password) {
		e.AddErrors("password", "Неверный пароль")
	}
	/*err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
		e.AddErrors("password", "Не верный пароль")
	}*/

	if e.HasErrors() {
		e.Message = "Не верно указаны данные"
		return "", e
	}

	expiresAt := time.Now().UTC().Add(time.Minute * 20).Unix()
	claims := JWT{
		user.ID,
		0,
		user.SignedAccountID,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AuthServer",
		},
	}
	return claims.CreateCryptoToken()
}

// создает короткий jwt-токен для пользователя. Весьма опасная фукнция
func (user *User) CreateJWTToken() (string, error) {

	// Делаем предзагрузку аккаунтов, чтобы потом их еще раз не подгружать
	if err := db.Preload("Accounts").First(user).Error; err != nil {
		return "", err
	}

	expiresAt := time.Now().UTC().Add(time.Minute * 20).Unix()
	claims := JWT{
		user.ID,
		0,
		user.SignedAccountID,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AuthServer",
		},
	}
	return claims.CreateCryptoToken()

}

// Вход в аккаунт из-под пользователя. Делает необходимые проверки и возвращает новый токен/
func (user *User) LoginInAccount(account_id uint) (string, error) {

	var e u.Error

	if user.ID < 1 || account_id < 1 {
		e.Message = "Внутренняя ошибка во входе в аккаунт"
		return "", e
	}

	// 1. Проверяем что такой аккаунт существует

	// 2. Проверяем, что у пользователя есть к нему доступ

	// 3. Создаем долгий ключ для входа
	expiresAt := time.Now().UTC().Add(time.Hour * 2).Unix()
	claims := JWT{
		user.ID,
		account_id,
		user.SignedAccountID,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "GUI Server",
		},
	}

	return claims.CreateCryptoToken()

}

// создает Invite для пользователя
func (user *User) CreateInviteForUser (email string, sendMail bool) error {

	// 1. Создаем токен для нового пользователя
	eat := &EmailAccessToken{DestinationEmail:email, OwnerID:user.ID, ActionType: "invite-user"}
	err := eat.Create()
	if err != nil {
		return u.Error{Message:"Неудалось создать приглашение"}
	}

	// 2. Посылаем уведомление на почту
	if sendMail {
		if err := eat.SendMail(); err != nil {
			return u.Error{Message:"Неудалось отправить приглашение"}
		}
	}
	// user.SendNotification()...

	return nil
}