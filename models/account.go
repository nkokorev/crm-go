package models

import (
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type Account struct {
	ID uint `json:"id"`

	// данные аккаунта
	Name string `json:"name" gorm:"type:varchar(255)"`
	Website string `json:"website" gorm:"type:varchar(255)"` // спорно
	Type string `json:"type" gorm:"type:varchar(255)"` // спорно

	// API Интерфейс
	ApiEnabled bool `json:"apiEnabled" gorm:"default:true;not null"`

	// UI-API Интерфейс (https://ui.api.ratuscrm.com / https://ratuscrm.com/ui-api)
	UiApiPublicEnabled bool `json:"uiApiPublicEnabled" gorm:"default:false;not null"` // Возможно ли подклчюение по публичному UI-API интерфейсу (через https://ui.api.ratuscrm.com)
	UiApiAesEnabled bool `json:"uiApiAesEnabled" gorm:"default:true;not null"` // Включение AES-128/CFB шифрования
	UiApiAesKey string `json:"uiApiAesKey" gorm:"type:varchar(16);default:null;"` // 128-битный ключ шифрования
	UiApiJwtKey string `json:"uiApiJwtKey" gorm:"type:varchar(32);default:null;"` // 128-битный ключ шифрования
	UiApiEnabledUserRegistration bool `json:"uiApiEnabledUserRegistration" gorm:"default:true;not null"` // Разрешить регистрацию через UI-API интерфейс
	UiApiUserRegistrationInvitationOnly bool `json:"uiApiUserRegistrationInvitationOnly" gorm:"default:false;not null"` // Регистрация новых пользователей только по приглашению

	// настройки авторизации.
	// Разделяется AppAuth и ApiAuth -
	VisibleToClients bool `json:"visibleToClients" gorm:"default:true"` // скрывать аккаунт в списке доступных для пользователей с ролью 'client'. Нужно для системных аккаунтов.
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

// создает аккаунт
func CreateAccount (a Account) (_ *Account, err error) {

	// Создаем ключи для UI API
	a.UiApiAesKey, err = utils.CreateAes128Key()
	if err != nil {
		return nil, err
	}

	a.UiApiJwtKey =  utils.CreateHS256Key()


	// Создание аккаунта
	if err := db.Omit("ID", "DeletedAt").Create(&a).Error; err != nil {
		return nil, err
	}

	return &a, nil
}

func (account *Account) CreateToAccount () error {

	// Верификация данных

	// Создание аккаунта
	if err := db.Create(account).Error; err != nil {
		return err
	}

	return nil
}

// осуществляет поиск по a.ID
func GetAccount (id uint) (a Account, err error) {
	err = db.Model(&Account{}).First(&a, id).Error
	return a, err
}

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

// # Delete
func (account *Account) Delete () error {
	return db.Model(Account{}).Where("id = ?", account.ID).Delete(account).Error
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



func (account Account) CreateApiKey() (*ApiKey, error) {

	// 1. Привязываем к аккаунту
	key := &ApiKey{AccountID:account.ID, Name:"Test api key", Status:true}

	// 2. Создаем
	if err := key.create(); err != nil {
		return nil, err
	}

	return key, nil
}

func (account *Account) GetApiKeys() error {
	return db.Preload("ApiKeys").First(&account).Error
}

func (account *Account) DeleteApiKey(key *ApiKey) error {
	return key.delete()
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