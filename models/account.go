package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"time"
)

type authMethod string

const (
	username  authMethod = "username"
	email authMethod = "email"
	phone authMethod = "phone"
)

/*func (p *authMethod) Scan(value interface{}) error {
	*p = authMethod(value.([]byte))
	return nil
}

func (p authMethod) Value() (driver.Value, error) {
	return string(p), nil
}*/

type Account struct {
	ID uint `json:"id"`

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

	// Тип авторизация
	//AuthMethod authMethod `json:"authBy" gorm:"enum('username', 'email', 'mobilePhone');not null;default:'email';"`
	AuthMethod authMethod `json:"authBy" sql:"type:auth_method;not null;default:'email'"`
	EnabledUserRegistration bool `json:"EnabledUserRegistration" gorm:"default:true;not null"` // Разрешить регистрацию новых пользователей?
	UserRegistrationInvitationOnly bool `json:"UserRegistrationInvitationOnly" gorm:"default:false;not null"` // Регистрация новых пользователей только по приглашению (в том числе и клиентов)



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
	account.CreatedAt = time.Now().UTC()

	//account.UiApiJwtKey =  utils.CreateHS256Key()
	//scope.SetColumn("ui_api_jwt_key", "fjdsfdfsjkfskjfds")
	//scope.SetColumn("ID", uuid.New())
	return nil
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

	outAccount.AuthMethod = account.AuthMethod
	outAccount.EnabledUserRegistration = account.EnabledUserRegistration
	outAccount.UserRegistrationInvitationOnly = account.UserRegistrationInvitationOnly

	outAccount.VisibleToClients = account.VisibleToClients
	outAccount.ClientsAreAllowedToLogin = account.ClientsAreAllowedToLogin

	// Создание аккаунта
	if err := db.Omit("ID").Create(&outAccount).Error; err != nil {
		return nil, err
	}

	return &outAccount, nil
}

// чит функция для развертывания, т.к. нельзя создать аккаунт из-под несуществующего пользователя
func CreateMainAccount() (*Account, error) {

	if !db.Model(&Account{}).First(&Account{}, "id = 1").RecordNotFound() {
		log.Println("RatusCRM account уже существует")
		return nil, nil
	}

	return (Account{
		Name:"RatusCRM",
		UiApiEnabled:false,
		UiApiAesEnabled:true,
		EnabledUserRegistration:false,
		UserRegistrationInvitationOnly:false,
		ApiEnabled: false,
		AuthMethod: username,
	}).create()
}

// проверяем входящие данные для создания или обновления аккаунта и возвращаем описательную ошибку
func (account Account) ValidateInputs() error {

	if len(account.Name) < 2 {
		return utils.Error{Message:"Ошибки в заполнении формы", Errors: map[string]interface{}{"name":"Имя компании должно содержать минимум 2 символа"}}
	}

	if len(account.Name) > 42 { // 256:8 (UTF-8) - хз)
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

func (account Account) CreateUser(input User) (*User, error) {
	input.SignedAccountID = account.ID
	return input.create()
}




// !!! ### Выше функции покрытые тестами ###



func (account *Account) GetToAccount () error {
	return db.First(account, account.ID).Error
}

// сохраняет ВСЕ необходимые поля, кроме id, deleted_at и возвращает в Account обновленные данные
func (account *Account) Save () error {
	return db.Model(Account{}).Omit("id", "deleted_at").Save(account).Find(account, "id = ?", account.ID).Error
}

// обновляет данные аккаунта кроме id, deleted_at и возвращает в Account обновленные данные
func (account *Account) Update (input interface{}) error {
	return db.Model(Account{}).Where("id = ?", account.ID).Omit("id", "deleted_at").Update(input).Find(account, "id = ?", account.ID).Error
}

// # Soft Delete
func (account *Account) SoftDelete () error {
	return db.Where("id = ?", account.ID).Delete(account).Error
}

// # Hard Delete
func (account *Account) HardDelete () error {
	return db.Unscoped().Where("id = ?", account.ID).Delete(account).Error
}

// удаляет аккаунт с концами
func (account *Account) DeleteUnscoped () error {
	return db.Model(Account{}).Where("id = ?", account.ID).Unscoped().Delete(account).Error
}


// ### Account inner USER func
// todo: пересмотреть работу функций под AccountUser

// добавляет пользователя в аккаунт. Если пользователь уже в аккаунте, то ничего не произойдет.
func (account *Account) AppendUser (user *User) error {
	return db.Model(&user).Association("accounts").Append(account).Error
}

func (account *Account) RemoveUser (user *User) error {
	return db.Model(&user).Association("accounts").Delete(account).Error
}

// загружает список обычных пользователей аккаунта
func (account *Account) GetUsers () error {
	return db.Preload("Users").First(&account).Error
}

// ### Account inner func API (+UI) KEYS

func (account *Account) GetApiKeys() error {
	return db.Preload("ApiKeys").First(&account).Error
}



// ### Stock functions
func (account Account) StockCreate(stock *Stock) error {
	stock.AccountID = account.ID
	return stock.Create()
}
func (account *Account) StockLoad() (err error) {
	account.Stocks, err = (Stock{}).GetAll(account.ID)
	return err
}


// ### Account inner func Products
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
