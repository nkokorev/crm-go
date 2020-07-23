package models

import (
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"time"
)

type Payment struct {
	
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Идентификатор платежа в Яндекс.Кассе или у другого посредника.
	ExternalId	uint	`json:"externalId" gorm:"type:int;index;default:null"`

	// статус платежа:  [pending, waiting_for_capture, succeeded и canceled]
	Status	string	`json:"status" gorm:"type:varchar(32);not null"`
	Paid 	bool 		`json:"paid" gorm:"type:bool;default:false;"` // признак оплаты платежа, для быстрой выборки

	Test 	bool 		`json:"test" gorm:"type:bool;default:false;"` // признак тестовой платежа
	
	// объем платежа по факту
	// AmountValue 	float64	`json:"amountValue" gorm:"type:numeric;default:0"`
	// AmountCurrency 	string 	`json:"amountCurrency" gorm:"type:varchar(3);default:'RUB'"` // сумма валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html
	Amount  Amount	`json:"amount" gorm:"type:JSONB;"`

	// Каков "приход" за вычетом комиссии посредника.
	// IncomeValue 	float64 `json:"incomeValue" gorm:"type:numeric;default:0"`
	// IncomeCurrency 	string 	`json:"incomeCurrency" gorm:"type:varchar(3);default:'RUB'"` // сумма валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html
	IncomeAmount  Amount	`json:"income_amount"`

	// Сумма, которая вернулась пользователю. Присутствует, если у этого платежа есть успешные возвраты.
	Refundable 				bool 	`json:"refundable" gorm:"type:bool;default:false;"` // Возможность провести возврат по API
	// RefundedAmountValue 	float64	`json:"refundedAmountValue" gorm:"type:numeric;default:0"`
	// RefundedAmountCurrency 	string 	`json:"refundedAmountCurrency" gorm:"type:varchar(3);default:'RUB'"` // сумма валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html
	RefundedAmount  Amount	`json:"refunded_amount"`

	// описание транзакции, которую в Я.Кассе пользователь увидит при оплате
	Description 	string 	`json:"description" gorm:"type:varchar(255);default:''"`

	// Получатель платежа на стороне Сервиса. В Яндекс кассе это магазин и канал внутри я.кассы.
	// Нужен, если вы разделяете потоки платежей в рамках одного аккаунта или создаете платеж в адрес другого аккаунта.
	// RecipientAccountId	string	`json:"recipientAccountId" gorm:"type:varchar(255);default:''"` // Идентификатор магазина в Яндекс.Кассе.
	// RecipientGatewayId	string	`json:"recipientGatewayId" gorm:"type:varchar(255);default:''"` // Идентификатор субаккаунта - для разделения потоков платежей в рамках одного аккаунта.
	Recipient	Recipient `json:"_recipient"`

	// Способ оплаты платежа = {type:"bank_card", id:"", saved:true, card:""}. Может быть и другой платеж, в зависимости от OwnerType
	// PaymentMethod	postgres.Jsonb	`json:"paymentMethod" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	PaymentMethod	PaymentMethod	`json:"payment_method_data"`

	// Сохранение платежных данных (с их помощью можно проводить повторные безакцептные списания ).
	SavePaymentMethod 	bool 	`json:"savePaymentMethod" gorm:"type:bool;default:false"`

	// Автоматический прием  поступившего платежа. Со стороны Я.Кассы
	Capture	bool	`json:"capture" gorm:"type:bool;default:true"`


	// Способ подтверждения платежа. Присутствует, когда платеж ожидает подтверждения от пользователя
	// Object = {type:"", confirmation_token:""}
	// Confirmation	postgres.Jsonb	`json:"confirmation" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
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

	ConfirmationUrl	string	`json:"confirmation_url" gorm:"type:varchar(255);"`

	// #### Внутренние данные #####


	// Внутренний Id объекта типа платежа Яндекс.Касса, кэш или у другого посредника.
	OwnerId	uint	`json:"ownerId" gorm:"type:int"` // Id в
	// таблица или тип объекта: [yandex_payment,cash, ...] // тут надо бы доработать список
	OwnerType	string `json:"ownerType" gorm:"type:varchar(255);default:''"`

	// ID заказа в RatusCRM
	OrderId	uint	`json:"orderId" gorm:"type:int"` // Id заказа в системе

	ExternalCapturedAt 	time.Time  `json:"externalCapturedAt"` // Время подтверждения платежа, UTC
	ExternalExpiresAt 	time.Time  `json:"externalExpiresAt"`  // Время, до которого вы можете бесплатно отменить или подтвердить платеж.
	ExternalCreatedAt 	time.Time  `json:"externalCreatedAt"`  // Время создания заказа, UTC

	// Внутреннее время
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"-" sql:"index"`
}

//Объект платежа - кто-то, что-то вам заплатил. Или хочет заплатить. Или должен...
type Amount struct {
	Value 	float64	`json:"value" gorm:"type:numeric;default:0"`
	Currency 	string 	`json:"currency" gorm:"type:varchar(3);default:'RUB'"` // сумма валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html
}

type PaymentMethod struct {
	Type 	string `json:"type" gorm:"type:varchar(32);"`
}

type Confirmation struct {
	Type 	string `json:"type" gorm:"type:varchar(32);"` // embedded, redirect, external, qr
	ReturnUrl 	string `json:"return_url" gorm:"type:varchar(255);"`
}

type Recipient struct {
	AccountId	string	`json:"account_id" gorm:"type:varchar(32);default:''"` // Идентификатор магазина в Яндекс.Кассе.
	GatewayId	string	`json:"gateway_id" gorm:"type:varchar(32);default:''"` // Идентификатор субаккаунта - для разделения потоков платежей в рамках одного аккаунта.
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
	Amount	Amount	`json:"amount" ` // Идентификатор субаккаунта - для разделения потоков платежей в рамках одного аккаунта.
	Status	string	`json:"status" gorm:"type:varchar(50);default:''"` // Идентификатор субаккаунта - для разделения потоков платежей в рамках одного аккаунта.
}

// ############# Entity interface #############
func (payment Payment) GetId() uint { return payment.Id }
func (payment *Payment) setId(id uint) { payment.Id = id }
func (payment Payment) GetAccountId() uint { return payment.AccountId }
func (payment *Payment) setAccountId(id uint) { payment.AccountId = id }
func (Payment) SystemEntity() bool { return false }
// ############# Entity interface #############

func (Payment) PgSqlCreate() {
	db.CreateTable(&Payment{})
	db.Model(&Payment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (payment *Payment) BeforeCreate(scope *gorm.Scope) error {
	payment.Id = 0
	return nil
}

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

	err := db.First(&payment, id).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
func (payment *Payment) load() error {
	if payment.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить Payment - не указан  Id"}
	}

	err := db.First(payment,payment.Id).Error
	if err != nil {
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

		err := db.Model(&Payment{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&Payment{}).
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
	return db.Model(payment).Where("id", payment.Id).Omit("id", "account_id").Updates(input).Error
}

func (payment Payment) delete () error {
	return db.Model(Payment{}).Where("id = ?", payment.Id).Delete(payment).Error
}
// ######### END CRUD Functions ############
