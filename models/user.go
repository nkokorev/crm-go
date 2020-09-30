package models

import (
	"errors"
	"fmt"
	"github.com/fatih/structs"
	u "github.com/nkokorev/crm-go/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
	"time"
)

type User struct {
	Id        		uint 		`json:"id" gorm:"primaryKey"`
	HashId 			string 		`json:"hash_id" gorm:"type:varchar(12);uniqueIndex;not null;"` // публичный Id для защиты от спама/парсинга
	IssuerAccountId uint 		`json:"issuer_account_id" gorm:"index;not null;"`
	IssuerAccountIdBeta uint 	`json:"issuerAccountId" gorm:"-"`

	Username 	*string 		`json:"username" gorm:"type:varchar(255);index;"` // уникальный, т.к. через него вход в главный аккаунт
	Email 		*string 		`json:"email" gorm:"type:varchar(255);index;"`
	PhoneRegion *string 		`json:"phone_region" gorm:"type:varchar(3);not null;default:'RU';"` // нужно проработать формат данных
	PhoneRegionBeta *string 	`json:"phoneRegion" gorm:"-"` // нужно проработать формат данных

	Phone		*string 		`json:"phone" gorm:"type:varchar(32);"` // нужно проработать формат данных
	Password 	*string 		`json:"-" gorm:"type:varchar(255);"` // json:"-"

	Name 		*string 		`json:"name" gorm:"type:varchar(64)"`
	Surname 	*string 		`json:"surname" gorm:"type:varchar(64)"`
	Patronymic 	*string 		`json:"patronymic" gorm:"type:varchar(64)"`

	// Разрешен ли вход, через app.ratuscrm.com  - deprecated
	EnabledAuthFromApp	bool	`json:"enabled_auth_from_app" gorm:"type:bool;default:false;"`

	// deprecated!!
	Subscribed			bool		`json:"subscribed" gorm:"type:bool;default:true;"` // Есть ли подписка на общее рассылки.
	SubscribedAt 		*time.Time 	`json:"subscribed_at"`
	SubscribedAtBeta 	*time.Time 	`json:"subscribedAt" gorm:"-"`
	UnsubscribedAt 		*time.Time 	`json:"unsubscribed_at"` // << last
	UnsubscribedAtBeta 	*time.Time 	`json:"unsubscribedAt" gorm:"-"` // << last

	// manual, gui, api, - deprecated!!
	SubscriptionReason		*string 	`json:"subscription_reason" gorm:"type:varchar(32);"`
	SubscriptionReasonBeta	*string 	`json:"subscriptionReason" gorm:"-"`

	// deprecated!!
	// UnsubscribedReason	*string `json:"unsubscribedReason" gorm:"default:null"` // << see mta-bounced...

	DefaultAccountId 		*uint 	`json:"default_account_id"` // указывает какой аккаунт по дефолту загружать
	// InvitedUserId 		*uint 	`json:"invited_user_id"` // кто его пригласил

	// Верификация, сброс пароля и т.д.
	EmailVerifiedAt *time.Time `json:"email_verified_at"` // дата подтверждения email-а (автоматически проставляется, если методом верфикации пользователя был подтвержден email)
	PhoneVerifiedAt *time.Time `json:"phone_verified_at"` // дата подтверждения телефона (автоматически проставляется, если методом верфикации пользователя был подтвержден телефон)
	PasswordResetAt *time.Time `json:"password_reset_at"`

	CreatedAt 		time.Time 	`json:"created_at"`
	CreatedAtBeta 	time.Time 	`json:"createdAt" gorm:"-"`
	UpdatedAt 		time.Time 	`json:"updated_at"`
	UpdatedAtBeta 	time.Time 	`json:"updatedAt" gorm:"-"`
	DeletedAt 		gorm.DeletedAt 	`json:"-" sql:"index"`

	AccountUser 	*AccountUser	`json:"account_user" gorm:"preload"` // WTF
	// Roles 		[]Role 			`json:"roles" gorm:"many2many:account_users;"`
	Accounts 		[]Account 		`json:"accounts" gorm:"many2many:account_users;"`
	Companies 		[]Company 		`json:"companies" gorm:"many2many:company_users;"`
	UsersSegments 	[]UsersSegment	`json:"users_segments" gorm:"many2many:users_segment_users;"`
}


func (User) PgSqlCreate() {
	if db.Migrator().HasTable(&User{}) { return }

	if err := db.Migrator().CreateTable(&User{}); err != nil { log.Fatal("(User) PgSqlCreate(): ", err) }

	// db.Model(&User{}).AddForeignKey("issuer_account_id", "accounts(id)", "SET DEFAULT", "CASCADE")
	// db.Model(&User{}).AddForeignKey("invited_user_id", "users(id)", "SET NULL", "CASCADE")
	err := db.Exec("ALTER TABLE users    \n    ADD CONSTRAINT users_issuer_account_id_fkey FOREIGN KEY (issuer_account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n--     ADD CONSTRAINT users_default_account_id_fkey FOREIGN KEY (default_account_id) REFERENCES accounts(id) ON DELETE SET NULL ON UPDATE CASCADE,\n--     ADD CONSTRAINT users_invited_user_id_fkey FOREIGN KEY (invited_user_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE,\n    ADD CONSTRAINT users_chk_unique check ((username is not null) or (email is not null) or (phone is not null));\n\ncreate unique index uix_users_issuer_account_id_username_email_mobile_phone ON users (issuer_account_id,username,email,phone);\n-- create unique index uix_users_username_account_id_sku ON users (issuer_account_id,username) where username is not null;\ncreate unique index uix_users_username_account_id_sku ON users (issuer_account_id,username) where (length(username) > 0);\n\n-- alter table users alter column default_account_id set default null;\n-- alter table users alter column invited_user_id set default null;\n\n-- alter table  users ADD CONSTRAINT users_chk_unique check ((username is not null) or (email is not null) or (phone is not null));\n").Error
	if err != nil {
		log.Fatal("Error PgSqlCreate User: ", err)
	}

	if err := db.SetupJoinTable(&User{}, "Accounts", &AccountUser{}); err != nil {
		log.Fatal(err)
	}

	if err := db.SetupJoinTable(&User{}, "Companies", &CompanyUser{}); err != nil {
		log.Fatal(err)
	}

	err = db.SetupJoinTable(&User{}, "UsersSegments", &UsersSegmentUser{})
	if err != nil {
		log.Fatal(err)
	}
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	user.Id = 0
	user.HashId = strings.ToLower(u.RandStringBytesMaskImprSrcUnsafe(12, true))

	/*if user.Subscribed {
		time := time.Now().UTC()
		user.SubscribedAt = &time
	}*/
	/*if user.Name != nil {
		str := *user.Name

		for len(str) > 0 {
			r, size := utf8.DecodeRuneInString(str)
			fmt.Printf("%c %v\n", r, size)

			str = str[size:]
			user.Name = &str
		}
	}*/
	// user.CreatedAt = time.Now().UTC()
	return nil
}

func (user *User) AfterCreate(tx *gorm.DB) error {
	// AsyncFire(*Event{}.UserCreated(user.IssuerAccountId, user.Id))
	AsyncFire(NewEvent("UserCreated", map[string]interface{}{"account_id":user.IssuerAccountId, "user_id":user.Id}))
	return nil
}
func (user *User) BeforeUpdate(tx *gorm.DB) (err error) {

	// fmt.Println(user.Subscribed)
	// fmt.Println(scope.Value.(*User).Subscribed)

	/*_input, ok := scope.Value.(*User)
	if ok {
		if (user.Subscribed != _input.Subscribed) {
			fmt.Println("Статус изменен!")
		}
	}*/
	// fmt.Printf("%T", scope.Value)


	return nil
}
func (user *User) AfterUpdate(tx *gorm.DB) (err error) {
	// AsyncFire(Event{}.UserUpdated(user.IssuerAccountId, user.Id))
	// AsyncFire(*NewEvent("UserUpdated", map[string]interface{}{"accountId":user.IssuerAccountId, "userId":user.Id}))
	return nil
}
func (user *User) AfterFind(tx *gorm.DB) (err error) {
	if user.PhoneRegion != nil {
		user.PhoneRegionBeta = user.PhoneRegion
	}
	if user.SubscribedAt != nil {
		user.SubscribedAtBeta = user.SubscribedAt
	}
	if user.UnsubscribedAt != nil {
		user.UnsubscribedAtBeta = user.UnsubscribedAt
	}
	if user.SubscriptionReason != nil {
		user.SubscriptionReasonBeta = user.SubscriptionReason
	}
	if user.SubscriptionReason != nil {
		user.SubscriptionReasonBeta = user.SubscriptionReason
	}
	user.CreatedAtBeta = user.CreatedAt
	user.UpdatedAtBeta = user.UpdatedAt

	user.IssuerAccountIdBeta = user.IssuerAccountId

	return nil
}

func (user User) create () (*User, error) {

	// !!! Проверка существования такого же пользователя для склейки - на строне аккаунта / контроллера !!!
	// !!! Проверка обязательных полей для конкретных настроек аккаунта на стороне аккаунта / контроллера !!!
	if err := user.ValidateCreate(); err != nil {
		return nil, err
	}

	// Теперь создаем crypto-пароль

	if user.Password != nil {
		password, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = u.STRp(string(password))
	}


	// fix
	var timeNow = time.Now()
	user.EmailVerifiedAt = &timeNow

	var userReturn = user
	// var userReturn = User{}


	if err := db.Create(&userReturn).First(&userReturn).Error; err != nil {
		return nil, err
	}

	// AsyncFire(*Event{}.UserCreated(user.IssuerAccountId, userReturn.Id))

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
func (user *User) load(preloads []string) error {

	err := db.Model(user).Preload("Roles").First(user,user.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (User) getByHashId(hashId string) (*User, error) {
	user := User{}

	err := db.First(&user, "hash_id = ?", hashId).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
func (user *User) update (input map[string]interface{}) error {

	delete(input,"roles")
	delete(input,"account_user")
	delete(input,"accountUsers")
	delete(input,"account")
	delete(input,"accounts")
	delete(input,"companies")
	delete(input,"users_segments")

	delete(input,"subscribedAt")
	delete(input,"unsubscribedAt")
	delete(input,"subscriptionReason")
	delete(input,"createdAt")
	delete(input,"updatedAt")
	delete(input,"issuerAccountId")
	delete(input,"phoneRegion")
	delete(input,"roleId")


	if err := u.ConvertMapVarsToUINT(&input, []string{"issuer_account_id","default_account_id","invited_user_id"}); err != nil {
		return err
	}

	err := db.Set("gorm:association_autoupdate", false).
		Model(user).Omit("id", "hash_id", "issuer_account_id", "created_at", "updated_at").Updates(input).Error
	if err != nil {
		 return err
	}
	
	return nil
}
func (user *User) save () error {

	return db.Set("gorm:association_autoupdate", false).
		Model(user).Omit("id", "hash_id", "issuer_account_id", "created_at", "updated_at").Save(user).Error

}
func (user *User) delete () error {
	return db.Model(&User{}).Where("id = ?", user.Id).Delete(user).Error
}
func getUserById(userId uint) (*User,error) {
	user := User{}
	err := db.Model(&User{}).First(&user, userId).Error
	return &user, err
}
func getUnscopedUserById(userId uint) (*User,error) {
	user := User{}
	err := db.Model(&User{}).Unscoped().First(&user, userId).Error
	return &user, err
}

// ####### Все что выше покрыто тестами (прямым и косвенными) ####### //

// ######### ACCOUNT @@

func (account Account) UpdateUser(userId uint, input map[string]interface{}) (*User, error) {

	// Проверка не нужна, т.к. поиск пользователя ее уже имеет
	user, err := account.GetUser(userId)
	if err != nil {
		return nil, err
	}

	// Отметка на будущее для события
	_newStatusSubscribed, ok := input["subscribed"].(bool)
	_user := *user

	// fmt.Println("User subs: ", input["subscribed"])
	err = user.update(input)
	if err != nil { return nil, err }

	// todo: возможно стоит проверить _user => user
	// Если флаг подписки был изменен
	if ok && (_newStatusSubscribed != _user.Subscribed) {


		// Статус обновлен
		// _ = user.Unsubscribing()

		// AsyncFire(*Event{}.UserUpdateSubscribeStatus(account.Id, _user.Id))
		AsyncFire(NewEvent("UserUpdateSubscribeStatus", map[string]interface{}{"account_id":account.Id, "user_id":_user.Id}))

		// флаги подписки / отписки
		if _newStatusSubscribed {
			// AsyncFire(*Event{}.UserSubscribed(account.Id, _user.Id))
			AsyncFire(NewEvent("UserSubscribed", map[string]interface{}{"account_id":account.Id, "user_id":_user.Id}))
		} else {
			// AsyncFire(*Event{}.UserUnsubscribed(account.Id, _user.Id))
			AsyncFire(NewEvent("UserUnsubscribed", map[string]interface{}{"account_id":account.Id, "user_id":_user.Id}))
		}

	}

	AsyncFire(NewEvent("UserUpdated", map[string]interface{}{"account_id":account.Id, "user_id":user.Id}))
	// AsyncFire(*Event{}.UserUpdated(account.Id, user.Id))

	return user, err
}
// осуществляет поиск по Id
func (account Account) GetUserById (userId uint) (*User, error) {

	user, err := User{}.get(userId)
	if err != nil {
		return nil, err
	}

	// Проверим, что пользователь имеет доступ к аккаунта
	aUser := AccountUser{}
	if err := db.Model(AccountUser{}).First(&aUser, "account_id = ? AND user_id = ?", account.Id, user.Id).Error;err !=nil {
		return nil, errors.New("Пользователь не найден")
	}

	return user, err
}

func (user *User) Get () error {
	/*return db.Preload("Accounts", func(db *gorm.DB) *gorm.DB {
		return db.Order(("accaunts.id DESC"))
	}).Find(user).Error*/

	//return db.Preload("Account").First(user,user.Id).Error
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
	if err := db.Unscoped().First(&User{}, user.Id).Error;err != nil {
		return false
	} else {
		return true
	}
}
func (User) ExistEmail(email string) bool {
	if err := db.Unscoped().First(&User{},"email = ?", email).Error;err != nil {
		return false
	} else {
		return true
	}
}
func (User) ExistUsername(username string) bool {
	if err := db.Unscoped().First(&User{},"username = ?", username).Error;err != nil {
		return false
	} else {
		return true
	}
}
func (user User) DepersonalizedDataMap() *map[string]interface{} {

	// получаем карту
	userMap := make(map[string]interface{})
	structs.FillMap(user, userMap)

	// 2.1 очищаем данные пользователя
	delete(userMap, "Id")
	delete(userMap, "IssuerAccountId")
	delete(userMap, "Password")
	delete(userMap, "DefaultAccountHashId")
	delete(userMap, "InvitedUserId")
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
	if user.Email != nil {
		email = true

		if err := u.EmailValidation(*user.Email); err != nil {
			if user.Email != nil && *user.Email != "demo-user@example.com" {
				return u.Error{Message:"Проверьте правильность заполнения формы", Errors: map[string]interface{}{"email":err.Error()}}
			}
		}
	}

	// 2. Проверка username пользователя
	if user.Username != nil {
		username = true
		if err := u.VerifyUsername(user.Username); err != nil {
			return u.Error{Message:"Проверьте правильность заполнения формы", Errors: map[string]interface{}{"username" : err.Error()}}
		}
	}

	// 3. Проверка телефона
	if user.Phone != nil  {
		phone = true
		if err := u.VerifyPhone(*user.Phone, *user.PhoneRegion); err != nil {
			return u.Error{Message:"Не верный формат телефона",Errors: map[string]interface{}{"phone" : err.Error()}}
		}
	}

	// 4. Проверка на одно из трех
	if !(username || email || phone ) {
		return u.Error{Message:"Отсутствуют обязательные поля", Errors: map[string]interface{}{"username":"Необходимо заполнить поле", "email":"Необходимо заполнить поле", "mobilePhone":"Необходимо заполнить поле"}}
	}

	// 4. Проверка password
	if user.Password != nil {
		if err := u.VerifyPassword(*user.Password); err != nil {
			return u.Error{Message:"Проверьте правильность заполнения формы", Errors: map[string]interface{}{"password" : err.Error()}}
		}
	}


	// 5. Проверка дополнительных полей на длину
	if user.Name != nil && len([]rune(*user.Name)) > 64 {
		e.AddErrors("name", "Имя слишком длинное" )
	}
	if user.Surname != nil && len([]rune(*user.Surname)) > 64 {
		e.AddErrors("surname", "Фамилия слишком длинная")
	}
	if user.Patronymic != nil && len([]rune(*user.Patronymic)) > 64 {
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
	if err := bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password)); err != nil {
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
	if user.Id < 1 {
		return errors.New("Внутрення ошибка из-за загрузки доступных аккаунтов")
	}
	return db.Preload("Accounts").First(&user).Error
}

type AccountUserAuth = struct {
	AccountUser
	Account Account	`json:"account"`
	Role 	Role	`json:"role"`
}

// Возвращает массив доступных аккаунтов с ролью в аккаунте
func (user User) AccountList() ([]AccountUser, error) {


	// aUsers := make([]AcoountUserAuth,0)
	var aUsers []AccountUser

	err := db.Model(&AccountUser{}).Preload("Role").Preload("Account").Find(&aUsers, "user_id = ?", user.Id).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, errors.New("Не удалось загрузить данные пользователя")
	}

	return aUsers, nil
}

// проверяет доступ и возвращает данные аккаунта
func (user *User) GetAccount(account *Account) error {
	return db.Model(user).Where("user_id = ?", user.Id).Association("Accounts").Find(account)
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
	aUser, err := account.AppendUser(user, *role)
	if err != nil || aUser == nil {
		return nil, err
	}

	// 3. Назначает роль owner


	return account, nil
}

// удаляет аккаунт, если пользователь имеет такие права
func (user *User) DeleteAccount(a *Account) error {

	// 1. Проверяем доступ "= проверяем права на удаление аккаунта"
	if db.Model(user).Where("account_id = ?", a.Id).Association("Accounts").Count() == 0 {
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
	eat := &EmailAccessToken{DestinationEmail: email, OwnerId:user.Id, ActionType: "invite-user"}
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


// =============

func (user User) GetDepersonalizedData() interface{} {
	return &user
}

// Пользователь самостоятельно отписывается или его отписывает система
func (user *User) Unsubscribing() error {
	input := map[string]interface{} {
		"subscribed":false,
		"unsubscribed_at":time.Now().UTC(),
	}

	// Событие отписки отслеживается в функции update()
	if err := user.update(input); err != nil { return nil }

	return nil
}

/*func (user User) GetHashKeyForValid() string {
	data := []byte(user.HashId + "RatusCRM#2020")
	hexHash := md5.Sum(data)
	return hex.EncodeToString(hexHash[:])
}
*/
// userHashId = уникальная 12-символ строка у пользователя, mtaHistoryHashId - сгенерированный будущий ключ
// func (account Account) GetUnsubscribeUrl(user User, mtaHistory MTAHistory) string {
func (account Account) GetUnsubscribeUrl(userHashId, mtaHistoryHashId string) string {

	AppEnv := os.Getenv("APP_ENV")
	crmHost := ""
	switch AppEnv {
	case "local":
		crmHost = "http://tracking.crm.local"
	case "public":
		crmHost = "https://tracking.ratuscrm.com"
	default:
		crmHost = "https://tracking.ratuscrm.com"
	}

	return crmHost + "/accounts/" +  account.HashId + "/e/unsubscribe?u=" + userHashId + "&hi=" + mtaHistoryHashId
}

func (account Account) GetPixelUrl(mtaHistoryHashId string) string {

	// var unsubscribeUrl := "http://tracking.crm.local/accounts/3niyoz4vucpz/e/unsubscribe?u=keqcfnymylb9&i=5&hi=vlbkv0bf9yr8"

	AppEnv := os.Getenv("APP_ENV")
	crmHost := ""
	switch AppEnv {
	case "local":
		crmHost = "http://tracking.crm.local"
	case "public":
		crmHost = "https://tracking.ratuscrm.com"
	default:
		crmHost = "https://tracking.ratuscrm.com"
	}

	// http://tracking.crm.local/accounts/3niyoz4vucpz/e/open?hi=mggigw8fiy9c
	return crmHost + "/accounts/" +  account.HashId + "/e/open?hi=" + mtaHistoryHashId
}

func (EmailTemplate) GetPixelHTML(pixelUrl string) string {

	// return `<img style="width: 1px;height: 1px;opacity: 0;" src='` + pixelUrl + `'/>`
	return fmt.Sprintf("<img style=\"width: 1px;height: 1px;opacity: 0;\" src=\"%v\"/>", pixelUrl)
}

