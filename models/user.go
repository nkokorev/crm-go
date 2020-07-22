package models

import (
	"errors"
	"github.com/fatih/structs"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	u "github.com/nkokorev/crm-go/utils"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

type User struct {
	ID        	uint `json:"id" gorm:"primary_key"`
	HashID string `json:"hashID" gorm:"type:varchar(12);unique_index;not null;"` // публичный ID для защиты от спама/парсинга
	IssuerAccountID uint `json:"issuerAccountID" gorm:"index;not null"`
	
	Username 	string `json:"username" gorm:"type:varchar(255);unique_index;default:null;"` // уникальный, т.к. через него вход в главный аккаунт
	Email 		string `json:"email" gorm:"type:varchar(255);index;default:null;"`
	PhoneRegion string `json:"phoneRegion" gorm:"type:varchar(3);not null;default:'RU';"` // нужно проработать формат данных
	Phone		string `json:"phone" gorm:"type:varchar(32);default:null;"` // нужно проработать формат данных
	Password 	string `json:"-" gorm:"type:varchar(255);default:null;"` // json:"-"

	Name 		string `json:"name" gorm:"type:varchar(64)"`
	Surname 	string `json:"surname" gorm:"type:varchar(64)"`
	Patronymic 	string `json:"patronymic" gorm:"type:varchar(64)"`

	//Role 		string `json:"role" gorm:"type:varchar(255);default:'client'"`
	Roles 		[]Role `json:"roles" gorm:"many2many:account_users;preload"`
	AccountUser *AccountUser	`json:"accountUser" gorm:"preload"`
	// Account 	Account `json:"account" gorm:"preload" sql:"-"`

	EnabledAuthFromApp	bool	`json:"enabledAuthFromApp" gorm:"type:bool;default:false;"` // Разрешен ли вход, через app.ratuscrm.com

	DefaultAccountID uint `json:"defaultAccountID" gorm:"type:varchar(12);default:null;"` // указывает какой аккаунт по дефолту загружать
	InvitedUserID uint `json:"-" gorm:"default:NULL"` // указывает какой аккаунт по дефолту загружать

	// Верификация, сброс пароля и т.д.
	EmailVerifiedAt *time.Time `json:"emailVerifiedAt" gorm:"default:null"` // дата подтверждения email-а (автоматически проставляется, если методом верфикации пользователя был подтвержден email)
	PhoneVerifiedAt *time.Time `json:"phoneVerifiedAt" gorm:"default:null"` // дата подтверждения телефона (автоматически проставляется, если методом верфикации пользователя был подтвержден телефон)
	PasswordResetAt *time.Time `json:"passwordResetAt" gorm:"default:null"`


	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"-" sql:"index"`

	//Profile UserProfile `json:"profile" gorm:"preload"`

	Accounts []Account `json:"-" gorm:"many2many:account_users;preload"`
}

type UserAndRole struct {
	User
	RoleID uint `json:"roleID"`
}

func (User) PgSqlCreate() {
	db.CreateTable(&User{})

	db.Exec("ALTER TABLE users\n    \n--     ADD CONSTRAINT users_issuer_account_id_fkey FOREIGN KEY (issuer_account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n--     ADD CONSTRAINT users_default_account_hash_id_fkey FOREIGN KEY (default_account_hash_id) REFERENCES accounts(hash_id) ON DELETE SET NULL ON UPDATE CASCADE,\n--     ADD CONSTRAINT users_invited_user_id_fkey FOREIGN KEY (invited_user_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE,\nADD CONSTRAINT users_chk_unique check ((username is not null) or (email is not null) or (phone is not null));\n\ncreate unique index uix_users_issuer_account_id_username_email_mobile_phone ON users (issuer_account_id,username,email,phone);\n\n-- alter table  users ADD CONSTRAINT users_chk_unique check ((username is not null) or (email is not null) or (phone is not null));\n")
	db.Model(&User{}).AddForeignKey("issuer_account_id", "accounts(id)", "SET DEFAULT", "CASCADE")
	// db.Model(&User{}).AddForeignKey("default_account_hash_id", "accounts(id)", "SET NULL", "CASCADE")
	db.Model(&User{}).AddForeignKey("invited_user_id", "users(id)", "SET NULL", "CASCADE")
}

func (user *User) BeforeCreate(scope *gorm.Scope) (err error) {
	user.ID = 0
	user.HashID = strings.ToLower(u.RandStringBytesMaskImprSrcUnsafe(12, true))
	// user.CreatedAt = time.Now().UTC()
	return nil
}


func (user User) create () (*User, error) {

	// !!! Проверка существования такого же пользователя для склейки - на строне аккаунта / контроллера !!!
	// !!! Проверка обязательных полей для конкретных настроек аккаунта на стороне аккаунта / контроллера !!!
	if err := user.ValidateCreate(); err != nil {
		return nil, err
	}

	// Теперь создаем crypto-пароль
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.Password = string(password)
	// fix
	var timeNow = time.Now()
	user.EmailVerifiedAt = &timeNow

	var userReturn = user

	if err := db.Create(&userReturn).First(&userReturn).Error; err != nil {
		return nil, err
	}

	event.AsyncFire(Event{}.UserCreated(user.IssuerAccountID, userReturn.ID))

	return &userReturn, nil
}

func (User) get(id uint) (*User, error) {
	user := User{}

	err := db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
/*func (User) getWithAUser(accountID, id uint) (*User, error) {
	user := User{}

	err := db.Preload("AccountUser", func(db *gorm.DB) *gorm.DB {
		return db.Where("account_id = ?", accountID).Select(AccountUser{}.SelectArrayWithoutBigObject())
	}).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}*/
func (user *User) load() error {

	err := db.Model(user).Preload("Roles").First(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (User) getByHashID(hashID string) (*User, error) {
	user := User{}

	err := db.First(&user, "hash_id = ?", hashID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (user *User) update (input map[string]interface{}) error {

	err := db.Set("gorm:association_autoupdate", false).
		Model(user).Omit("id", "hash_id", "issuer_account_id", "created_at", "updated_at").Updates(input).Error
	if err != nil {
		 return err
	}
	
	// event.AsyncFire(Event{}.UserUpdated(user.IssuerAccountID, user.ID))

	return nil
}
func (user *User) save () error {

	return db.Set("gorm:association_autoupdate", false).
		Model(user).Omit("id", "hash_id", "issuer_account_id", "created_at", "updated_at").Save(user).Error

}

func (user User) delete () error {
	return db.Model(&User{}).Where("id = ?", user.ID).Delete(user).Error
}

func getUserByID(userID uint) (*User,error) {
	user := User{}
	err := db.Model(&User{}).First(&user, userID).Error
	return &user, err
}

func getUnscopedUserByID(userID uint) (*User,error) {
	user := User{}
	err := db.Model(&User{}).Unscoped().First(&user, userID).Error
	return &user, err
}

// ####### Все что выше покрыто тестами (прямым и косвенными) ####### //

// ######### ACCOUNT @@

func (account Account) UpdateUser(userID uint, input map[string]interface{}) (*User, error) {

	// Проверка не нужна, т.к. поиск пользователя ее уже имеет
	user, err := account.GetUser(userID)
	if err != nil {
		return nil, err
	}
	
	err = user.update(input)
	if err != nil { return nil, err }

	event.AsyncFire(Event{}.UserUpdated(account.ID, user.ID))

	return user, err
}
// осуществляет поиск по ID
func (account Account) GetUserByID (userID uint) (*User, error) {

	user, err := User{}.get(userID)
	if err != nil {
		return nil, err
	}

	// Проверим, что пользователь имеет доступ к аккаунта
	aUser := AccountUser{}
	if db.Model(AccountUser{}).First(&aUser, "account_id = ? AND user_id = ?", account.ID, user.ID).RecordNotFound() {
		return nil, errors.New("Пользователь не найден")
	}

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
func (User) GetByUsername (username string) (*User, error) {
	var user User
	if err := db.First(&user,"username = ?", username).Error; err != nil {return nil, err}

	return &user, nil
}

func (user User) Exist() bool {
	return !db.Unscoped().First(&User{}, user.ID).RecordNotFound()
}

func (User) ExistEmail(email string) bool {
	return !db.Unscoped().First(&User{},"email = ?", email).RecordNotFound()
}

func (User) ExistUsername(username string) bool {
	return !db.Unscoped().First(&User{},"username = ?", username).RecordNotFound()
}

func (user User) DepersonalizedDataMap() *map[string]interface{} {

	// получаем карту
	userMap := make(map[string]interface{})
	structs.FillMap(user, userMap)

	// 2.1 очищаем данные пользователя
	delete(userMap, "ID")
	delete(userMap, "IssuerAccountID")
	delete(userMap, "Password")
	delete(userMap, "DefaultAccountHashID")
	delete(userMap, "InvitedUserID")
	delete(userMap, "EmailVerifiedAt")
	delete(userMap, "PhoneVerifiedAt")
	delete(userMap, "PasswordResetAt")
	delete(userMap, "CreatedAt")
	delete(userMap, "UpdatedAt")
	delete(userMap, "DeletedAt")

	return &userMap
}

// Проверка НЕ нулевых входящих полей для СОЗДАНИЯ пользователя
func (user User) ValidateCreate() error {

	var e u.Error
	var username, email, phone bool

	// 1. Проверка email отдельной функцией
	if len(user.Email) > 0 {
		email = true

		if err := u.EmailValidation(user.Email); err != nil {
			return u.Error{Message:"Проверьте правильность заполнения формы", Errors: map[string]interface{}{"email":err.Error()}}
		}
	}

	// 2. Проверка username пользователя
	if len(user.Username) > 0 {
		username = true
		if err := u.VerifyUsername(user.Username); err != nil {
			return u.Error{Message:"Проверьте правильность заполнения формы", Errors: map[string]interface{}{"username" : err.Error()}}
		}
	}

	// 3. Проверка телефона
	if len(user.Phone) > 0 {
		phone = true
		if err := u.VerifyPhone(user.Phone, user.PhoneRegion); err != nil {
			return u.Error{Message:"Не верный формат телефона",Errors: map[string]interface{}{"phone" : err.Error()}}
		}
	}

	// 4. Проверка на одно из трех
	if !(username || email || phone ) {
		return u.Error{Message:"Отсутствуют обязательные поля", Errors: map[string]interface{}{"username":"Необходимо заполнить поле", "email":"Необходимо заполнить поле", "mobilePhone":"Необходимо заполнить поле"}}
	}

	// 4. Проверка password
	if len(user.Password) > 0 {
		if err := u.VerifyPassword(user.Password); err != nil {
			return u.Error{Message:"Проверьте правильность заполнения формы", Errors: map[string]interface{}{"password" : err.Error()}}
		}
	}


	// 5. Проверка дополнительных полей на длину
	if len([]rune(user.Name)) > 64 {
		e.AddErrors("name", "Имя слишком длинное" )
	}
	if len([]rune(user.Surname)) > 64 {
		e.AddErrors("surname", "Фамилия слишком длинная")
	}
	if len([]rune(user.Patronymic)) > 64 {
		e.AddErrors("patronymic", "Отчетство слишком длинное" )
	}

	// Чекаем мелкие ошибки
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


// ### Account's FUNC ###

// Загружает список аккаунтов...
func (user *User) LoadAccounts() error {
	if user.ID < 1 {
		return errors.New("Внутрення ошибка из-за загрузки доступных аккаунтов")
	}
	return db.Preload("Accounts").First(&user).Error
}

type AcoountUserAuth = struct {
	AccountUser
	Account Account	`json:"account"`
	Role Role	`json:"role"`
}

// Возвращает массив доступных аккаунтов с ролью в аккаунте
func (user User) AccountList() ([]AcoountUserAuth, error) {


	aUsers := make([]AcoountUserAuth,0)

	err := db.Model(&AccountUser{}).Preload("Role").Preload("Account").Find(&aUsers, "user_id = ?", user.ID).Error;
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.New("Не удалось загрузить данные пользователя")
	}

	return aUsers, nil
}

// проверяет доступ и возвращает данные аккаунта
func (user *User) GetAccount(account *Account) error {
	return db.Model(user).Where("user_id = ?", user.ID).Association("Accounts").Find(account).Error
}

// Только пользователь RatusCRM может создавать новые аккаунты
func (user User) CreateAccount(input Account) (*Account,error) {

	// Проверяем пользователя и его роль в RatusCRM.

	// 1. Создаем аккаунт
	account, err := input.create()
	if err != nil {
		return nil, err
	}


	role, err := account.GetRoleByTag(RoleOwner)
	if err != nil {
		return nil, err
	}

	// 2. Привязываем аккаунт к пользователю
	aUser, err := account.AppendUser(user, *role);
	if err != nil || aUser == nil {
		return nil, err
	}

	// 3. Назначает роль owner


	return account, nil
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

// создает Invite для пользователя
func (user *User) CreateInviteForUser (email string, sendMail bool) error {

	// 1. Создаем токен для нового пользователя
	eat := &EmailAccessToken{DestinationEmail: email, OwnerID:user.ID, ActionType: "invite-user"}
	err := eat.Create()
	if err != nil {
		return u.Error{Message:"Не удалось создать приглашение"}
	}

	// 2. Посылаем уведомление на почту
	if sendMail {
		if err := eat.SendMail(); err != nil {
			return u.Error{Message:"Не удалось отправить приглашение"}
		}
	}
	// user.SendNotification()...

	return nil
}
