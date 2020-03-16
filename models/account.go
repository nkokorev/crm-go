package models

import (
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"strings"
	"time"
)

type authMethod string

const (
	username  authMethod = "username"
	email authMethod = "email"
	phone authMethod = "phone"
)

type Account struct {
	ID uint `json:"id" gorm:"primary_key"`
	HashID string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный ID для защиты от спама/парсинга

	// данные аккаунта
	Name string `json:"name" gorm:"type:varchar(255)"`
	Website string `json:"website" gorm:"type:varchar(255)"` // спорно
	Type string `json:"type" gorm:"type:varchar(255)"` // спорно

	// API Интерфейс
	ApiEnabled bool `json:"apiEnabled" gorm:"default:true;not null"` // включен ли API интерфейс у аккаунта (false - все ключи отключаются, есть ли смысл в нем?)

	// UI-API Интерфейс (https://ui.api.ratuscrm.com / https://ratuscrm.com/ui-api)
	UiApiEnabled bool `json:"uiApiEnabled" gorm:"default:false;not null"` // Принимать ли запросы через публичный UI-API интерфейсу (через https://ui.api.ratuscrm.com)
	UiApiAesEnabled bool `json:"uiApiAesEnabled" gorm:"default:true;not null"` // Включение AES-128/CFB шифрования для публичного UI-API
	UiApiAesKey string `json:"uiApiAesKey" gorm:"type:varchar(16);default:null;"` // 128-битный ключ шифрования
	UiApiJwtKey string `json:"uiApiJwtKey" gorm:"type:varchar(32);default:null;"` // 128-битный ключ шифрования

	// Регистрация новых пользователей через UI/API

	//AuthMethod authMethod `json:"authBy" gorm:"enum('username', 'email', 'mobilePhone');not null;default:'email';"`
	//AuthMethod authMethod `json:"authBy" sql:"type:auth_method;not null;default:'email'"` // Дефолтный вариант авторизации todo: может массивом т.к. их может быть несколько?
	UiApiAuthMethods pq.StringArray `json:"uiApiAuthMethods" sql:"type:varchar(32)[];default:'{email}'"` // Доступные способы авторизации (проверяется в контроллере)
	UiApiEnabledUserRegistration bool `json:"uiApiEnabledUserRegistration" gorm:"default:true;not null"` // Разрешить регистрацию новых пользователей?
	UiApiUserRegistrationInvitationOnly bool `json:"uiApiUserRegistrationInvitationOnly" gorm:"default:false;not null"` // Регистрация новых пользователей только по приглашению (в том числе и клиентов)
	UiApiUserRegistrationRequiredFields pq.StringArray `json:"uiApiUserRegistrationRequiredFields" gorm:"type:varchar(32)[];default:'{email}'"` // список обязательных НЕ нулевых полей при регистрации новых пользователей через UI/API
	UiApiUserEmailDeepValidation bool `json:"uiApiUserEmailDeepValidation" gorm:"default:false;not null"` // глубокая проверка почты пользователя на предмет существования

	UserVerificationMethodID uint `json:"userVerificationMethodId" gorm:"type:int;default:null"` // метод
	UiApiEnabledLoginNotVerifiedUser bool `json:"uiApiEnabledLoginNotVerifiedUser" gorm:"default:false;"` // разрешать ли пользователю входить в аккаунт без завершенной верфикации?


	// настройки авторизации.
	// Разделяется AppAuth и ApiAuth -
	VisibleToClients bool `json:"visibleToClients" gorm:"default:false"` // скрывать аккаунт в списке доступных для пользователей с ролью 'client'. Нужно для системных аккаунтов.
	ClientsAreAllowedToLogin bool `json:"allowToLogin_for_clients" gorm:"default:false"` // запрет на вход в ratuscrm для пользователей с ролью 'client' (им не будет выдана авторизация).

	AuthForbiddenForClients bool `json:"authForbiddenForClients" gorm:"default:false"` // запрет авторизации для для пользователей с ролью 'client'.

	//ForbiddenForClient bool `json:"forbidden_for_client" gorm:"default:false"` // запрет на вход через приложение app.ratuscrm.com для пользователей с ролью 'client'

	CreatedAt 	time.Time `json:"createdAt"`
	UpdatedAt 	time.Time `json:"-"`
	DeletedAt 	*time.Time `json:"-" sql:"index"`

	//Users 		[]User `json:"-" gorm:"many2many:user_accounts"`
	Users 		[]User `json:"-" gorm:"many2many:account_users"`
	ApiKeys 	[]ApiKey `json:"-"`

	Products 	[]Product `json:"-"`
	Stocks		[]Stock `json:"-"`
}

func (account *Account) BeforeCreate(scope *gorm.Scope) error {
	account.ID = 0
	account.HashID = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	//account.HashID = utils.GetMD5Hash(account.Name + "RatusCRM" + time.Now().UTC().String())
	account.CreatedAt = time.Now().UTC()

	//account.UiApiJwtKey =  utils.CreateHS256Key()
	//scope.SetColumn("ui_api_jwt_key", "fjdsfdfsjkfskjfds")
	//scope.SetColumn("ID", uuid.New())
	return nil
}

func (Account) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Account{})
	db.Exec("ALTER TABLE accounts \n--     ADD CONSTRAINT uix_email_account_id_parent_id unique (email,account_id,parent_id),\n    ADD CONSTRAINT accounts_user_verification_method_id_fkey FOREIGN KEY (user_verification_method_id) REFERENCES user_verification_methods(id) ON DELETE CASCADE ON UPDATE CASCADE;\n--     ADD CONSTRAINT users_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,\n--     ALTER COLUMN parent_id SET DEFAULT NULL,\n--     ADD CONSTRAINT users_default_account_id_fkey FOREIGN KEY (default_account_id) REFERENCES accounts(id) ON DELETE SET NULL ON UPDATE CASCADE,    \n--     ADD CONSTRAINT users_invited_user_id_fkey FOREIGN KEY (invited_user_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE;\n\n-- create unique index uix_user_id_account_id_email_parent_id_not_null ON users (account_id,email,parent_id) WHERE parent_id IS NOT NULL;\n-- create unique index uix_account_id_email_parent_id_when_null ON users (account_id,email,parent_id) WHERE parent_id IS NULL;\n")

	// 2. Создаем Главный аккаунт через спец. функцию
	_, err := CreateMainAccount()
	if err != nil {
		log.Fatal("Неудалось создать главный аккаунт. Ошибка: ", err)
	}

	// 3. Создаем API-ключ в аккаунте
/*	_, err = mAcc.CreateApiKey(ApiKey{Name:"Api key for Postman"})
	if err != nil {
		log.Fatalf("Неудалось создать API ключ для аккаунта: %v, Error: %s", mAcc.Name, err)
	}*/
}

func (account *Account) Reset() { account = &Account{} }

func (account Account) create () (*Account, error) {

	var err error
	var outAccount Account // returned var

	if err := account.ValidateInputs(); err !=nil {
		return nil, err
	}

	// Создаем ключи для UI API
	outAccount.UiApiAesKey, err = utils.CreateAes128Key()
	if err != nil {
		return nil, err
	}

	outAccount.UiApiJwtKey =  utils.CreateHS256Key()

	// Копируем, то что можно использовать при создании
	outAccount.Name = account.Name
	outAccount.Website = account.Website
	outAccount.Type = account.Type

	outAccount.ApiEnabled = account.ApiEnabled

	outAccount.UiApiEnabled = account.UiApiEnabled
	outAccount.UiApiAesEnabled = account.UiApiAesEnabled

	// Регистрация новых пользователей через UI/API
	outAccount.UiApiAuthMethods = account.UiApiAuthMethods
	outAccount.UiApiEnabledUserRegistration = account.UiApiEnabledUserRegistration
	outAccount.UiApiUserRegistrationInvitationOnly = account.UiApiUserRegistrationInvitationOnly
	outAccount.UiApiUserRegistrationRequiredFields = account.UiApiUserRegistrationRequiredFields
	outAccount.UiApiUserEmailDeepValidation = account.UiApiUserEmailDeepValidation

	outAccount.UserVerificationMethodID = account.UserVerificationMethodID
	outAccount.UiApiEnabledLoginNotVerifiedUser = account.UiApiEnabledLoginNotVerifiedUser

	outAccount.VisibleToClients = account.VisibleToClients
	outAccount.ClientsAreAllowedToLogin = account.ClientsAreAllowedToLogin

	// Создание аккаунта
	if err := db.Omit("ID").Create(&outAccount).Error; err != nil {
		return nil, err
	}

	return &outAccount, nil
}

func CreateMainAccount() (*Account, error) {

	// Проверяем есть ли Главны Аккаунт
	_, err := GetMainAccount()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	dvc, err := GetUserVerificationTypeByCode(VerificationMethodEmailAndPhone)
	if err != nil || dvc == nil {
		return nil, errors.New("Неудалось получить код двойной верификации по телефону и почте")
	}

	return (Account{
		Name:"RatusCRM",
		UiApiEnabled:false,
		UiApiAesEnabled:true,
		UiApiEnabledUserRegistration:false,
		UiApiUserRegistrationInvitationOnly:false,
		ApiEnabled: false,
		UiApiAuthMethods: pq.StringArray{"username,email,phone"},
		UiApiUserRegistrationRequiredFields: pq.StringArray{"username,email,phone"},

		UserVerificationMethodID: dvc.ID,
		UiApiEnabledLoginNotVerifiedUser: false,
	}).create()
}

func (account Account) ValidateInputs() error {

	if len(account.Name) < 2 {
		return utils.Error{Message:"Ошибки в заполнении формы", Errors: map[string]interface{}{"name":"Имя компании должно содержать минимум 2 символа"}}
	}

	if len(account.Name) > 64 {
		return utils.Error{Message:"Ошибки в заполнении формы", Errors: map[string]interface{}{"name":"Имя компании должно быть не более 42 символов"}}
	}

	if len(account.Website) > 255 {
		return utils.Error{Message:"Ошибки в заполнении формы", Errors: map[string]interface{}{"website":"Слишком длинный url"}}
	}

	if len(account.Type) > 255 {
		return utils.Error{Message:"Ошибки в заполнении формы", Errors: map[string]interface{}{"type":"Слишком длинный текст"}}
	}

	return nil
}

func GetAccount (id uint) (*Account, error) {
	var account Account
	err := db.Model(&Account{}).First(&account, id).Error
	return &account, err
}

func GetMainAccount() (*Account, error) {
	var account Account
	err := db.Model(&Account{}).First(&account, "id = 1 AND name = 'RatusCRM'").Error
	//if err != nil { account.Reset() }
	return &account, err
}

func GetAccountByHash (hashId string) (*Account, error) {
	var account Account
	err := db.Model(&Account{}).First(&account, "hash_id = ?", hashId).Error
	return &account, err
}

func (Account) Exist(id uint) bool {
	return !db.Model(Account{}).First(&Account{}, id).RecordNotFound()
}


// ### API KEY ###

func (account Account) CreateApiKey (input ApiKey) (*ApiKey, error) {
	if account.ID < 1 {
		return nil, utils.Error{Message:"Внутреняя ошибка платформы", Errors: map[string]interface{}{"apiKey":"Неудалось привязать ключ к аккаунте"}}
	}
	input.AccountID = account.ID
	return input.create()
}

func (account Account) GetApiKey(token string) (*ApiKey, error) {
	apiKey, err := GetApiKey(token)
	if err != nil {
		return nil, err
	}

	if apiKey.AccountID != account.ID {
		return nil, errors.New("ApiKey не принадлежит аккаунту")
	}

	return apiKey, nil
}

func (account Account) DeleteApiKey(token string) error {

	apiKey, err := account.GetApiKey(token)
	if err != nil {
		return err
	}

	return apiKey.delete()
}

func (account Account) UpdateApiKey(token string, input ApiKey) (*ApiKey, error) {
	apiKey, err := account.GetApiKey(token)
	if err != nil {
		return nil, err
	}

	err = apiKey.update(input)

	return apiKey,err

}


// #### User ####

func (account Account) CreateUser(input User, v_opt... accessRole) (*User, error) {

	var err error
	var username, email, phone bool
	var role accessRole

	// Проверяем роль
	if len(v_opt) > 0 {
		role = v_opt[0]
		// нельзя создать пользователя с ролью Owner
		if role == RoleOwner {
			role = RoleAdmin
		}
	} else {
		role = RoleClient
	}

	input.IssuerAccountID = account.ID

	// ### !!!! Проверка входящих данных !!! ### ///

	if len(input.Username) > 0 {
		username = true
		if err := utils.VerifyUsername(input.Username); err != nil {
			return nil, utils.Error{Message:"Проверьте правильность заполнения формы", Errors: map[string]interface{}{"username" : err.Error()}}
		}
	}

	if len(input.Email) > 0 {
		email = true
		if account.UiApiUserEmailDeepValidation {
			if err := utils.EmailDeepValidation(input.Email); err != nil {
				return nil, utils.Error{Message:"Проверьте правильность заполнения формы", Errors: map[string]interface{}{"email":err.Error()}}
			}
		} else {
			if err := utils.EmailValidation(input.Email); err != nil {
				return nil, utils.Error{Message:"Проверьте правильность заполнения формы", Errors: map[string]interface{}{"email":err.Error()}}
			}
		}
	}

	if len(input.Phone) > 0 {
		phone = true

		if input.PhoneRegion == "" {
			input.PhoneRegion = "RU" // todo тут можно по IP определить где находиться пользователь +/-
		}

		// Устанавливаем нужный формат
		input.Phone, err = utils.ParseE164Phone(input.Phone, input.PhoneRegion)
		if err != nil {
			return nil, utils.Error{Message:"Ошибка в формате телефонного номера", Errors: map[string]interface{}{"inviteToken":"Пожалуйста, укажите номер телефона в международном формате"}}
		}

	}

	// 5. One of username. email and phone must be!
	if !(username || email || phone ) {
		return nil, utils.Error{Message:"Отсутствуют обязательные поля", Errors: map[string]interface{}{"username":"Необходимо заполнить поле", "email":"Необходимо заполнить поле", "mobilePhone":"Необходимо заполнить поле"}}
	}

	// Проверка дублирование полей
	if account.existUserUsername(input.Username) {
		return nil, utils.Error{Message:"Данные уже есть", Errors: map[string]interface{}{"username":"Данный username уже используется"}}
	}
	if account.existUserEmail(input.Email) {
		return nil, utils.Error{Message:"Данные уже есть", Errors: map[string]interface{}{"username":"Данный email уже используется"}}
	}
	if account.existUserPhone(input.Phone) {
		return nil, utils.Error{Message:"Данные уже есть", Errors: map[string]interface{}{"username":"Данный телефон уже используется"}}
	}

	// создаем пользователя
	u, err := input.create()
	if err != nil || u == nil {
		return u, err
	}

	// Автоматически добавляем пользователя в аккаунт
	aUser, err := account.AppendUser(*u, role);
	if err != nil || aUser == nil {
		return nil, err
	}
	return u, nil
}

func (account Account) GetUserById(userId uint) (*User, error) {
	user := User{}

	err := db.Model(&User{}).Where("issuer_account_id = ?", account.ID).First(&user, userId).Error

	return &user, err
}

func (account Account) GetUserByUsername (username string) (*User, error) {

	if username == "" {
		return nil, gorm.ErrRecordNotFound
	}

	user := User{}

	err := db.Model(&User{}).Where("issuer_account_id = ? AND username = ?", account.ID, username).First(&user).Error

	return &user, err
}

func (account Account) GetUserByEmail (email string) (*User, error) {
	if email == "" {
		return nil, gorm.ErrRecordNotFound
	}

	user := User{}

	err := db.Model(&User{}).Where("issuer_account_id = ? AND email = ?", account.ID, email).First(&user).Error

	return &user, err
}

func (account Account) GetUserByPhone (phone, region string) (*User, error) {
	if phone == "" {
		return nil, gorm.ErrRecordNotFound
	}

	if region == "" {
		region = "RU"
	}

	phone, _ = utils.ParseE164Phone(phone, region)

	user := User{}

	err := db.Model(&User{}).Where("issuer_account_id = ? AND phone = ?", account.ID, phone).First(&user).Error

	return &user, err
}

// Проверяет существование пользователя вообще
func (account Account) ExistUser(user User) bool {
	return !db.Model(&User{}).First(&User{}, user.ID).RecordNotFound()
}

// Проверяет существоание пользователя в контексте текущего аккаунта
func (account Account) ExistAccountUser(user User) bool {

	if db.Model(&AccountUser{}).Where("account_id = ? AND user_id = ?", account.ID, user.ID).Find(&AccountUser{}).RecordNotFound() {
		return false
	} else {
		return true
	}

}

// Если пользователь не найден - вернет gorm.ErrRecordNotFound
func (account Account) GetAccountUser(user User) (*AccountUser, error) {

	aUser := AccountUser{}

	if db.NewRecord(account) || db.NewRecord(user) {
		return nil, errors.New("GetUserRole: Аккаунта или пользователя не существует!")
	}

	err := db.Model(&AccountUser{}).
		Where("account_id = ? AND user_id = ?", account.ID, user.ID).
		Preload("Role").
		Preload("Account").
		Preload("User").
		First(&aUser).Error;

	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	if &aUser == nil {
		return nil, errors.New("Не удалось создать пользователя")
	}

	if err == gorm.ErrRecordNotFound {
		return nil, err
	}

	return &aUser, nil
}

// добавляет пользователя в аккаунт. Если пользователь уже в аккаунте, то роль будет обновлена
func (account Account) AppendUser(user User, tag accessRole) (*AccountUser, error) {

	acs := AccountUser{} // return value

	if db.NewRecord(user) || !account.ExistUser(user) {
		return nil, errors.New("Необходимо создать сначала пользователя!")
	}

	// узнаем роль, чтобы потом получить его ID
	rSet, err := GetRole(tag)
	if err != nil || rSet == nil {
		return nil, err
	}

	// проверяем, относится ли пользователь к аккаунту
	if account.ExistAccountUser(user) {
		return nil, errors.New("Невозможно добавить пользователя в аккаунт, т.к. он в нем уже есть.")
	} else {
		// создаем
		acs.AccountId = account.ID
		acs.UserId = user.ID
		acs.RoleId = rSet.ID

		aUser, err := acs.create();
		if err != nil || aUser == nil{
			return nil, errors.New("Ошибка при создании AccountUser.")
		}

		return aUser, nil
	}
}

// !!!!!! ### Выше функции покрытые тестами ### !!!!!!!!!!1

// В случае успеха возвращает jwt-token
func (account Account) AuthUserByUsername(username, password string, onceLogin_opt... bool) (string, error)  {

	var e utils.Error

	user, err := account.GetUserByUsername(username)
	if err != nil || user == nil {
		return "", errors.New("Пользователь не найден")
	}
	
	// если пользователь не найден temp.Username == nil, то пароль не будет искаться, т.к. он будет равен нулю (не с чем сравнивать)
	if !user.ComparePassword(password) {
		e.AddErrors("password", "Неверный пароль")
	}
	
	if e.HasErrors() {
		e.Message = "Проверьте указанные данные"
		return "", e
	}

	expiresAt := time.Now().UTC().Add(time.Minute * 20).Unix()

	claims := JWT{
		user.ID,
		account.ID,
		user.IssuerAccountID,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AuthServer",
		},
		*user,
		account,
	}

	return claims.CreateCryptoToken()
}


// *** New functions ****

// Обязательно ли поле при создании пользователя (username, email, phone)
func (account Account) userRequiredField(field string) bool {
	for _, v := range account.UiApiUserRegistrationRequiredFields {
		if v == field {
			return true
		}
	}
	return false
}

// Проверяет поля в input на не нулевость в соответствие настройкам аккаунта
func (account Account) ValidationUserRegReqFields(input User) error {
	var e utils.Error
	for _,v := range account.UiApiUserRegistrationRequiredFields {
		switch v {
		case "username":
			if len(input.Username) == 0 {
				e.AddErrors("username","Поле обязательно к заполнению")
			}

		case "email":
			if len(input.Email) == 0 {
				e.AddErrors("email","Поле обязательно к заполнению")
			}

		case "phone":
			if len(input.Phone) == 0 {
				e.AddErrors("phone","Поле обязательно к заполнению")
			}

		}
	}

	if e.HasErrors() {
		return utils.Error{Message:"Проверьте правильность заполнения формы", Errors: e.Errors}
	} else {
		return nil
	}
}

func (account Account) IsVerifiedUser(userId uint) (bool, error) {
	user, err := account.GetUserById(userId)
	if err != nil {
		return false, utils.Error{Message:"Пользователь не найден"}
	}

	methods, err := GetUserVerificationTypeById(account.UserVerificationMethodID)
	if err != nil {
		return false, err
	}

	status := false

	switch methods.Tag {
		case VerificationMethodEmail:
			status = user.EmailVerifiedAt != nil
		case VerificationMethodPhone:
			status = user.PhoneVerifiedAt != nil
		case VerificationMethodEmailAndPhone:
			status = user.EmailVerifiedAt != nil && user.PhoneVerifiedAt != nil
	}


	return status, nil
}

func (account *Account) RemoveUser (user *User) error {
	return db.Model(&user).Association("accounts").Delete(account).Error
}

func (account Account) GetUserRole (user User) (*Role, error) {

	var role Role
	if db.NewRecord(account) || db.NewRecord(user) {
		return nil, errors.New("GetUserRole: Аккаунта или пользователя не существует!")
	}

	aUser, err := account.GetAccountUser(user)
	if err != nil || aUser == nil {
		return nil, err
	}

	if aUser.Role.ID == 0 {
		return nil, errors.New("Не удалось загрузить роль пользователя")
	}
	role = aUser.Role
	
	return &role, nil
}

func (account Account) GetUserAccessRole (user User) (*accessRole, error) {

	if db.NewRecord(account) || db.NewRecord(user) {
		return nil, errors.New("Аккаунта или пользователя не существует!")
	}
	// Сначала получаем общую роль
	role, err := account.GetUserRole(user)
	if err != nil || role == nil || db.NewRecord(role) {
		return nil, err
	}
	aRole := role.Tag
	
	return &aRole, err
}


// Авторизация пользователя со всеми паралельными процессами
func (account Account) AuthUserByEmail(email, password string) (jwt string, err error) {

	// 1. Находим пользователя по email
	//user := GetUserById

	// 2. Проверяем пароль

	// 3. Создаем jwt-


	// 4. Записываем факт авторизации

	// 5. Возвращаем jwt
	return "", nil
}

// выдает пользователю JWT-токен
func (account Account) getUserJwt(userId uint) (jwt string, err error) {
	return "", nil
}


// Дотошно ищет схожего пользователя по username, email и телефону.
func (account Account) existUserUsername(username string) bool {
	if username == "" {
		return false
	}
	return db.Model(&User{}).Where("account_id = ? AND username = ?", account.ID, username).RecordNotFound()
}

func (account Account) existUserEmail(email string) bool {
	if email == "" {
		return false
	}
	return db.Model(&User{}).Where("account_id = ? AND email = ?", account.ID, email).RecordNotFound()
}

func (account Account) existUserPhone(phone string) bool {
	if phone == "" {
		return false
	}
	return db.Model(&User{}).Where("account_id = ? AND phone = ?", account.ID, phone).RecordNotFound()
}

// Возвращает наиболее похожего пользователя (пользователей?) по username, email или телефону в зависимости от типа авторизации


// !!!!!! ### Новая партия на ТЕСТЫ  ### !!!!!!!!!!
func (account *Account) GetToAccount () error {
	return db.First(account, account.ID).Error
}

// сохраняет ВСЕ необходимые поля, кроме id, deleted_at и возвращает в Account обновленные данные
func (account *Account) Save () error {
	return db.Model(&Account{}).Omit("id", "deleted_at").Save(account).Find(account, "id = ?", account.ID).Error
}

// обновляет данные аккаунта кроме id, deleted_at и возвращает в Account обновленные данные
func (account *Account) Update (input interface{}) error {
	return db.Model(&Account{}).Where("id = ?", account.ID).Omit("id", "deleted_at").Update(input).Find(account, "id = ?", account.ID).Error
}

// # Soft Delete
func (account *Account) SoftDelete () error {
	return db.Where("id = ?", account.ID).Delete(account).Error
}

// # Hard Delete
func (account *Account) HardDelete () error {
	return db.Model(&Account{}).Unscoped().Where("id = ?", account.ID).Delete(account).Error
}

// удаляет аккаунт с концами
func (account *Account) DeleteUnscoped () error {
	return db.Model(&Account{}).Where("id = ?", account.ID).Unscoped().Delete(account).Error
}

// ### Account inner func API (+UI) KEYS

func (account *Account) GetApiKeys() error {
	return db.Preload("ApiKeys").First(&account).Error
}

// ### Stock functions ### //
func (account Account) StockCreate(stock *Stock) error {
	stock.AccountID = account.ID
	return stock.Create()
}

func (account *Account) StockLoad() (err error) {
	account.Stocks, err = (Stock{}).GetAll(account.ID)
	return err
}


// ### Account inner func Products ### //
func (account Account) ProductCreate(p *Product) error {
	p.AccountID = account.ID
	return p.Create()
}

func (account *Account) ProductLoad() (err error) {
	account.Products, err = (Product{}).GetAll(account.ID)
	return err
	//return db.Preload("Products").Preload("Products.Offers").First(&a).Error
}

// EAVAttributes
func (account Account) CreateEavAttribute(ea *EavAttribute) error {
	ea.AccountID = account.ID
	return ea.create()
}
