package models

import (
	"errors"
	"fmt"
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
	username authMethod = "username"
	email    authMethod = "email"
	phone    authMethod = "phone"
)

type Account struct {
	ID     uint   `json:"-" gorm:"primary_key"`
	HashID string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный ID для защиты от спама/парсинга

	// данные аккаунта
	Name    string `json:"name" gorm:"type:varchar(255)"`

	// todo: ой бы доработать это все, брат! Ввести типы аккаунтов (первоначальные настройки). А список вебсайтов отдельным объектом.
	Website string `json:"website" gorm:"type:varchar(255)"` // спорно
	Type    string `json:"type" gorm:"type:varchar(255)"`    // нужно имхо

	// API Интерфейс
	ApiEnabled bool `json:"apiEnabled" gorm:"default:true;not null"` // включен ли API интерфейс у аккаунта (false - все ключи отключаются, есть ли смысл в нем?)

	// UI-API Интерфейс (https://ui.api.ratuscrm.com / https://ratuscrm.com/ui-api)
	UiApiEnabled    bool   `json:"uiApiEnabled" gorm:"default:false;not null"`         // Принимать ли запросы через публичный UI-API интерфейсу (через https://ui.api.ratuscrm.com)
	UiApiAesEnabled bool   `json:"uiApiAesEnabled" gorm:"default:true;not null"`          // Включение AES-128/CFB шифрования для публичного UI-API
	UiApiAesKey     string `json:"-" gorm:"type:varchar(16);default:null;"` // 128-битный ключ шифрования
	UiApiJwtKey     string `json:"-" gorm:"type:varchar(32);default:null;"` // 128-битный ключ шифрования

	// Регистрация новых пользователей через UI/API
	UiApiAuthMethods                    pq.StringArray `json:"-" sql:"type:varchar(32)[];default:'{email}'"`  // Доступные способы авторизации (проверяется в контроллере)
	UiApiEnabledUserRegistration        bool           `json:"-" gorm:"default:true;not null"`                // Разрешить регистрацию новых пользователей?
	UiApiUserRegistrationInvitationOnly bool           `json:"-" gorm:"default:false;not null"`               // Регистрация новых пользователей только по приглашению (в том числе и клиентов)
	UiApiUserRegistrationRequiredFields pq.StringArray `json:"-" gorm:"type:varchar(32)[];default:'{email}'"` // список обязательных НЕ нулевых полей при регистрации новых пользователей через UI/API
	UiApiUserEmailDeepValidation        bool           `json:"-" gorm:"default:false;not null"`               // глубокая проверка почты пользователя на предмет существования

	UserVerificationMethodID         uint `json:"-" gorm:"type:int;default:null"` // метод
	UiApiEnabledLoginNotVerifiedUser bool `json:"-" gorm:"default:false;"`        // разрешать ли пользователю входить в аккаунт без завершенной верфикации?

	// Storage
	DiskSpaceAvailable uint `json:"diskSpaceAvailable" gorm:"type:bigint;default:524288000"` // в байтах - общий размер дискового пр-а (def: 500mb)

	// настройки авторизации.
	// Разделяется AppAuth и ApiAuth -
	VisibleToClients         bool `json:"visibleToClients" gorm:"default:false"` // отображать аккаунт в списке доступных для пользователей с ролью 'client'. Нужно для системных аккаунтов.
	ClientsAreAllowedToLogin bool `json:"clientsAreAllowedToLogin" gorm:"default:true"`                 // запрет на вход в ratuscrm для пользователей с ролью 'client' (им не будет выдана авторизация).
	AuthForbiddenForClients bool `json:"-" gorm:"default:true"` // запрет авторизации для для пользователей с ролью 'client'.

	// до этого место принимаются изменения для UPDATE метода
	//ForbiddenForClient bool `json:"forbidden_for_client" gorm:"default:false"` // запрет на вход через приложение app.ratuscrm.com для пользователей с ролью 'client'

	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `json:"-" sql:"index"`

	//Users 		[]User `json:"-" gorm:"many2many:user_accounts"`
	AccountUsers []AccountUser `json:"-"`
	Users   []User   `json:"-" gorm:"many2many:account_users"`
	ApiKeys []ApiKey `json:"-"`

	Products []Product `json:"-"`
	Stocks   []Stock   `json:"-"`
}

// ###
func (Account) PgSqlCreate() {

	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&Account{})
	db.Exec("ALTER TABLE accounts \n--     ADD CONSTRAINT uix_email_account_id_parent_id unique (email,account_id,parent_id),\n    ADD CONSTRAINT accounts_user_verification_method_id_fkey FOREIGN KEY (user_verification_method_id) REFERENCES user_verification_methods(id) ON DELETE CASCADE ON UPDATE CASCADE;\n--     ADD CONSTRAINT users_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,\n--     ALTER COLUMN parent_id SET DEFAULT NULL,\n--     ADD CONSTRAINT users_default_account_id_fkey FOREIGN KEY (default_account_id) REFERENCES accounts(id) ON DELETE SET NULL ON UPDATE CASCADE,    \n--     ADD CONSTRAINT users_invited_user_id_fkey FOREIGN KEY (invited_user_id) REFERENCES users(id) ON DELETE SET NULL ON UPDATE CASCADE;\n\n-- create unique index uix_user_id_account_id_email_parent_id_not_null ON users (account_id,email,parent_id) WHERE parent_id IS NOT NULL;\n-- create unique index uix_account_id_email_parent_id_when_null ON users (account_id,email,parent_id) WHERE parent_id IS NULL;\n")

	// 2. Создаем Главный аккаунт через спец. функцию
	_, err := CreateMainAccount()
	if err != nil {
		log.Fatal("Не удалось создать главный аккаунт. Ошибка: ", err)
	}
}

func (account *Account) BeforeCreate(scope *gorm.Scope) (err error) {

	account.ID = 0
	
	account.HashID = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	account.CreatedAt = time.Now().UTC()

	account.UiApiAesKey, err = utils.CreateAes128Key()
	if err != nil {
		return err
	}
	
	account.UiApiJwtKey = utils.CreateHS256Key()

	//account.UiApiJwtKey =  utils.CreateHS256Key()
	//scope.SetColumn("ui_api_jwt_key", "fjdsfdfsjkfskjfds")
	//scope.SetColumn("ID", uuid.New())
	return nil
}

func (account *Account) Reset() { account = &Account{} }

func (account Account) create() (*Account, error) {
	if err := account.ValidateInputs(); err != nil {
		return nil, err
	}
	err := db.Create(&account).Error
	return &account, err
}

func CreateMainAccount() (*Account, error) {

	// Проверяем есть ли Главны Аккаунт
	_, err := GetMainAccount()
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	dvc, err := GetUserVerificationTypeByCode(VerificationMethodEmailAndPhone)
	if err != nil || dvc == nil {
		return nil, errors.New("Не удалось получить код двойной верификации по телефону и почте")
	}

	acc, err := (Account{
		Name:                                "RatusCRM",
		HashID: "",
		Type: "main",
		UiApiEnabled:                        false,
		UiApiAesEnabled:                     true,
		UiApiEnabledUserRegistration:        false,
		UiApiUserRegistrationInvitationOnly: false,
		ApiEnabled:                          false,
		UiApiAuthMethods:                    pq.StringArray{"username,email,phone"},
		UiApiUserRegistrationRequiredFields: pq.StringArray{"username,email,phone"},

		UserVerificationMethodID:         dvc.ID,
		UiApiEnabledLoginNotVerifiedUser: false,

		VisibleToClients:         false, // клиенты не должны видеть что есть
		AuthForbiddenForClients:  true,  // клиенты должны заходить, но не видить ратус срм в списке
		ClientsAreAllowedToLogin: true,  // клиенты должны заходить, но не видить ратус срм в списке
	}).create()
	if err != nil {
		return nil, err
	}

	return acc, nil
}

func (account Account) ValidateInputs() error {

	if len(account.Name) < 2 {
		return utils.Error{Message: "Ошибки в заполнении формы", Errors: map[string]interface{}{"name": "Имя компании должно содержать минимум 2 символа"}}
	}

	if len(account.Name) > 64 {
		return utils.Error{Message: "Ошибки в заполнении формы", Errors: map[string]interface{}{"name": "Имя компании должно быть не более 42 символов"}}
	}

	if len(account.Website) > 255 {
		return utils.Error{Message: "Ошибки в заполнении формы", Errors: map[string]interface{}{"website": "Слишком длинный url"}}
	}

	if len(account.Type) > 255 {
		return utils.Error{Message: "Ошибки в заполнении формы", Errors: map[string]interface{}{"type": "Слишком длинный текст"}}
	}

	return nil
}

func GetAccount(id uint) (*Account, error) {
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

func (account Account) IsMainAccount() bool {
	return account.ID == 1 && account.Name == "RatusCRM" && account.Type == "main"
}

func GetAccountByHash(hashId string) (*Account, error) {
	var account Account
	err := db.Model(&Account{}).First(&account, "hash_id = ?", hashId).Error
	return &account, err
}

/*func (account *Account) update(input interface{}) error {
	return db.Model(account).Omit("id", "hash_id", "created_at", "updated_at", "deleted_at").Updates(structs.Map(input)).Error
}*/
// Нужны бы проверки на потенциально опасные элементы в обновлении
func (account *Account) Update(input map[string]interface{}) error {
	return db.Model(account).Where("id = ?", account.ID).
		Omit("id", "hash_id", "disk_space_available", "created_at", "updated_at", "deleted_at").
		Update(input).Error
}

func (Account) Exist(id uint) bool {
	return !db.Model(Account{}).First(&Account{}, id).RecordNotFound()
}

// ########### User ###########

func (account Account) CreateUser(input User, v_opt ...accessRole) (*User, error) {

	if account.ID < 1 {
		return nil, errors.New("Не верно указан контекст аккаунта")
	}

	var err error
	var username, email, phone bool
	var role accessRole

	// Утверждаем роль пользователя аккаунта
	if len(v_opt) > 0 {
		role = v_opt[0]
		// нельзя создать пользователя с ролью Owner
		if role == RoleOwner && input.Email != "kokorevn@gmail.com" {
			role = RoleAdmin
		}
	} else {
		role = RoleClient
	}

	// Утверждаем main-account пользователя
	input.IssuerAccountID = account.ID

	// ### !!!! Проверка входящих данных !!! ### ///
	if len(input.Username) > 0 {

		username = true
		if err := utils.VerifyUsername(input.Username); err != nil {
			return nil, utils.Error{Message: "Проверьте правильность заполнения формы", Errors: map[string]interface{}{"username": err.Error()}}
		}
	}

	if len(input.Email) > 0 {
		email = true
		/*if account.UiApiUserEmailDeepValidation {
			if err := utils.EmailDeepValidation(input.Email); err != nil {
				return nil, utils.Error{Message: "Проверьте правильность заполнения формы", Errors: map[string]interface{}{"email": err.Error()}}
			}
		} else {
			if err := utils.EmailValidation(input.Email); err != nil {
				return nil, utils.Error{Message: "Проверьте правильность заполнения формы", Errors: map[string]interface{}{"email": err.Error()}}
			}
		}*/
	}

	if len(input.Phone) > 0 {
		phone = true

		if input.PhoneRegion == "" {
			input.PhoneRegion = "RU"
		}

		// Устанавливаем нужный формат
		input.Phone, err = utils.ParseE164Phone(input.Phone, input.PhoneRegion)
		if err != nil {
			return nil, utils.Error{Message: "Ошибка в формате телефонного номера", Errors: map[string]interface{}{"inviteToken": "Пожалуйста, укажите номер телефона в международном формате"}}
		}

	}

	// 5. One of username. email and phone must be!
	if !(username || email || phone) {
		return nil, utils.Error{Message: "Отсутствуют обязательные поля", Errors: map[string]interface{}{"username": "Необходимо заполнить поле", "email": "Необходимо заполнить поле", "phone": "Необходимо заполнить поле"}}
	}

	// Проверка дублирование полей
	if account.existUserByUsername(input.Username) {
		return nil, utils.Error{Message: "Проверьте правильность заполнения формы", Errors: map[string]interface{}{"username": "Данный username уже используется"}}
	}
	if account.existUserByEmail(input.Email) {

		return nil, utils.Error{Message: "Данные уже есть", Errors: map[string]interface{}{"email": "Этот почтовый адрес уже используется"}}
	}
	if account.existUserByPhone(input.Phone) {
		return nil, utils.Error{Message: "Данные уже есть", Errors: map[string]interface{}{"phone": "Данный телефон уже используется"}}
	}

	// создаем пользователя
	u, err := input.create()
	if err != nil || u == nil {
		return u, err
	}

	// Автоматически добавляем пользователя в аккаунт
	aUser, err := account.AppendUser(*u, role)
	if err != nil || aUser == nil {
		return nil, err
	}
	return u, nil
}

func (account Account) GetUser(userId uint) (*User, error) {

	user, err := User{}.get(userId)
	if err != nil {
		return nil, err
	}

	// Проверим, что пользователь имеет доступ к аккаунта
	aUser := AccountUser{}
	if db.Model(AccountUser{}).First(&aUser, "account_id = ? AND user_id = ?", account.ID, user.ID).RecordNotFound() {
		return nil, errors.New("Пользователь не найден")
	}

	return user, nil
}

func (account Account) GetUserByHashId(hashId string) (*User, error) {
	user, err := User{}.getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	// Проверим, что пользователь имеет доступ к аккаунта
	aUser := AccountUser{}
	if db.Model(AccountUser{}).First(&aUser, "account_id = ? AND user_id = ?", account.ID, user.ID).RecordNotFound() {
		return nil, errors.New("Пользователь не найден")
	}

	return user, nil
}

func (account Account) GetUserByUsername(username string) (*User, error) {

	if username == "" {
		return nil, gorm.ErrRecordNotFound
	}

	user := User{}

	err := db.Model(&User{}).Where("issuer_account_id = ? AND username = ?", account.ID, username).First(&user).Error

	return &user, err
}

func (account Account) GetUserByEmail(email string) (*User, error) {
	if email == "" {
		return nil, gorm.ErrRecordNotFound
	}

	user := User{}

	err := db.Model(&User{}).Where("issuer_account_id = ? AND email = ?", account.ID, email).First(&user).Error

	return &user, err
}

func (account Account) GetUserByPhone(phone, region string) (*User, error) {
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

// pagination user list
func (account Account) GetUserList(offset, limit int, search string) ([]AccountUser, uint, error) {

	aUsers := make([]AccountUser,0)

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"
		
		err := db.Model(&AccountUser{}).Preload("User").
			Limit(limit).
			Offset(offset).
			Joins("LEFT JOIN users ON account_users.user_id = users.id").
			Where("account_id = ?", account.ID).
			Find(&aUsers, "users.username LIKE ? OR users.email LIKE ? OR users.phone LIKE ? OR users.name LIKE ? OR users.surname LIKE ?", search,search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		// WORK!!!!
		//err := db.Model(&User{}).Joins("LEFT JOIN account_users ON account_users.user_id = users.id").Where("account_id = ?", account.ID).Find(&users).Error
		//err := db.Model(&AccountUser{}).Preload("User").Joins("LEFT JOIN users ON account_users.user_id = users.id").Where("account_id = ?", account.ID).Find(&users).Error

		err := db.Model(&AccountUser{}).Preload("User").
			Limit(limit).
			Offset(offset).
			Joins("LEFT JOIN users ON account_users.user_id = users.id").
			Find(&aUsers, "account_id = ?", account.ID).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}
	}

	var count uint
	err := db.Model(&AccountUser{}).Count(&count).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема клиентской базы"}
	}

	return aUsers, count, nil
}

func (Account) ExistUser(user User) bool {
	return !db.Model(&User{}).First(&User{}, user.ID).RecordNotFound()
}

// Тоже, что и ExitUser, только в контексте аккаунта
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
		First(&aUser).Error

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

		aUser, err := acs.create()
		if err != nil || aUser == nil {
			return nil, errors.New("Ошибка при создании AccountUser.")
		}

		return aUser, nil
	}
}

// !!!!!! ### Выше функции покрытые тестами ### !!!!!!!!!!1

// Ищет пользователя, авторизует и в случае успеха возвращает пользователя и jwt-token. issuerAccount - место, где берутся коды
func (account Account) AuthorizationUserByUsername(username, password string, onceLogin,rememberChoice bool, issuerAccount *Account) (user *User, token string, err error) {


	var e utils.Error

	// Проверяем, есть ли вообще такой пользователь в аккаунте
	user, err = account.GetUserByUsername(username)
	if err != nil || user == nil {
		return nil, "", errors.New("Пользователь не найден")
	}

	// Проверяем пароль, чтобы авторизовать пользователя
	if !user.ComparePassword(password) {
		e.AddErrors("password", "Неверный пароль")
	}

	// Если есть какие-то ошибки - сбрасываем автоирзацию
	if e.HasErrors() {
		e.Message = "Проверьте указанные данные"
		return nil, "", e
	}

	if rememberChoice {
		user.DefaultAccountHashId = account.HashID
		updateData := struct {
			DefaultAccountHashId string
		}{account.HashID}
		if err := user.Update(&updateData); err != nil {
			return nil, "", errors.New("Не удалось авторизовать пользователя")
		}
	}

	token, err = account.AuthorizationUser(*user, false, issuerAccount)
	if err != nil || token == "" {
		return nil, "", errors.New("Не удалось авторизовать пользователя")
	}

	return user, token, nil
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
	for _, v := range account.UiApiUserRegistrationRequiredFields {
		switch v {
		case "username":
			if len(input.Username) == 0 {
				e.AddErrors("username", "Поле обязательно к заполнению")
			}

		case "email":
			if len(input.Email) == 0 {
				e.AddErrors("email", "Поле обязательно к заполнению")
			}

		case "phone":
			if len(input.Phone) == 0 {
				e.AddErrors("phone", "Поле обязательно к заполнению")
			}

		}
	}

	if e.HasErrors() {
		return utils.Error{Message: "Проверьте правильность заполнения формы", Errors: e.Errors}
	} else {
		return nil
	}
}

func (account Account) IsVerifiedUser(userId uint) (bool, error) {
	user, err := account.GetUser(userId)
	if err != nil {
		return false, utils.Error{Message: "Пользователь не найден"}
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

func (account *Account) RemoveUser(user *User) error {
	return db.Model(&user).Association("accounts").Delete(account).Error
}

func (account Account) GetUserRole(user User) (*Role, error) {

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

func (account Account) GetUserAccessRole(user User) (*accessRole, error) {

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
func (account Account) existUserByUsername(username string) bool {
	if username == "" {
		return false
	}
	return !db.Model(&User{}).Where("issuer_account_id = ? AND username = ?", account.ID, username).First(&User{}).RecordNotFound()
}

func (account Account) existUserByEmail(email string) bool {
	if email == "" {
		return false
	}
	return !db.Model(&User{}).Where("issuer_account_id = ? AND email = ?", account.ID, email).First(&User{}).RecordNotFound()
}

func (account Account) existUserByPhone(phone string) bool {
	if phone == "" {
		return false
	}
	return !db.Model(&User{}).Where("issuer_account_id = ? AND phone = ?", account.ID, phone).First(&User{}).RecordNotFound()
}

// Возвращает наиболее похожего пользователя (пользователей?) по username, email или телефону в зависимости от типа авторизации

// !!!!!! ### Новая партия на ТЕСТЫ  ### !!!!!!!!!!
func (account *Account) GetToAccount() error {
	return db.First(account, account.ID).Error
}

// сохраняет ВСЕ необходимые поля, кроме id, deleted_at и возвращает в Account обновленные данные
func (account *Account) Save() error {
	return db.Model(&Account{}).Omit("id", "deleted_at").Save(account).Find(account, "id = ?", account.ID).Error
}

// обновляет данные аккаунта кроме id, deleted_at и возвращает в Account обновленные данные
// func (account *Account) Update(input interface{}) error {


// # Soft Delete
func (account *Account) SoftDelete() error {
	return db.Where("id = ?", account.ID).Delete(account).Error
}

// # Hard Delete
func (account *Account) HardDelete() error {
	return db.Model(&Account{}).Unscoped().Where("id = ?", account.ID).Delete(account).Error
}

// удаляет аккаунт с концами
func (account *Account) DeleteUnscoped() error {
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

// ### JWT Crypto ### !!!!!!!!!!1

func (account Account) GetAuthTokenWithClaims(claims JWT) (cryptToken string, err error) {

	if claims.AccountID < 1 || claims.UserID < 1 {
		return "", errors.New("Не удалось обновить ключ безопасности")
	}

	//Create JWT token
	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)
	tokenString, err := token.SignedString([]byte(account.UiApiJwtKey))
	if err != nil {
		return
	}

	// Encode jwt-token
	cryptToken, err = JWT{}.encrypt([]byte(account.UiApiAesKey), tokenString)
	if err != nil {
		return
	}

	return
}

// Просто получает token
func (account Account) GetAuthToken(user User, workAccount Account) (cryptToken string, err error) {
	if account.ID < 1 || user.ID < 1 {
		return "", errors.New("Не удалось обновить ключ безопасности")
	}

	expiresAt := time.Now().UTC().Add(time.Minute * 120).Unix()

	claims := JWT{
		user.ID,
		workAccount.ID,
		account.ID,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AppServer",
		},
	}

	token := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), claims)
	tokenString, err := token.SignedString([]byte(account.UiApiJwtKey))
	if err != nil {
		return "", errors.New("Ошибка создания ключа безопастности")
	}

	// Encode jwt-token
	cryptToken, err = JWT{}.encrypt([]byte(account.UiApiAesKey), tokenString)
	if err != nil {
		return "", errors.New("Ошибка создания ключа безопастности")
	}

	return
}

// Авторизует пользователя в аккаунте
func (account Account) AuthorizationUser(user User, rememberChoice bool, issuerAccount *Account) (cryptToken string, err error) {

	if account.ID < 1 || user.ID < 1 {
		return "", errors.New("Не удалось обновить ключ безопасности")
	}

	// Запоминаем аккаунт для будущих входов
	user.DefaultAccountHashId = account.HashID

	updateData := struct {
		DefaultAccountHashId string
	}{}

	if rememberChoice {
		updateData.DefaultAccountHashId = account.HashID
	} else {
		//updateData.DefaultAccountID = 0
	}

	if err := user.Update(&updateData); err != nil {
		return "", errors.New("Не удалось авторизовать пользователя")
	}

	token, err := issuerAccount.GetAuthToken(user, account)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (account Account) CreateCryptoTokenForUser(user User) (cryptToken string, err error) {

	if account.ID < 1 || user.ID < 1 {
		return "", errors.New("Не удалось обновить ключ безопасности")
	}

	expiresAt := time.Now().UTC().Add(time.Minute * 20).Unix()

	claims := JWT{
		UserID:          user.ID,
		AccountID:       account.ID,
		IssuerAccountID: user.IssuerAccountID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AppServer",
		},
	}

	return account.GetAuthTokenWithClaims(claims)
}

func (account Account) ParseToken(decryptedToken string, claims *JWT) (err error) {

	if account.ID < 1 {
		return errors.New("Ошибка обновления ключа безопастности")
	}
	// получаем библиотечный токен
	token, err := jwt.ParseWithClaims(decryptedToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("JWT: Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(account.UiApiJwtKey), nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return errors.New("Ошибка в обработке ключа безопастности")
	}

	return nil
}

func (account Account) DecryptToken(token string) (tk string, err error) {

	tk, err = JWT{}.decrypt([]byte(account.UiApiAesKey), token)
	return
}

// декодирует token по внутреннему ключу, который берется из параметров аккаунта
func (account Account) ParseAndDecryptToken(cryptToken string) (*JWT, error) {

	var claims JWT // return value

	// AES decrypt
	tokenStr, err := account.DecryptToken(cryptToken)
	if err != nil {
		return nil, err
	}

	// JWT parse
	err = account.ParseToken(string(tokenStr), &claims)
	if err != nil {
		return nil, err
	}
	return &claims, err

}

// ===============================================
