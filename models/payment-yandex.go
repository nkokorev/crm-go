package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"github.com/satori/go.uuid"
	"net/http"
	"strings"
)

type PaymentYandex struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	HashId 		string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный Id для защиты от спама/парсинга
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	WebSiteId	uint 	`json:"webSiteId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится

	Code 		string	`json:"code" gorm:"type:varchar(16);default:'payment_yandexes';"` // Для идентификации

	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя интеграции магазина "<name>"
	Label 		string 	`json:"label" gorm:"type:varchar(128);default:''"` // 'Онлайн-оплата картой'

	// ####### API PaymentYandex ####### //

	// Для авторизации Basic Auth: username:password
	ApiKey	string	`json:"apiKey" gorm:"type:varchar(128);"` // ApiKey от яндекс кассы
	ShopId	string	`json:"shopId" gorm:"type:int;"` // shop id от яндекс кассы

	// URL для уведомлений со стороны Я.Кассы.
	// URL		string 	`json:"url" gorm:"type:varchar(255);"`
	EnabledIncomingNotifications	bool	`json:"enabledIncomingNotifications" gorm:"type:bool;default:true"` // обрабатывать ли уведомления от Я.Кассы.
	// ####### Внутренние данные ####### //

	// Возврат после платежа или отмена для пользователя
	ReturnUrl		string 	`json:"returnUrl" gorm:"type:varchar(255);"`

	// Включен ли данный способ оплаты
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // обрабатывать ли вебхук
	// Description 		string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание что к чему)

	// Сохранение платежных данных (с их помощью можно проводить повторные безакцептные списания ).
	SavePaymentMethod 	bool 	`json:"savePaymentMethod" gorm:"type:bool;default:false"`

	// Автоматический прием  поступившего платежа. Со стороны Я.Кассы
	Capture	bool	`json:"capture" gorm:"type:bool;default:true"`

	WebSite	WebSite `json:"webSite" gorm:"preload"`

	// !!! deprecated !!!
	// PaymentOption   PaymentOption `gorm:"polymorphic:Owner;"`
}

func (PaymentYandex) PgSqlCreate() {
	db.CreateTable(&PaymentYandex{})
	db.Model(&PaymentYandex{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&PaymentYandex{}).AddForeignKey("web_site_id", "web_sites(id)", "CASCADE", "CASCADE")
}
func (paymentYandex *PaymentYandex) BeforeCreate(scope *gorm.Scope) error {
	paymentYandex.Id = 0
	paymentYandex.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	return nil
}
// ############# Entity interface #############
func (paymentYandex PaymentYandex) GetId() uint { return paymentYandex.Id }
func (paymentYandex *PaymentYandex) setId(id uint) { paymentYandex.Id = id }
func (paymentYandex PaymentYandex) GetAccountId() uint { return paymentYandex.AccountId }
func (paymentYandex *PaymentYandex) setAccountId(id uint) { paymentYandex.AccountId = id }
func (PaymentYandex) SystemEntity() bool { return false }
// ############# END OF Entity interface #############

// ############# Payment Method interface #############
func (paymentYandex PaymentYandex) CreatePayment(order Order) (*Payment, error) {

	_p := Payment {
		AccountId: paymentYandex.AccountId,
		Paid: false,
		Amount: order.Amount,
		IncomeAmount: order.Amount,
		RefundedAmount: PaymentAmount{AccountId: order.AccountId, Value: float64(0), Currency: "RUB"},
		Description:  fmt.Sprintf("Заказ №%v в магазине AiroCliamte", order.Id),  // Видит клиент
		PaymentMethodData: PaymentMethodData{Type: "bank_card"}, // вообще еще вопрос
		Confirmation: Confirmation{Type: "redirect", ReturnUrl: paymentYandex.ReturnUrl},
		Receipt: Receipt{
			Customer: Customer{
				Email: order.Customer.Email,
				Phone: order.Customer.Phone,
				FullName: order.Customer.Name + " " + order.Customer.Surname,
			},
			Items: order.CartItems,
			Email: order.Customer.Email,
			Phone: order.Customer.Phone,
		},
		/*Recipient: Recipient{
			AccountId: paymentYandex.ShopId,
		},*/

		// Чтобы понять какой платеж был оплачен!!!
		Metadata: postgres.Jsonb{ RawMessage: utils.MapToRawJson(map[string]interface{}{
			"orderId":order.Id,
			"accountId":paymentYandex.AccountId,
		})},
		SavePaymentMethod: paymentYandex.SavePaymentMethod,
		OwnerId: paymentYandex.Id,
		OwnerType: "payment_yandexes",
		Capture: paymentYandex.Capture,

		OrderId: order.Id,
	}

	// создаем внутри платеж
	entity, err := _p.create()
	if err != nil {
		return nil, err
	}
	payment := entity.(*Payment)

	// Вызываем Yandex для созданного платежа
	err = paymentYandex.ExternalCreate(payment)
	if err != nil {
		return nil, err
	}

	return payment, nil
}
func (paymentYandex PaymentYandex) GetWebSiteId() uint { return paymentYandex.WebSiteId }
func (paymentYandex PaymentYandex) GetCode() string { return paymentYandex.Code }
// ############# END OF Payment Method interface #############

// ######### CRUD Functions ############
func (paymentYandex PaymentYandex) create() (Entity, error)  {
	py := paymentYandex
	if err := db.Create(&py).Error; err != nil {
		return nil, err
	}

	if err := py.GetPreloadDb(false,true).First(&py,py.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &py

	return entity, nil
}

func (PaymentYandex) get(id uint) (Entity, error) {

	var paymentYandex PaymentYandex

	err := paymentYandex.GetPreloadDb(false,false).First(&paymentYandex, id).Error
	if err != nil {
		return nil, err
	}
	return &paymentYandex, nil
}
func (PaymentYandex) getByHashId(hashId string) (*PaymentYandex, error) {
	paymentYandex := PaymentYandex{}

	err := paymentYandex.GetPreloadDb(false,false).First(&paymentYandex, "hash_id = ?", hashId).Error
	if err != nil {
		return nil, err
	}
	return &paymentYandex, nil
}
func (paymentYandex *PaymentYandex) load() error {
	if paymentYandex.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить PaymentYandex - не указан  Id"}
	}

	err := paymentYandex.GetPreloadDb(false,true).First(paymentYandex,paymentYandex.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (PaymentYandex) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return  PaymentYandex{}.getPaginationList(accountId, 0, 100, sortBy, "")
}

func (PaymentYandex) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	paymentYandexs := make([]PaymentYandex,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&PaymentYandex{}).GetPreloadDb(false,false).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&paymentYandexs, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentYandex{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&PaymentYandex{}).GetPreloadDb(false,false).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("WebSite").
			Find(&paymentYandexs).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&PaymentYandex{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(paymentYandexs))
	for i,_ := range paymentYandexs {
		entities[i] = &paymentYandexs[i]
	}

	return entities, total, nil
}

func (paymentYandex *PaymentYandex) update(input map[string]interface{}) error {
	return paymentYandex.GetPreloadDb(false,false).
		Model(paymentYandex).Where("id", paymentYandex.Id).Omit("id", "account_id").Updates(input).Error
}

func (paymentYandex *PaymentYandex) delete () error {
	return paymentYandex.GetPreloadDb(true,true).Where("id = ?", paymentYandex.Id).Delete(paymentYandex).Error
}
// ######### END CRUD Functions ############

// ########## Work function ############
func (paymentYandex PaymentYandex) ExternalCreate(payment *Payment) error {

	sendData := struct{
		Amount PaymentAmount `json:"amount"`
		PaymentMethodData PaymentMethodData `json:"payment_method_data"`
		SavePaymentMethod 	bool 	`json:"save_payment_method"`
		Capture	bool	`json:"capture" `
		Confirmation	Confirmation	`json:"confirmation"`
		Description 	string 	`json:"description"`
		Metadata	postgres.Jsonb	`json:"metadata"`
		Receipt Receipt `json:"receipt"`

	}{
		Amount: payment.Amount,
		PaymentMethodData: payment.PaymentMethodData,
		SavePaymentMethod: payment.SavePaymentMethod,
		Capture: payment.Capture,
		Confirmation: payment.Confirmation,
		Description: payment.Description,
		Metadata: payment.Metadata,
		Receipt: payment.Receipt,
	}


	url := "https://payment.yandex.net/api/v3/payments"

	// Собираем JSON данные
	body, err := json.Marshal(sendData)
	if err != nil {
		return utils.Error{Message: "Не удалось разобрать JSON платежа", Errors: map[string]interface{}{"paymentOption":err.Error()}}
	}

	uuidV4, err := uuid.NewV4()
	if err != nil {
		return utils.Error{Message: "Не удалось создать UUID для создания платежа", Errors: map[string]interface{}{"paymentOption":err.Error()}}
	}

	// crate new request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return utils.Error{Message: "Не удалось создать http-запрос для создания платежа",
			Errors: map[string]interface{}{"paymentOption":err.Error()}}
	}

	request.Header.Set("Idempotence-Key", uuidV4.String())
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(paymentYandex.ShopId, paymentYandex.ApiKey)

	// Делаем вызов
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Ошибка запроса для Yandex кассы: %v", err),
			Errors: map[string]interface{}{"paymentOption":err.Error()}}
	}
	defer response.Body.Close()


	if response.StatusCode == 200 {

		var responseRequest map[string]interface{}
		if err := json.NewDecoder(response.Body).Decode(&responseRequest); err != nil {
			return utils.Error{Message: fmt.Sprintf("Ошибка разбора ответа от Yandex кассы: %v", err),
					Errors: map[string]interface{}{"paymentOption":err.Error()}}
		}

		if err = payment.update(map[string]interface{}{
			"externalId":responseRequest["id"],
			"externalCreatedAt":responseRequest["created_at"],
			"paid":responseRequest["paid"],
			"status":responseRequest["status"],
			"test":responseRequest["test"],
		}); err != nil {
			return utils.Error{Message: "Ошибка сохранения responseRequest от Яндекс кассы",Errors: map[string]interface{}{"paymentOption":err.Error()}}
		}

		var confirmation struct {
			Type 	string `json:"type" gorm:"type:varchar(32);"` // embedded, redirect, external, qr
			ReturnUrl 	string `json:"return_url" gorm:"type:varchar(255);"`
			ConfirmationUrl 	string `json:"confirmation_url" gorm:"type:varchar(255);"`
		}

		jsonString, err := json.Marshal(responseRequest["confirmation"])
		if err != nil {
			return utils.Error{Message: fmt.Sprintf("Ошибка разбора ответа от Yandex кассы 1: %v", err),
					Errors: map[string]interface{}{"paymentOption":err.Error()}}
		}

		if err := json.NewDecoder(bytes.NewBuffer(jsonString)).Decode(&confirmation); err != nil {
			return utils.Error{Message: fmt.Sprintf("Ошибка разбора ответа от Yandex кассы: %v", err),
					Errors: map[string]interface{}{"paymentOption":err.Error()}}
		}

		payment.ConfirmationUrl = confirmation.ConfirmationUrl
		if err := payment.update(map[string]interface{}{"ConfirmationUrl":confirmation.ConfirmationUrl}); err != nil {
			return err
		}

	} else {
		var responseRequest map[string]interface{}
		if err := json.NewDecoder(response.Body).Decode(&responseRequest); err != nil {
			return utils.Error{Message: fmt.Sprintf("Ошибка разбора ответа от Yandex кассы: %v", err),
				Errors: map[string]interface{}{"paymentOption":err.Error()}}
		}
		fmt.Println(responseRequest)
		return utils.Error{Message: fmt.Sprintf("Ответ сервера Яндекс.Кассы: %v", response.StatusCode),Errors: map[string]interface{}{"paymentOption":"Проблема с сервером Яндекс.Кассы"}}
	}

	return nil
}


func (paymentYandex PaymentYandex) SetPaymentOption(paymentOption PaymentOption) error {
	if err := db.Model(&paymentYandex).Association("PaymentOption").Append(paymentOption).Error; err != nil {
		return err
	}

	return nil
}

func (account Account) GetPaymentYandexByHashId(hashId string) (*PaymentYandex, error) {
	paymentYandex, err := (PaymentYandex{}).getByHashId(hashId)
	if err != nil {
		return nil, err
	}

	if paymentYandex.AccountId != account.Id {
		return nil, errors.New("Объект принадлежит другому аккаунту")
	}

	return paymentYandex, nil
}

func (paymentYandex PaymentYandex) GetPreloadDb(autoUpdate bool, getModel bool) *gorm.DB {
	_db := db

	if autoUpdate { _db.Set("gorm:association_autoupdate", false) }
	if getModel { _db.Model(&paymentYandex) }

	return _db.Preload("WebSite")
}