package models

import (
	"database/sql"
	"errors"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"time"
)

type Payment struct {
	
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"type:int;index;not null;"` // Публичный ID заказа внутри магазина
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	// Идентификатор платежа в Яндекс.Кассе или у другого посредника.
	ExternalId	string	`json:"external_id" gorm:"type:varchar(128);index;"`

	// статус платежа:  [pending, waiting_for_capture, succeeded и canceled]
	Status	string		`json:"status" gorm:"type:varchar(32);default:'pending'"`
	Paid 	bool 		`json:"paid" gorm:"type:bool;default:false;"` // признак оплаты платежа, для быстрой выборки

	Test 	bool 		`json:"test" gorm:"type:bool;default:false;"` // признак тестовой платежа
	
	// объем платежа по факту
	// AmountValue 	float64	`json:"amountValue" gorm:"type:numeric;default:0"`
	// AmountCurrency 	string 	`json:"amountCurrency" gorm:"type:varchar(3);default:'RUB'"` // сумма валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html
	AmountId  	uint			`json:"amount_id" gorm:"type:int;not null;"`
	Amount  	PaymentAmount	`json:"amount"`

	// Каков "приход" за вычетом комиссии посредника.
	// IncomeValue 	float64 `json:"incomeValue" gorm:"type:numeric;default:0"`
	// IncomeCurrency 	string 	`json:"incomeCurrency" gorm:"type:varchar(3);default:'RUB'"` // сумма валюты в  ISO-4217 https://www.iso.org/iso-4217-currency-codes.html
	IncomeAmountId  uint	`json:"income_amount_id" gorm:"type:int;"`
	IncomeAmount  	PaymentAmount	`json:"income_amount" gorm:"-"`

	// Сумма, которая вернулась пользователю. Присутствует, если у этого платежа есть успешные возвраты.
	Refundable 			bool 	`json:"refundable" gorm:"type:bool;default:false;"` // Возможность провести возврат по API
	RefundedAmountId  	uint	`json:"refunded_amount_id" gorm:"type:int;not null;"`
	RefundedAmount  	PaymentAmount	`json:"refunded_amount"`

	// описание транзакции, которую в Я.Кассе пользователь увидит при оплате
	Description 	string 	`json:"description" gorm:"type:varchar(255);default:''"`

	// Получатель платежа на стороне Сервиса. В Яндекс кассе это магазин и канал внутри я.кассы.
	// Нужен, если вы разделяете потоки платежей в рамках одного аккаунта или создаете платеж в адрес другого аккаунта.
	Recipient	Recipient `json:"recipient" gorm:"-"`

	Receipt		Receipt `json:"receipt" gorm:"-"`

	// Способ оплаты платежа = {type:"bank_card", id:"", saved:true, card:""}. Может быть и другой платеж, в зависимости от OwnerType
	PaymentMethodData	PaymentMethodData	`json:"payment_method_data" gorm:"-"`

	// Сохранение платежных данных (с их помощью можно проводить повторные безакцептные списания ).
	SavePaymentMethod 	bool 	`json:"save_payment_method" gorm:"type:bool;default:false"`

	// Автоматический прием  поступившего платежа. Со стороны Я.Кассы
	Capture				bool	`json:"capture" gorm:"type:bool;default:true"`

	// Способ подтверждения платежа. Присутствует, когда платеж ожидает подтверждения от пользователя
	Confirmation		Confirmation	`json:"confirmation" gorm:"-"`

	// Статус доставки данных для чека в онлайн-кассу (pending, succeeded или canceled).
	ReceiptRegistration string	`json:"receipt_registration" gorm:"type:varchar(32);"`

	// Любые дополнительные данные. Передаются в виде набора пар «ключ-значение» и возвращаются в ответе от Яндекс.Кассы.
	// Ограничения: максимум 16 ключей, имя ключа не больше 32 символов, значение ключа не больше 512 символов.
	Metadata	datatypes.JSON	`json:"metadata" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`

	// Комментарий к статусу canceled: {party:"[yandex_checkout, payment_network, merchant]", reason:"..."}
	// CancellationDetails	postgres.Jsonb	`json:"cancellation_details" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	CancellationDetails 	CancellationDetails `json:"cancellation_details" gorm:"-"`

	// Данные об авторизации платежа. {rrn:"", auth_code:""}
	// AuthorizationDetails	postgres.Jsonb	`json:"authorization_details" gorm:"type:JSONB;DEFAULT '{}'::JSONB"`
	AuthorizationDetails	AuthorizationDetails	`json:"authorization_details" gorm:"-"`

	// Данные о распределении денег {account_id:"", amount:"", status:"[waiting_for_capture, succeeded, canceled]"}
	Transfers	Transfers	`json:"_transfers" gorm:"-"`

	// URL на который переадресуется пользователь
	ConfirmationUrl	string	`json:"confirmation_url" gorm:"type:varchar(255);"`

	// #### Внутренние данные #####

	// Объекты для определения типа платежа внутри CRM: yandex, chase ..
	OwnerId			uint	`json:"owner_id" gorm:"type:int;not null;"` // Id в
	OwnerType		string 	`json:"owner_type" gorm:"type:varchar(255);not null;"`

	// ID заказа в RatusCRM
	OrderId			uint	`json:"order_id" gorm:"type:int"` // Id заказа в системе
	// Order Order		`json:"order"`

	ExternalCapturedAt 	time.Time  `json:"external_captured_at"` // Время подтверждения платежа, UTC
	ExternalExpiresAt 	time.Time  `json:"external_expires_at"`  // Время, до которого вы можете бесплатно отменить или подтвердить платеж.
	ExternalCreatedAt 	time.Time  `json:"external_created_at"`  // Время создания заказа, UTC

	// PaymentOptions 	[]PaymentOption `json:"paymentOptions" gorm:"many2many:payment_options_payments;preload"`
	PaymentMethodId 	uint	`json:"payment_method_id" gorm:"type:int;not null;"`
	PaymentMethodType 	string	`json:"payment_method_type" gorm:"type:varchar(32);not null;default:'payment_yandexes'"`
	PaymentMethod 		PaymentMethod `json:"payment_method" gorm:"-"`

	// Внутреннее время
	PaidAt 		*time.Time  `json:"paid_at"`
	CreatedAt 	time.Time  `json:"created_at"`
	UpdatedAt 	time.Time  `json:"updated_at"`
	DeletedAt 	*time.Time `json:"-" sql:"index"`
}

func (Payment) PgSqlCreate() {
	if err := db.Migrator().CreateTable(&Payment{}); err != nil {log.Fatal(err)}
	// db.Model(&Payment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&Payment{}).AddForeignKey("amount_id", "payment_amounts(id)", "RESTRICT", "CASCADE")
	// db.Model(&Payment{}).AddForeignKey("income_amount_id", "payment_amounts(id)", "RESTRICT", "CASCADE")
	// db.Model(&Payment{}).AddForeignKey("refunded_amount_id", "payment_amounts(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE payments " +
		"ADD CONSTRAINT payments_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"DROP CONSTRAINT IF EXISTS fk_orders_payment," +
		"DROP CONSTRAINT IF EXISTS fk_payments_amount," +
		"DROP CONSTRAINT IF EXISTS fk_payments_refunded_amount;").Error
		// "ADD CONSTRAINT payments_amount_id_fkey FOREIGN KEY (amount_id) REFERENCES payment_amounts(id) ON DELETE RESTRICT ON UPDATE CASCADE," +
		// "ADD CONSTRAINT payments_income_amount_id_fkey FOREIGN KEY (income_amount_id) REFERENCES payment_amounts(id) ON DELETE RESTRICT ON UPDATE CASCADE," +
		// "ADD CONSTRAINT payments_refunded_amount_id_fkey FOREIGN KEY (refunded_amount_id) REFERENCES payment_amounts(id) ON DELETE RESTRICT ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
	err = db.Exec("ALTER TABLE payments " +
		"DROP CONSTRAINT IF EXISTS fk_orders_amount," +
		"DROP CONSTRAINT IF EXISTS fk_payments_amount," +
		"DROP CONSTRAINT IF EXISTS fk_payments_refunded_amount;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (payment *Payment) BeforeCreate(tx *gorm.DB) error {
	payment.Id = 0

	// PublicId
	var lastIdx sql.NullInt64
	err := db.Model(&Payment{}).Where("account_id = ?",  payment.AccountId).
		Select("max(public_id)").Row().Scan(&lastIdx)
	if err != nil && err != gorm.ErrRecordNotFound { return err }
	payment.PublicId = 1 + uint(lastIdx.Int64)

	return nil
}
func (payment *Payment) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(payment)
	} else {
		_db = _db.Model(&Payment{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"Amount","IncomeAmount","RefundedAmount","PaymentMethod"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}

func (payment *Payment) AfterCreate(tx *gorm.DB) error {
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
func (payment *Payment) AfterFind(tx *gorm.DB) (err error) {


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
// псевдо..
type Customer struct {
	FullName	string	`json:"full_name"`
	Inn	string	`json:"-"`
	Email	string	`json:"email"`
	Phone	string	`json:"phone"`
}

// Получатель платежа
type Recipient struct {
	AccountId string	`json:"account_id" gorm:"type:varchar(32);default:''"` // Идентификатор магазина в Яндекс.Кассе.
	GatewayId string	`json:"gateway_id" gorm:"type:varchar(32);default:''"` // Идентификатор субаккаунта - для разделения потоков платежей в рамках одного аккаунта.
}

type Receipt struct {
	Customer  Customer `json:"customer"`
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
	_item := payment
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (Payment) get(id uint, preloads []string) (Entity, error) {
	var item Payment

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}
func (Payment) getByExternalId(externalId string) (*Payment, error) {
	payment := Payment{}

	err := payment.GetPreloadDb(false,true, nil).First(&payment, "external_id = ?", externalId).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}
func (payment *Payment) load(preloads []string) error {
	if payment.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить CartItem - не указан  Id"}
	}

	err := payment.GetPreloadDb(false, false, preloads).First(payment, payment.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (payment *Payment) loadByPublicId(preloads []string) error {

	if payment.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Payment - не указан  Id"}
	}

	if err := payment.GetPreloadDb(false,false, preloads).
		First(payment, "account_id = ? AND public_id = ?", payment.AccountId, payment.PublicId).Error; err != nil {
		return err
	}

	return nil
}

func (Payment) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return Payment{}.getPaginationList(accountId, 0, 25, sortBy, "", nil,preload)
}
func (Payment) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	payments := make([]Payment,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&Payment{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&payments, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Payment{}).GetPreloadDb(false,false,nil).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&Payment{}).GetPreloadDb(false,false,preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&payments).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&Payment{}).GetPreloadDb(false,false,nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(payments))
	for i := range payments {
		entities[i] = &payments[i]
	}

	return entities, total, nil
}
func (Payment) getByEvent(eventName string) (*Payment, error) {

	wh := Payment{}

	if err := (&Payment{}).GetPreloadDb(false, true, nil).First(&wh, "event_type = ?", eventName).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}

func (payment *Payment) update(input map[string]interface{}, preloads []string) error {
	delete(input,"amount")
	delete(input,"income_amount")
	delete(input,"refunded_amount")
	delete(input,"recipient")
	delete(input,"receipt")
	delete(input,"payment_method_data")
	delete(input,"confirmation")
	delete(input,"cancellation_details")
	delete(input,"authorization_details")
	delete(input,"payment_method")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id","external_id","amount_id","income_amount_id","refunded_amount_id","owner_id","order_id","payment_method_id"}); err != nil {
		return err
	}
	if _, ok := input["externalCreatedAt"]; ok {
		input["external_created_at"] = input["externalCreatedAt"]
		delete(input,"externalCreatedAt")
	}
	if _, ok := input["externalId"]; ok {
		input["external_id"] = input["externalId"]
		delete(input,"externalId")
	}
	if _, ok := input["paidAt"]; ok {
		input["paid_at"] = input["paidAt"]
		delete(input,"paidAt")
	}


	if err := payment.GetPreloadDb(false, false, nil).Where("id = ?", payment.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := payment.GetPreloadDb(false,false, preloads).First(payment, payment.Id).Error
	if err != nil {
		return err
	}

	return nil
}

func (payment *Payment) delete () error {

	var idx = make([]uint,0)
	idx = append(idx,payment.AmountId)
	idx = append(idx,payment.IncomeAmountId)
	idx = append(idx,payment.RefundedAmountId)

	if err := (PaymentAmount{}).deletes(idx); err != nil {
	  return err
	}

	return payment.GetPreloadDb(true,false,nil).Where("id = ?", payment.Id).Delete(payment).Error
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

