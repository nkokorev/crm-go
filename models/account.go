package models

import "time"

type Account struct {
	ID uint `json:"id"`

	// данные аккаунта
	Name string `json:"name" gorm:"type:varchar(255)"`
	Website string `json:"website" gorm:"type:varchar(255)"` // спорно
	Type string `json:"type" gorm:"type:varchar(255)"` // спорно

	// Настройки аккаунта

	// UI-API Интерфейс

	UiApiEnabled bool `json:"uiApiEnabled" gorm:"default:false;"` // Принимать ли api запросы по адресу https://ui.api.ratuscrm.com
	UiApiEnabledUserRegistration bool `json:"uiApiEnabledUserRegistration" gorm:"default:true"` // Включить регистрацию через UI-API интерфейс

	UserRegistrationAllow bool `json:"-" gorm:"default:false"` // разрешено ли создавать новых пользователей
	UserRegistrationInviteOnly bool `json:"userRegistrationInviteOnly" gorm:"default:false;"`

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

// создает аккаунт
func (a *Account) Create () error {

	// Верификация данных

	// Создание аккаунта
	if err := db.Create(a).Error; err != nil {
		return err
	}



	return nil
}

// осуществляет поиск по a.ID
func (a *Account) Get () error {
	return db.First(a,a.ID).Error
}

// сохраняет ВСЕ необходимые поля, кроме id, deleted_at и возвращает в Account обновленные данные
func (a *Account) Save () error {
	return db.Model(Account{}).Omit("id", "deleted_at").Save(a).Find(a, "id = ?", a.ID).Error
}

// обновляет данные аккаунта кроме id, deleted_at и возвращает в Account обновленные данные
func (a *Account) Update (input interface{}) error {
	return db.Model(Account{}).Where("id = ?", a.ID).Omit("id", "deleted_at").Update(input).Find(a, "id = ?", a.ID).Error
}

// # Delete
func (a *Account) Delete () error {
	return db.Model(Account{}).Where("id = ?", a.ID).Delete(a).Error
}

// удаляет аккаунт с концами
func (a *Account) DeleteUnscoped () error {
	return db.Model(Account{}).Where("id = ?", a.ID).Unscoped().Delete(a).Error
}


// ### Account inner USER func
// todo: пересмотреть работу функций под AccountUser

// добавляет пользователя в аккаунт. Если пользователь уже в аккаунте, то ничего не произойдет.
func (a *Account) AppendUser (user *User) error {
	return db.Model(&user).Association("accounts").Append(a).Error
}

func (a *Account) RemoveUser (user *User) error {
	return db.Model(&user).Association("accounts").Delete(a).Error
}

// загружает список обычных пользователей аккаунта
func (a *Account) GetUsers () error {
	return db.Preload("Users").First(&a).Error
}

// ### Account inner func API KEYS

func (a *Account) CreateApiToken(key *ApiKey) error {
	// 1. Привязываем к аккаунту
	key.AccountID = a.ID

	// 2. Создаем
	if err := key.create(); err != nil {
		return err
	}

	return nil
}

func (a *Account) GetApiKeys() error {
	return db.Preload("ApiKeys").First(&a).Error
}

func (a *Account) DeleteApiKey(key *ApiKey) error {
	return key.delete()
}


// ### Stock functions
func (a Account) StockCreate(stock *Stock) error {
	stock.AccountID = a.ID
	return stock.Create()
}
func (a *Account) StockLoad() (err error) {
	a.Stocks, err = (Stock{}).GetAll(a.ID)
	return err
}

// ### Account inner func Products
func (a Account) ProductCreate(p *Product) error {
	p.AccountID = a.ID
	return p.Create()
}
func (a *Account) ProductLoad() (err error) {
	a.Products, err = (Product{}).GetAll(a.ID)
	return err
	//return db.Preload("Products").Preload("Products.Offers").First(&a).Error
}




// EAVAttributes
func (a Account) CreateEavAttribute(ea *EavAttribute) error {
	ea.AccountID = a.ID
	return ea.create()
}