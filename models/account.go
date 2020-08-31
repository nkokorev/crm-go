package models

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
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
	Id     	uint   `json:"id" gorm:"primaryKey"`
	HashId 	string `json:"hash_id" gorm:"type:varchar(12);unique_index;not null;"` // публичный Id для защиты от спама/парсинга

	Name    string `json:"name" gorm:"type:varchar(255)"`

	// todo: ой бы доработать это все, брат! Ввести типы аккаунтов (первоначальные настройки). А список вебсайтов отдельным объектом.
	Website string `json:"website" gorm:"type:varchar(255)"` // спорно
	Type    string `json:"type" gorm:"type:varchar(255)"`    // нужно имхо

	// API Интерфейс
	ApiEnabled bool `json:"api_enabled" gorm:"default:true;not null"` // включен ли API интерфейс у аккаунта (false - все ключи отключаются, есть ли смысл в нем?)

	// UI-API Интерфейс (https://ui.api.ratuscrm.com / https://ratuscrm.com/ui-api)
	UiApiEnabled    bool   `json:"ui_api_enabled" gorm:"default:false;not null"`        // Принимать ли запросы через публичный UI-API интерфейсу (через https://ui.api.ratuscrm.com)
	UiApiAesEnabled bool   `json:"ui_api_aes_enabled" gorm:"default:true;not null"`     // Включение AES-128/CFB шифрования для публичного UI-API
	UiApiAesKey     string `json:"-" gorm:"type:varchar(16);default:null;"` 			// 128-битный ключ шифрования
	UiApiJwtKey     string `json:"-" gorm:"type:varchar(32);default:null;"` 			// 128-битный ключ шифрования

	// Регистрация новых пользователей через UI/API
	// UiApiAuthMethods                    datatypes.JSON `json:"-" sql:"type:varchar(32)[];default:'{email}'"`  // Доступные способы авторизации (проверяется в контроллере)
	UiApiAuthMethods                    datatypes.JSON `json:"-" `  // Доступные способы авторизации (проверяется в контроллере)
	UiApiEnabledUserRegistration        bool           `json:"-" gorm:"default:true;not null"`                // Разрешить регистрацию новых пользователей?
	UiApiUserRegistrationInvitationOnly bool           `json:"-" gorm:"default:false;not null"`               // Регистрация новых пользователей только по приглашению (в том числе и клиентов)
	UiApiUserRegistrationRequiredFields datatypes.JSON `json:"-" ` // список обязательных НЕ нулевых полей при регистрации новых пользователей через UI/API
	UiApiUserEmailDeepValidation        bool           `json:"-" gorm:"default:false;not null"`               // глубокая проверка почты пользователя на предмет существования

	UserVerificationMethodId         uint `json:"-" gorm:"type:int;default:null"` // метод
	UiApiEnabledLoginNotVerifiedUser bool `json:"-" gorm:"default:false;"`        // разрешать ли пользователю входить в аккаунт без завершенной верфикации?

	// Storage
	DiskSpaceAvailable 	int64 `json:"disk_space_available" gorm:"type:bigint;default:524288000"` // в байтах - общий размер дискового пр-а (def: 500mb)

	// настройки авторизации.
	// Разделяется AppAuth и ApiAuth -
	VisibleToClients         bool `json:"visible_to_clients" gorm:"default:false"` // отображать аккаунт в списке доступных для пользователей с ролью 'client'. Нужно для системных аккаунтов.
	ClientsAreAllowedToLogin bool `json:"clients_are_allowed_to_login" gorm:"default:true"`                 // запрет на вход в ratuscrm для пользователей с ролью 'client' (им не будет выдана авторизация).
	AuthForbiddenForClients  bool `json:"-" gorm:"default:true"` // запрет авторизации для для пользователей с ролью 'client'.

	// до этого место принимаются изменения для UPDATE метода
	//ForbiddenForClient bool `json:"forbidden_for_client" gorm:"default:false"` // запрет на вход через приложение app.ratuscrm.com для пользователей с ролью 'client'

	CreatedAt 	time.Time  `json:"-"`
	UpdatedAt 	time.Time  `json:"-"`
	// DeletedAt *time.Time `json:"-" sql:"index"`
	DeletedAt 	gorm.DeletedAt `json:"-" sql:"index"`

	//Users 		[]User `json:"-" gorm:"many2many:user_accounts"`
	AccountUsers []AccountUser `json:"-"`
	Users   	[]*User   	`json:"-" gorm:"many2many:account_users"`
	ApiKeys 	[]ApiKey 	`json:"-" gorm:"-"`
	Products 	[]*Product 	`json:"-" gorm:"-"`
	// Stocks   []Stock   `json:"-"`
}

// ###
func (Account) PgSqlCreate() {
	if db.Migrator().HasTable(&Account{}) { return }

	// 1. Создаем таблицу и настройки в pgSql
	if err := db.Migrator().AutoMigrate(&Account{}); err != nil {log.Fatal(err)}
	// db.CreateTable(&Account{})

	// 2. Создаем Главный аккаунт через спец. функцию
	_, err := CreateMainAccount()
	if err != nil {
		log.Fatal("Не удалось создать главный аккаунт. Ошибка: ", err)
	}
}

func (account *Account) BeforeCreate(tx *gorm.DB) (err error) {

	account.Id = 0
	
	account.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	account.CreatedAt = time.Now().UTC()

	account.UiApiAesKey, err = utils.CreateAes128Key()
	if err != nil {
		return err
	}
	
	account.UiApiJwtKey = utils.CreateHS256Key()

	//account.UiApiJwtKey =  utils.CreateHS256Key()
	//scope.SetColumn("ui_api_jwt_key", "fjdsfdfsjkfskjfds")
	//scope.SetColumn("Id", uuid.New())
	return nil
}

func (account *Account) Reset() { account = &Account{} }

func (account Account) create() (*Account, error) {
	if err := account.ValidateInputs(); err != nil {
		return nil, err
	}
	acc := account
	if err := db.Create(&acc).Error; err != nil { return nil, err }

	return &acc, nil
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

	// ar := `["username","email","phone"]`
	// ar := `{}`
	// user1Attrs := `{"age":18,"name":"json-1","orgs":{"orga":"orga"},"tags":["tag1","tag2"]}`
	
	acc, err := (Account{
		Name:                                "RatusCRM",
		HashId: "",
		Type: "main",
		UiApiEnabled:                        false,
		UiApiAesEnabled:                     true,
		UiApiEnabledUserRegistration:        false,
		UiApiUserRegistrationInvitationOnly: false,
		ApiEnabled:                          false,
		UiApiAuthMethods:                    datatypes.JSON(utils.StringArrToRawJson([]string{"username","email","phone"})),
		// UiApiUserRegistrationRequiredFields: datatypes.JSON(utils.StringArrToRawJson([]string{"username","email","phone"})),
		UiApiUserRegistrationRequiredFields: datatypes.JSON(utils.StringArrToRawJson([]string{"username","email","phone"})),

		UserVerificationMethodId:         dvc.Id,
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
	if err := db.First(&account, "id = 1 AND name = 'RatusCRM'").Error; err != nil { return nil, err}

	return &account, nil
}

func (account Account) IsMainAccount() bool {
	return account.Id == 1 && account.Name == "RatusCRM" && account.Type == "main"
}

func GetAccountByHash(hashId string) (*Account, error) {
	var account Account
	err := db.Model(&Account{}).First(&account, "hash_id = ?", hashId).Error
	return &account, err
}

// Нужны бы проверки на потенциально опасные элементы в обновлении
func (account *Account) Update(input map[string]interface{}) error {
	return db.Model(account).Where("id = ?", account.Id).
		Omit("id", "hash_id", "disk_space_available", "created_at", "updated_at", "deleted_at").
		Updates(input).Error
}

func (Account) Exist(id uint) bool {
	if err := db.Model(Account{}).First(&Account{}, id).Error; err != nil { return false }
	return true
}

// ########### User ###########

func (account Account) CreateUser(input User, role Role) (*User, error) {

	if account.Id < 1 {
		return nil, errors.New("Не верно указан контекст аккаунта")
	}

	var err error
	var username, email, phone bool

	// нельзя создать пользователя с ролью Owner
	/*if role == RoleOwner && input.Email != "kokorevn@gmail.com" {
		role = RoleAdmin
	}*/

	// Утверждаем main-account пользователя
	input.IssuerAccountId = account.Id

	// ### !!!! Проверка входящих данных !!! ### ///
	if input.Username != nil {

		username = true
		if err := utils.VerifyUsername(input.Username); err != nil {
			return nil, utils.Error{Message: "Проверьте правильность заполнения формы", Errors: map[string]interface{}{"username": err.Error()}}
		}
	}

	if input.Email != nil {
		email = true
		if err := utils.EmailValidation(*input.Email); err != nil {
			return nil, utils.Error{Message: "Проверьте правильность заполнения формы", Errors: map[string]interface{}{"email": err.Error()}}
		}
	}

	if input.Phone != nil {
		phone = true

		if *input.PhoneRegion == "" {
			input.PhoneRegion = utils.STRp("RU")
		}

		// Устанавливаем нужный формат
		_phone, err := utils.ParseE164Phone(*input.Phone, *input.PhoneRegion)
		if err != nil {
			return nil, utils.Error{Message: "Ошибка в формате телефонного номера", Errors: map[string]interface{}{"inviteToken": "Пожалуйста, укажите номер телефона в международном формате"}}
		}
		input.Phone = utils.STRp(_phone)

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
	user, err := input.create()
	if err != nil || user == nil {
		return user, err
	}

	// Автоматически добавляем пользователя в аккаунт
	_, err = account.AppendUser(*user, role)
	if err != nil {
		return nil, err
	}
	user, err = account.GetUserWithAUser(user.Id)
	if err != nil {
		return nil, err
	}

	return user, nil
}
func (account Account) UploadUsers(users []User, role Role) error {

	if account.Id < 1 {
		return errors.New("Не верно указан контекст аккаунта")
	}

	for _, user := range users {
		var err error
		var username, email, phone bool

		// Утверждаем main-account пользователя
		user.IssuerAccountId = account.Id

		// ### !!!! Проверка входящих данных !!! ### ///
		if len(*user.Username) > 0 {

			username = true
			if err := utils.VerifyUsername(user.Username); err != nil {
				continue
			}
		}

		if len(*user.Email) > 0 {
			email = true
			// if err := utils.EmailValidation(user.Email); err != nil {
			if err := utils.EmailValidation(*user.Email); err != nil {
				continue
			}
		}

		if len(*user.Phone) > 0 {
			phone = true

			if *user.PhoneRegion == "" {
				user.PhoneRegion = utils.STRp("RU")
			}

			// Устанавливаем нужный формат
			_phone, err := utils.ParseE164Phone(*user.Phone, *user.PhoneRegion)
			if err != nil {
				continue
			}
			user.Phone = utils.STRp(_phone)

		}

		// 5. One of username. email and phone must be!
		if !(username || email || phone) {
			continue
		}

		// Проверка дублирование полей
		if account.existUserByUsername(user.Username) {
			continue
		}
		if account.existUserByEmail(user.Email) {
			continue
		}
		if account.existUserByPhone(user.Phone) {
			continue
		}

		// создаем пользователя
		user, err := user.create()
		if err != nil || user == nil {
			continue
		}

		// Автоматически добавляем пользователя в аккаунт
		_, err = account.AppendUser(*user, role)
		if err != nil {
			continue
		}
	}

	return nil
}

// Возвращает пользователя везде, кроме главного аккаунта
func (account Account) GetUser(userId uint) (*User, error) {

	user, err := User{}.get(userId)
	if err != nil {
		return nil, err
	}

	// Проверим, что пользователь имеет доступ к аккаунту
	aUser := AccountUser{}
	if err := db.Model(AccountUser{}).First(&aUser, "account_id = ? AND user_id = ?", account.Id, user.Id).Error;err != nil {
		return nil, errors.New("Пользователь не найден")
	}


	return user, nil
}

func (account Account) GetUserWithAUser(userId uint) (*User, error) {

	var user User

	if err := db.Preload("AccountUser", func(db *gorm.DB) *gorm.DB {
		return db.Where("account_id = ?", account.Id).Select(AccountUser{}.SelectArrayWithoutBigObject())
	}).First(&user, userId).Error; err != nil { return nil, err }

	return &user, nil
}

func (account Account) GetUserByHashId(hashId string) (*User, error) {
	user, err := User{}.getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	// Проверим, что пользователь имеет доступ к аккаунта
	aUser := AccountUser{}
	if err := db.Model(AccountUser{}).First(&aUser, "account_id = ? AND user_id = ?", account.Id, user.Id).Error;err != nil {
		return nil, errors.New("Пользователь не найден")
	}

	return user, nil
}

// Получает пользователя в аккаунте, в том числе, если он просто клиент (не его issuer account)
func (account Account) GetUserByUsername(username string) (*User, error) {

	if username == "" {
		return nil, gorm.ErrRecordNotFound
	}

	// Просто ищем пользователя с таким username
	user, err := User{}.GetByUsername(username)
	if err != nil {
		return nil, utils.Error{Message: "Пользователь не найден"}
	}

	if !account.AccessUserById(user.Id) {
		return nil, utils.Error{Message: "Пользователь не найден"}
	}

	return user, err
}

// Получает пользователя в аккаунте, в том числе в RatusCRM, если он не является его прямым клиентом
func (account Account) GetUserForAuthAppByUsername(username string) (*User, error) {

	if username == "" { return nil, gorm.ErrRecordNotFound }

	// Просто ищем пользователя с таким username
	user, err := User{}.GetByUsername(username)
	if err != nil {
		return nil, utils.Error{ Message: "Пользователь не найден" }
	}

	// 2. Проверяем, имеет ли он доступ к целевому аккаунту
	if account.IsMainAccount() && user.IssuerAccountId != 1 {
		// если логинится в главном аккаунте и ему не разрешен доступ
	   if !user.EnabledAuthFromApp {
		   return nil, utils.Error{Message: "Вход через RatusCRM не разрешен"}
	   }

	} else {
		// Если это не RatusCRM аккаунт, то проверяем через AccountUser есть ли его Id
		if !account.AccessUserById(user.Id) {
			return nil, utils.Error{Message: "Пользователь не найден"}
		}
	}

	return user, err
}

// todo переписать функции
func (account Account) GetUserByEmail(email string) (*User, error) {
	if email == "" {
		return nil, gorm.ErrRecordNotFound
	}

	var user User

	err := db.Table("users").Joins("LEFT JOIN account_users ON account_users.user_id = users.id").
		Select("account_users.public_id, account_users.account_id, users.*").
		Where("account_users.account_id = ? AND users.email = ?", account.Id, email).
		First(&user).Error
	if err != nil {
		return nil, err
	}

	return &user, err
}

// todo переписать функции
func (account Account) GetUserByPhone(phone string, regions... string) (*User, error) {
	
	if phone == "" {
		return nil, gorm.ErrRecordNotFound
	}

	region := "RU"
	if len(regions) > 0 {
		region = regions[0]
	}

	_phone, _ := utils.ParseE164Phone(phone, region)

	var user User

	err := db.Table("users").Joins("LEFT JOIN account_users ON account_users.user_id = users.id").
		Select("account_users.public_id, account_users.account_id, users.*").
		Where("account_users.account_id = ? AND phone = ? AND phone_region = ?", account.Id, _phone, region).
		First(&user).Error
	if err != nil {
		// fmt.Println("Пользователь не найден по телефону! ", err)
		return nil, err
	}
	
	return &user, err
}


// pagination user list, учитывая роли по списку id
func (account Account) GetUsersByList(list []uint, sortBy string) ([]User, int64, error) {

	users := make([]User,0)
	var total int64


	err := db.Table("users").Joins("LEFT JOIN account_users ON account_users.user_id = users.id").
		Select("account_users.public_id, account_users.account_id, account_users.role_id, users.*").Order(sortBy).
		Where("account_users.account_id = ? AND users.id IN (?)", account.Id, list).Preload("Roles").
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	// Вычисляем total
	err = db.Model(&User{}).Joins("LEFT JOIN account_users ON account_users.user_id = users.id").
		Where("account_id = ? AND id IN (?)", account.Id, list).
		Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема клиентской базы"}
	}

	return users, total, nil
}

func (account Account) GetUserListPagination(offset, limit int, sortBy, search string, role []uint) ([]User, int64, error) {

	users := make([]User,0)
	var total int64

	if len(search) > 0 {

		search = "%"+search+"%"
		
		err := db.Table("users").Joins("left join account_users ON account_users.user_id = users.id").
			Select("account_users.account_id, account_users.role_id, users.*").
			Where("account_users.account_id = ? AND account_users.role_id IN (?)", account.Id, role).
			Order(sortBy).Offset(offset).Limit(limit).
			Find(&users, "hash_id ILIKE ? OR username ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR name ILIKE ? OR surname ILIKE ? OR patronymic ILIKE ?", search,search,search,search,search,search,search).Error
		if err != nil {
			return nil, 0, err
		}

		for i := range users {
			var aUser AccountUser
			err = db.Model(&aUser).Where("account_id = ? AND user_id = ?", account.Id, users[i].Id).First(&aUser).Error
			if err != nil {
				return nil, 0, err
			}
			users[i].AccountUser = &aUser
		}

		// Вычисляем total
		err = db.Model(&User{}).Joins("LEFT JOIN account_users ON account_users.user_id = users.id").
			Where("account_id = ? AND role_id IN (?)", account.Id, role).
			Where("hash_id ILIKE ? OR username ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR name ILIKE ? OR surname ILIKE ? OR patronymic ILIKE ?", search,search,search,search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема клиентской базы"}
		}

	} else {

		err := db.Table("users").Joins("left join account_users ON account_users.user_id = users.id").
			Select("account_users.account_id, account_users.role_id, users.*").
			Where("account_users.account_id = ? AND account_users.role_id IN (?)", account.Id, role).
			Order(sortBy).Offset(offset).Limit(limit).
			Find(&users).Error
		if err != nil {
			return nil, 0, err
		}

		for i := range users {
			var aUser AccountUser
			err = db.Model(&aUser).Where("account_id = ? AND user_id = ?", account.Id, users[i].Id).First(&aUser).Error
			if err != nil {
				return nil, 0, err
			}
			users[i].AccountUser = &aUser
		}

		// Вычисляем total
		if len(users) > 0 {
			err = db.Model(&User{}).Joins("LEFT JOIN account_users ON account_users.user_id = users.id").
				Where("account_id = ? AND role_id IN (?)", account.Id, role).
				Count(&total).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return nil, 0, utils.Error{Message: "Ошибка определения объема клиентской базы"}
			}
		}

	}
	

	return users, total, nil
}

func (account Account) ExistUser(user User) bool {

	// return !db.Model(&User{}).First(&User{}, user.Id).RecordNotFound()
	err := db.Model(&User{}).First(&User{}, user.Id).Error
	if err != nil {
		return false
	}

	return true
}

// Тоже, что и ExitUser, только в контексте аккаунта
func (account Account) ExistAccountUser(userId uint) bool {

	// var aUser AccountUser
	var count int64
	err := db.Model(&AccountUser{}).Where("account_id = ? AND user_id = ?", account.Id, userId).Count(&count).Error
	if err != nil {
		return false
	}
	if count > 0 { return true }
	// fmt.Println("count: ", count)
	// fmt.Println("Пользователь есть!")
	return false
}

// Если пользователь не найден - вернет gorm.ErrRecordNotFound
func (account Account) GetAccountUser(userId uint) (*AccountUser, error) {

	aUser := AccountUser{}

	if account.Id < 1 {
		return nil, errors.New("Аккаунта или пользователя не существует!")
	}

	err := db.Model(&AccountUser{}).
		Where("account_id = ? AND user_id = ?", account.Id, userId).
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

func (account Account) GetAccountUserByUsername(username string) (*AccountUser, error) {

	aUser := AccountUser{}

	if account.Id < 1 {
		return nil, errors.New("Аккаунта или пользователя не существует!")
	}

	err := db.Model(&AccountUser{}).
		Where("account_id = ? AND user_id = ?", account.Id, username).
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
func (account Account) AppendUser(user User, role Role) (*AccountUser, error) {

	acs := AccountUser{} // return value

	if user.Id < 1 || !account.ExistUser(user) {
		return nil, errors.New("Необходимо создать сначала пользователя!")
	}
	
	// проверяем, относится ли пользователь к аккаунту
	if account.ExistAccountUser(user.Id) {
		// обновляем роль
		// todo дописать..
		
		return nil, errors.New("Невозможно добавить пользователя в аккаунт, т.к. он в нем уже есть.")

	} else {

		_asc, err := account.SetUserRole(&user, role)
		if err != nil || _asc == nil {
			return nil, errors.New("Ошибка при добавлении пользователя в аккаунт")
		}

		acs = *_asc

		event.AsyncFire(Event{}.UserAppendedToAccount(account.Id, acs.UserId, acs.RoleId))
	}

	return &acs, nil
}

// !!!!!! ### Выше функции покрытые тестами ### !!!!!!!!!!1

// Ищет пользователя, авторизует и в случае успеха возвращает пользователя и jwt-token. issuerAccount - место, где берутся коды
// в app.ratuscrm.com контекст - RatusCRM, accountId = 1
func (account Account) AuthorizationUserByUsername(username, password string, onceLogin,rememberChoice bool, issuerAccount *Account) (*User, string, error) {

	var e utils.Error
	
	// Проверяем, можем ли мы выдать временную авторизацию в аккаунте (не то же самое, что войти в аккаунт!)
	user, err := account.GetUserForAuthAppByUsername(username)
	if err != nil {
		return nil, "", err
	}

	// 3. Проверяем пароль, чтобы авторизовать пользователя
	if !user.ComparePassword(password) {
		e.AddErrors("password", "Неверный пароль")
	}

	// Если есть какие-то ошибки - сбрасываем авторизацию
	if e.HasErrors() {
		e.Message = "Проверьте указанные данные"
		return nil, "", e
	}

	token, err := account.AuthorizationUser(*user, false, issuerAccount)
	if err != nil || token == "" {
		return nil, "", errors.New("Не удалось авторизовать пользователя")
	}

	return user, token, nil
}

// проверяет, имеет ли указанный пользователь доступ к аккаунту
func (account Account) AccessUserById(userId uint) bool {
	if userId < 1 {return false}
	// fmt.Printf("issuer_account_id = %v AND email = %v\n", account.Id, email)
	err := db.Model(&AccountUser{}).Where("account_id = ? AND user_id = ?", account.Id, userId).First(&AccountUser{}).Error
	if err != nil {
		return false
	}

	return true
}

// *** New functions ****

// Обязательно ли поле при создании пользователя (username, email, phone)
func (account Account) userRequiredField(field string) bool {
	jsonData := utils.ParseJSONBToString(account.UiApiUserRegistrationRequiredFields)

	for _, v := range jsonData {
		if v == field {
			return true
		}
	}
	return false
}

// Проверяет поля в input на не нулевость в соответствие настройкам аккаунта
func (account Account) ValidationUserRegReqFields(input User) error {
	var e utils.Error
	jsonData := utils.ParseJSONBToString(account.UiApiUserRegistrationRequiredFields)
	for _, v := range jsonData {
		switch v {
		case "username":
			if len(*input.Username) == 0 {
				e.AddErrors("username", "Поле обязательно к заполнению")
			}

		case "email":
			if len(*input.Email) == 0 {
				e.AddErrors("email", "Поле обязательно к заполнению")
			}

		case "phone":
			if len(*input.Phone) == 0 {
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

	methods, err := GetUserVerificationTypeById(account.UserVerificationMethodId)
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

// !!! удаляет пользователя soft !!!
func (account *Account) DeleteUser(user *User) error {

	if user.IssuerAccountId != account.Id {
		return utils.Error{Message: "Невозможно удалить пользователя т.к. он привязан к другому аккаунту"}
	}
	if err := db.Model(&user).Association("accounts").Delete(account); err != nil {
		return err
	}

	if err := user.delete(); err != nil {
		return err
	}

	event.AsyncFire(Event{}.UserDeleted(account.Id, user.Id))

	return nil
}

// !!! Если Issuer аккаунт удаляет пользователя - то он совсем soft delete()
func (account *Account) RemoveUser(user *User) error {
	if user.IssuerAccountId == account.Id {
		if err := user.delete(); err != nil {
			return err
		}
	} else {
	 	if err := db.Model(&user).Association("Accounts").Delete(account); err != nil {
	 		return err
		}
	}

	if err := db.Model(&user).Association("Accounts").Delete(account); err != nil {
		fmt.Println(err)
		return err
	}

	event.AsyncFire(Event{}.UserRemovedFromAccount(account.Id, user.Id))

	return nil
}

func (account Account) GetUserRole(user User) (*Role, error) {

	var role Role
	if account.Id < 1 || user.Id < 1 {
		return nil, errors.New("GetUserRole: Аккаунта или пользователя не существует!")
	}

	aUser, err := account.GetAccountUser(user.Id)
	if err != nil || aUser == nil {
		return nil, err
	}

	if aUser.Role.Id == 0 {
		return nil, errors.New("Не удалось загрузить роль пользователя")
	}
	role = aUser.Role

	return &role, nil
}

func (account Account) GetUserAccessRole(user User) (*AccessRole, error) {

	if account.Id < 1 || user.Id < 1 {
		return nil, errors.New("Аккаунта или пользователя не существует!")
	}
	// Сначала получаем общую роль
	role, err := account.GetUserRole(user)
	if err != nil || role == nil || role.Id < 1 {
		return nil, err
	}
	aRole := role.Tag

	return &aRole, err
}

// Обновление роли пользователя для текущего аккаунта
func (account Account) UpdateUserRole(user *User, role Role) error {

	if account.Id < 1 || user.Id < 1 {
		return errors.New("GetUserRole: Аккаунта или пользователя не существует!")
	}

	// Проверяем, есть ли
	aUser, err := account.GetAccountUser(user.Id)
	if err != nil || aUser == nil {
		return err
	}

	err = aUser.update(map[string]interface{}{"roleId":role.Id})
	if err != nil {
		return err
	}

	return nil
}

func (account Account) SetUserRole(user *User, role Role) (*AccountUser, error) {

	if account.Id < 1 || user.Id < 1 {
		return nil, errors.New("GetUserRole: Аккаунта или пользователя не существует!")
	}

	aUser := AccountUser{AccountId: account.Id, RoleId: role.Id, UserId: user.Id}

	err := db.Model(&AccountUser{}).Where("account_id = ? AND user_id = ?", account.Id, user.Id).
		FirstOrCreate(&aUser).
		Updates(map[string]interface{}{"role_id":role.Id}).Find(&aUser).Error
	if err != nil {
		fmt.Println("Ошибка: ", err)
		return nil, err
	}

	return &aUser, nil
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
func (account Account) existUserByUsername(username *string) bool {
	if username == nil {
		return false
	}
	if err := db.Model(&User{}).Where("issuer_account_id = ? AND username = ?", account.Id, username).First(&User{}).Error;err != nil {
		return false
	} else {
		return true
	}
}

func (account Account) existUserByEmail(email *string) bool {
	if email == nil {
		return false
	}
	// fmt.Printf("issuer_account_id = %v AND email = %v\n", account.Id, email)
	if err := db.Model(&User{}).Where("issuer_account_id = ? AND email = ?", account.Id, email).First(&User{}).Error; err != nil {
		return false
	} else {
		return true
	}
}

func (account Account) existUserByPhone(phone *string) bool {
	if phone == nil {
		return false
	}
	if err := db.Model(&User{}).Where("issuer_account_id = ? AND phone = ?", account.Id, phone).First(&User{}).Error; err != nil {
		return false
	} else {
		return true
	}
}

// Возвращает наиболее похожего пользователя (пользователей?) по username, email или телефону в зависимости от типа авторизации

// !!!!!! ### Новая партия на ТЕСТЫ  ### !!!!!!!!!!
func (account *Account) GetToAccount() error {
	return db.First(account, account.Id).Error
}

// сохраняет ВСЕ необходимые поля, кроме id, deleted_at и возвращает в Account обновленные данные
func (account *Account) Save() error {
	return db.Model(&Account{}).Omit("id", "deleted_at","created_at").Save(account).Find(account, "id = ?", account.Id).Error
}

// обновляет данные аккаунта кроме id, deleted_at и возвращает в Account обновленные данные
// func (account *Account) Update(input interface{}) error {


// # Soft Delete
func (account *Account) SoftDelete() error {
	return db.Where("id = ?", account.Id).Delete(account).Error
}

// # Hard Delete
func (account *Account) HardDelete() error {
	return db.Model(&Account{}).Unscoped().Where("id = ?", account.Id).Delete(account).Error
}

// удаляет аккаунт с концами
func (account *Account) DeleteUnscoped() error {
	return db.Model(&Account{}).Where("id = ?", account.Id).Unscoped().Delete(account).Error
}

// ### Account inner func API (+UI) KEYS

func (account *Account) GetApiKeys() error {
	return db.Preload("ApiKeys").First(&account).Error
}

// ### JWT Crypto ### !!!!!!!!!!1

func (account Account) GetAuthTokenWithClaims(claims JWT) (cryptToken string, err error) {

	if claims.AccountId < 1 || claims.UserId < 1 {
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
	if account.Id < 1 || user.Id < 1 {
		return "", errors.New("Не удалось обновить ключ безопасности")
	}

	expiresAt := time.Now().UTC().Add(time.Minute * 120).Unix()

	claims := JWT{
		user.Id,
		workAccount.Id,
		account.Id,
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

	if account.Id < 1 || user.Id < 1 {
		return "", errors.New("Не удалось обновить ключ безопасности")
	}

	// Запоминаем аккаунт для будущих входов
	// user.DefaultAccountHashId = account.HashId

	/*updateData := struct {
		DefaultAccountHashId string
	}{}

	if rememberChoice {
		updateData.DefaultAccountHashId = account.HashId
	} else {
		//updateData.DefaultAccountId = 0
	}

	if err := user.update(updateData); err != nil {
		return "", errors.New("Не удалось авторизовать пользователя")
	}*/

	token, err := issuerAccount.GetAuthToken(user, account)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (account Account) CreateCryptoTokenForUser(user User) (cryptToken string, err error) {

	if account.Id < 1 || user.Id < 1 {
		return "", errors.New("Не удалось обновить ключ безопасности")
	}

	expiresAt := time.Now().UTC().Add(time.Minute * 20).Unix()

	claims := JWT{
		UserId:          user.Id,
		AccountId:       account.Id,
		IssuerAccountId: user.IssuerAccountId,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AppServer",
		},
	}

	return account.GetAuthTokenWithClaims(claims)
}

func (account Account) ParseToken(decryptedToken string, claims *JWT) (err error) {

	if account.Id < 1 {
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

func (account Account) GetDepersonalizedData() interface{} {
	return &account
}