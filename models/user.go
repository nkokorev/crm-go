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
	//"gopkg.in/guregu/null.v3"
	"time"
)

type User struct {
	ID        	uint `json:"id"`
	//HashID 		string `json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Username 	string `json:"username" `
	Email 		string `json:"email"`
	Password 	string `json:"-"` // json:"-"

	Name 		string `json:"name"`
	Surname 	string `json:"surname"`
	Patronymic 	string `json:"patronymic"`
	//Phone	 	string `json:"patronymic"` // нужно проработать формат данных

	DefaultAccountID int `json:"default_account_id"` // указывает какой аккаунт по дефолту загружать

	EmailVerifiedAt *time.Time `json:"email_verified_at" gorm:"default:null"`
	//EmailVerification bool `json:"email_verification" gorm:"default:false"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`

	Accounts []Account `json:"-" gorm:"many2many:user_accounts"`
}

// структура настроек
type UserCreateSettings struct {
	SendEmailVerification bool
}

// ### CRUD FUNC ###

// Создает нового пользователя с новым ID
// Для единости интерфейса нельзя иметь обязательные переменные
func (user *User) Create (v_opt... UserCreateSettings ) error {

	// проверим входящие сообщения
	if err := user.ValidateCreate(); err != nil {
		return err
	}

	// Создаем крипто пароль
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(password)

	if db.Create(user).Error != nil {
		return err
	}


	// проверяем надо ли посылать уведомление для верификации аккаунта
	if len(v_opt) > 0 {
		createSettings := v_opt[0]

		if createSettings.SendEmailVerification {
			if err := user.SendEmailVerification(); err !=nil {
				return err
			}
		}
	}

	return nil
}

// осуществляет поиск по ID
func (user *User) Get () error {
	return db.First(user,user.ID).Error
}

// сохраняет все поля в модели, кроме id, deleted_at
func (user *User) Save () error {
	return db.Model(User{}).Omit("id", "deleted_at").Save(user).Find(user, "id = ?", user.ID).Error
}

// обновляет все схожие с интерфейсом поля, кроме id, username, deleted_at
func (user *User) Update (input interface{}) error {
	return db.Model(User{}).Where("id = ?", user.ID).Omit("id", "username", "deleted_at").Update(input).Find(user, "id = ?", user.ID).Error
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
		e.Message = "Неверно заполнены поля"
		return e
	}

	// 2. Проверка username пользователя
	if err := user.VerifyUsername(user.Username); err != nil {
		e.AddErrors("username", err.Error())
		e.Message = "Неверно заполнены поля"
		return e
	}

	// 3. Проверка password
	if err := user.VerifyPassword(user.Password); err != nil {
		e.AddErrors("password", err.Error())
		e.Message = "Неверно заполнены поля"
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
		e.Message = "Неверно заполнены поля"
		return e
	}

	return nil
}

func (user User) SendEmailVerification() error {

	// 1. Создаем токен
	if err := (&UserEmailAccessToken{UserID:user.ID, Type:"verification", Email:user.Email}).Create();err != nil {
		return err;
	}

	// 2. Посылаем уведомление с токеном
	// todo: send email verification

	return  nil
}

// проверяет token, в случае успеха удаляет его из БД
func (User) EmailVerified(token string) error {

	var user User
	var e u.Error
	uat := UserEmailAccessToken{Token:token}

	// 1. Пробуем найти токен
	err := uat.Get()
	if err != nil {
		e.Message = "Проверочный код не найден"
		return e
	}

	// 2. Проверяем тип токена (вообще это системная ошибка)
	if uat.Type != "verification" {
		e.Message = "Тип токена указан не верно"
		return e
	}

	// 3. Проверяем время жизни token
	if !time.Now().Add(-time.Hour * 24).Before(uat.CreatedAt) {
		e.Message = "Проверочный ключ устарел"
		return e
	}

	// 4. Проверяем существование пользователя с данными токена
	if err := db.First(&user, "id = ? AND email = ?", uat.UserID, uat.Email).Error; err != nil {
		e.Message = "Пользователь не найден"
		return e
	}

	// 5. Если все в порядке активируем учетную запись пользователя и сохраняем данные
	timeNow := time.Now()
	user.EmailVerifiedAt = &timeNow

	// 6. Сохраняем обновленные данные пользователя
	if err := user.Save(); err != nil {
		e.Message = "Неудалось обновить данные пользователя"
		return e
	}

	// 7. Удаляем проверочный код (больше не нужен)
	return uat.Delete()
	//return nil
}


// ### Account's FUNC ###

func (user *User) LoadAccounts() error {
	return db.Preload("Accounts").First(&user).Error
}

// проверяет доступ и возвращает данные аккаунта
func (user *User) GetAccount(account *Account) error {
	return db.Model(user).Where("user_id = ?", user.ID).Association("Accounts").Find(account).Error
}

// создание аккаунта от пользователя
func (user *User) CreateAccount(a *Account) error {

	// 1. Создаем аккаунт
	if err := a.Create(); err != nil {
		return err
	}

	// 2. Привязываем аккаунт к пользователю
	if err := a.AppendUser(user); err != nil {
		return err
	}

	// 3. Назначает роль owner

	return nil
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
	if err := a.Delete(); err != nil {
		return err
	}

	return nil
}


/// ### Auth FUNC ###

// Авторизует пользователя: загружает пользователя с предзагрузкой аккаунтов и, в случае успеха возвращает jwt-token
func (user *User) AuthLogin(username, password string, staySignedIn... bool) (string, error) {

	//user := &User{}
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
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil && err == bcrypt.ErrMismatchedHashAndPassword { //Password does not match!
		e.AddErrors("password", "Не верный пароль")
	}

	if e.HasErrors() {
		e.Message = "Не верно указаны данные"
		return "", e
	}

	expiresAt := time.Now().Add(time.Minute * 10).Unix()
	claims := JWT{
		user.ID,
		0,
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
	expiresAt := time.Now().Add(time.Hour * 2).Unix()
	claims := JWT{
		user.ID,
		account_id,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "GUI Server",
		},
	}

	return claims.CreateCryptoToken()

}