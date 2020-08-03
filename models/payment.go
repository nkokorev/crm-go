package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type Payment struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	PublicId	uint   	`json:"publicId" gorm:"type:int;index;not null;default:1"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Идентификатор платежа в Яндекс.Кассе или у другого посредника.
	ExternalId	string	`json:"externalId" gorm:"type:varchar(128);index;default:null"`

	// статус платежа:  [pending, waiting_for_capture, succeeded и canceled]
	Status	string	`json:"status" gorm:"type:varchar(32);default:'pending'"`
	Paid 	bool 		`json:"paid" gorm:"type:bool;default:false;"` // признак оплаты платежа, для быстрой выборки

	Test 	bool 		`json:"test" gorm:"type:bool;default:false;"` // признак тестовой платежа
	
	// объем платежа по факту
	// AmountValue 	float64	`json:"amountValue" gorm:"type:numeric;default:0"`
	// AmountCurrency 	string 	`json:"amountCurrency" gorm:"type:varchar(3);default:'RUB'"` // сумма валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html
	AmountId  uint	`json:"amountId" gorm:"type:int;not null;"`
	Amount  PaymentAmount	`json:"amount"`

	// Каков "приход" за вычетом комиссии посредника.
	// IncomeValue 	float64 `json:"incomeValue" gorm:"type:numeric;default:0"`
	// IncomeCurrency 	string 	`json:"incomeCurrency" gorm:"type:varchar(3);default:'RUB'"` // сумма валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html
	IncomeAmountId  uint	`json:"incomeAmountId" gorm:"type:int;"`
	IncomeAmount  	PaymentAmount	`json:"income_amount"`

	// Сумма, которая вернулась пользователю. Присутствует, если у этого платежа есть успешные возвраты.
	Refundable 				bool 	`json:"refundable" gorm:"type:bool;default:false;"` // Возможность провести возврат по API
	RefundedAmountId  uint	`json:"refundedAmountId" gorm:"type:int;not null;"`
	RefundedAmount  PaymentAmount	`json:"refunded_amount"`

	// описание транзакции, которую в Я.Кассе пользователь увидит при оплате
	Description 	string 	`json:"description" gorm:"type:varchar(255);default:''"`

	// Получатель платежа на стороне Сервиса. В Яндекс кассе это магазин и канал внутри я.кассы.
	// Нужен, если вы разделяете потоки платежей в рамках одного аккаунта или создаете платеж в адрес другого аккаунта.
	Recipient	Recipient `json:"recipient"`

	Receipt	Receipt `json:"receipt"`

	// Способ оплаты платежа = {type:"bank_card", id:"", saved:true, card:""}. Может быть и другой платеж, в зависимости от OwnerType
	PaymentMethodData	PaymentMethodData	`json:"payment_method_data"`

	// Сохранение платежных данных (с их помощью можно проводить повторные безакцептные списания ).
	SavePaymentMethod 	bool 	`json:"savePaymentMethod" gorm:"type:bool;default:false"`

	// Автоматический прием  поступившего платежа. Со стороны Я.Кассы
	Capture	bool	`json:"capture" gorm:"type:bool;default:true"`

	// Способ подтверждения платежа. Присутствует, когда платеж ожидает подтверждения от пользователя
	Confirmation	Confirmation	`json:"confirmation" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`

	// Статус доставки данных для чека в онлайн-кассу (pending, succeeded или canceled).
	ReceiptRegistration string	`json:"receiptRegistration" gorm:"type:varchar(32);"`

	// Любые дополнительные данные. Передаются в виде набора пар «ключ-значение» и возвращаются в ответе от Яндекс.Кассы.
	// Ограничения: максимум 16 ключей, имя ключа не больше 32 символов, значение ключа не больше 512 символов.
	Metadata	postgres.Jsonb	`json:"metadata" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`

	// Комментарий к статусу canceled: {party:"[yandex_checkout, payment_network, merchant]", reason:"..."}
	// CancellationDetails	postgres.Jsonb	`json:"cancellationDetails" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	CancellationDetails CancellationDetails `json:"cancellationDetails"`

	// Данные об авторизации платежа. {rrn:"", auth_code:""}
	// AuthorizationDetails	postgres.Jsonb	`json:"authorizationDetails" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	AuthorizationDetails	AuthorizationDetails	`json:"authorizationDetails"`

	// Данные о распределении денег {account_id:"", amount:"", status:"[waiting_for_capture, succeeded, canceled]"}
	Transfers	Transfers	`json:"_transfers"`

	// URL на который переадресуется пользователь
	ConfirmationUrl	string	`json:"confirmation_url" gorm:"type:varchar(255);"`

	// #### Внутренние данные #####

	// Объекты для определения типа платежа внутри CRM: yandex, chase ..
	OwnerId	uint	`json:"ownerId" gorm:"type:int;not null;"` // Id в
	OwnerType	string `json:"ownerType" gorm:"type:varchar(255);not null;"`

	// ID заказа в RatusCRM
	OrderId	uint	`json:"orderId" gorm:"type:int"` // Id заказа в системе
	// Order Order		`json:"order"`

	ExternalCapturedAt 	time.Time  `json:"externalCapturedAt"` // Время подтверждения платежа, UTC
	ExternalExpiresAt 	time.Time  `json:"externalExpiresAt"`  // Время, до которого вы можете бесплатно отменить или подтвердить платеж.
	ExternalCreatedAt 	time.Time  `json:"externalCreatedAt"`  // Время создания заказа, UTC

	// PaymentOptions 	[]PaymentOption `json:"paymentOptions" gorm:"many2many:payment_options_payments;preload"`
	PaymentMethodId 	uint	`json:"paymentMethodId" gorm:"type:int;not null;default:1"`
	PaymentMethodType 	string	`json:"paymentMethodType" gorm:"type:varchar(32);not null;default:'payment_yandexes'"`
	PaymentMethod 		PaymentMethod `json:"paymentMethod" gorm:"-"`

	// Внутреннее время
	PaidAt time.Time  `json:"paidAt"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

func (Payment) PgSqlCreate() {
	db.CreateTable(&Payment{})
	db.Model(&Payment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&Payment{}).AddForeignKey("amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
	db.Model(&Payment{}).AddForeignKey("income_amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
	db.Model(&Payment{}).AddForeignKey("refunded_amount_id", "payment_amounts(id)", "CASCADE", "CASCADE")
}
func (payment *Payment) BeforeCreate(scope *gorm.Scope) error {
	payment.Id = 0

	// PublicId
	lastIdx := uint(0)
	var ord Order

	err := db.Where("account_id = ?", payment.AccountId).Select("public_id").Last(&ord).Error;
	if err != nil && err != gorm.ErrRecordNotFound { return err}
	if err == gorm.ErrRecordNotFound {
		lastIdx = 0
	} else {
		lastIdx = ord.PublicId
	}
	payment.PublicId = lastIdx + 1

	return nil
}

func (payment *Payment) AfterCreate(scope *gorm.Scope) (error) {
	event.AsyncFire(Event{}.PaymentCreated(payment.AccountId, payment.Id))
	return nil
}
func (payment *Payment) AfterUpdate(tx *gorm.DB) (err error) {

	event.AsyncFire(Event{}.PaymentUpdated(payment.AccountId, payment.Id))

	// статус платежа:  [pending, waiting_for_capture, succeeded и canceled]
	if payment.Paid || payment.Status == "succeeded" {
		event.AsyncFire(Event{}.PaymentCompleted(payment.AccountId, payment.Id))
	}
	if payment.Status == "canceled" {
		event.AsyncFire(Event{}.PaymentCanceled(payment.AccountId, payment.Id))
	}

	return nil
}
func (payment *Payment) AfterDelete(tx *gorm.DB) (err error) {
	event.AsyncFire(Event{}.PaymentDeleted(payment.AccountId, payment.Id))
	return nil
}
func (payment *Payment) AfterFind() (err error) {


	if payment.PaymentMethodType != "" && payment.PaymentMethodId > 0 {
		// Get ALL Payment Methods
		method, err := Account{Id: payment.AccountId}.GetPaymentMethodByType(payment.PaymentMethodType, payment.PaymentMethodId)
		if err != nil { return err }
		payment.PaymentMethod = method
	}


	return nil
}

type Confirmation struct {
	Type 	string `json:"type" gorm:"type:varchar(32);"` // embedded, redirect, external, qr
	ReturnUrl 	string `json:"return_url" gorm:"type:varchar(255);"`
}

type Customer struct {
	FullName	string	`json:"full_name"`
	Inn	string	`json:"inn"`
	Email	string	`json:"email"`
	Phone	string	`json:"phone"`
}

// Получатель платежа
type Recipient struct {
	AccountId string	`json:"account_id" gorm:"type:varchar(32);default:''"` // Идентификатор магазина в Яндекс.Кассе.
	GatewayId string	`json:"gateway_id" gorm:"type:varchar(32);default:''"` // Идентификатор субаккаунта - для разделения потоков платежей в рамках одного аккаунта.
}

type Receipt struct {
	Customer  Customer
	Items	[]CartItem `json:"items"`
	// TaxSystemCode int `json:"tax_system_code"`
	Email	string	`json:"email"`
	Phone	string	`json:"phone"`
	AccountId string	`json:"account_id" gorm:"type:varchar(32);default:''"` // Идентификатор магазина в Яндекс.Кассе.
}

type CancellationDetails struct {
	Party	string	`json:"party" gorm:"type:varchar(32);default:''"`
	Reason	string	`json:"reason" gorm:"type:varchar(32);default:''"`
}
type AuthorizationDetails struct {
	// Retrieval Reference Number — уникальный идентификатор транзакции в системе эмитента. Используется при оплате банковской картой.
	Rrn	string	`json:"rrn" gorm:"type:varchar(32);default:''"`

	// Код авторизации банковской карты. Выдается эмитентом и подтверждает проведение авторизации.
	AuthCode	string	`json:"authCode" gorm:"type:varchar(50);default:''"` // Идентификатор субаккаунта - для разделения потоков платежей в рамках одного аккаунта.
}
type Transfers struct {
	// Retrieval Reference Number — уникальный идентификатор транзакции в системе эмитента. Используется при оплате банковской картой.
	AccountId	string	`json:"account_id" gorm:"type:varchar(32);default:''"`

	// Код авторизации банковской карты. Выдается эмитентом и подтверждает проведение авторизации.

	// Идентификатор субаккаунта - для разделения потоков платежей в рамках одного аккаунта.
	Amount	PaymentAmount	`json:"amount"`
	Status	string	`json:"status" gorm:"type:varchar(50);default:''"` // Идентификатор субаккаунта - для разделения потоков платежей в рамках одного аккаунта.
}
type PaymentMethodData struct {
	Type string `json:"type"`
}

// ############# Entity interface #############
func (payment Payment) GetId() uint { return payment.Id }
func (payment *Payment) setId(id uint) { payment.Id = id }
func (payment *Payment) setPublicId(publicId uint) { payment.PublicId = publicId }
func (payment Payment) GetAccountId() uint { return payment.AccountId }
func (payment *Payment) setAccountId(id uint) { payment.AccountId = id }
func (Payment) SystemEntity() bool { return false }
// ############# Entity interface #############

// ######### CRUD Functions ############
func (payment Payment) create() (Entity, error)  {
	// if err := db.Create(&payment).Find(&payment, payment.Id).Error; err != nil {
	wb := payment
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (Payment) get(id uint) (Entity, error) {

	var payment Payment

	err := payment.GetPreloadDb(false,false, true).First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
func (Payment) getByExternalId(externalId string) (*Payment, error) {
	payment := Payment{}

	err := db.First(&payment, "external_id = ?", externalId).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
func (payment *Payment) load() error {
	if payment.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Payment - не указан  Id"}
	}
	
	err := payment.GetPreloadDb(false,false, true).First(payment,payment.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (payment *Payment) loadByPublicId() error {


	if payment.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Payment - не указан  Id"}
	}

	if err := payment.GetPreloadDb(false,false, true).First(payment, "public_id = ?", payment.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (Payment) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	webHooks := make([]Payment,0)
	var total uint

	err := db.Model(&Payment{}).Limit(100).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&webHooks).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&Payment{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i,_ := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}

func (Payment) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	webHooks := make([]Payment,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&Payment{}).GetPreloadDb(true,false,true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Payment{}).GetPreloadDb(true,false,true).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&Payment{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Payment{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webHooks))
	for i,_ := range webHooks {
		entities[i] = &webHooks[i]
	}

	return entities, total, nil
}

func (Payment) getByEvent(eventName string) (*Payment, error) {

	wh := Payment{}

	if err := db.First(&wh, "event_type = ?", eventName).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}

func (payment *Payment) update(input map[string]interface{}) error {
	return db.Model(payment).Omit("id", "account_id").Updates(input).Error
	// return db.Model(Payment{}).Where("id = ?", payment.Id).Omit("id", "account_id").Updates(input).Error
}

func (payment *Payment) delete () error {

	var idx = make([]uint,0)
	idx = append(idx,payment.AmountId)
	idx = append(idx,payment.IncomeAmountId)
	idx = append(idx,payment.RefundedAmountId)

	if err := (PaymentAmount{}).deletes(idx); err != nil {
	  return err
	}

	return payment.GetPreloadDb(true,false,false).Where("id = ?", payment.Id).Delete(payment).Error
}
// ######### END CRUD Functions ############

func (account Account) GetPaymentByExternalId(externalId string) (*Payment, error) {
	payment, err := (Payment{}).getByExternalId(externalId)
	if err != nil {
		return nil, err
	}

	if payment.AccountId != account.Id {
		return nil, errors.New("Объект принадлежит другому аккаунту")
	}

	return payment, nil
}

func (payment *Payment) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {
	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&payment)
	} else {
		_db = _db.Model(&Payment{})
	}

	if preload {
		return _db.Preload("PaymentAmount")
	} else {
		return _db
	}
}

