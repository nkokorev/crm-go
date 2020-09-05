package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"github.com/satori/go.uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type PaymentYandex struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	HashId 		string  `json:"hash_id" gorm:"type:varchar(12);unique_index;not null;"` // публичный Id для защиты от спама/парсинга
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`
	WebSiteId	uint 	`json:"web_site_id" gorm:"type:int;index;"` // магазин, к которому относится

	Code 		string `json:"code" gorm:"type:varchar(32);default:'payment_yandex'"`
	Type 		string `json:"type" gorm:"type:varchar(32);default:'payment_yandexes';"` // Для идентификации

	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя интеграции магазина "<name>"
	Label 		string 	`json:"label" gorm:"type:varchar(128);default:''"` // 'Онлайн-оплата картой'

	// ####### API PaymentYandex ####### //

	// Для авторизации Basic Auth: username:password
	ApiKey	string	`json:"api_key" gorm:"type:varchar(128);"` // ApiKey от яндекс кассы
	ShopId	int	`json:"shop_id" gorm:"type:int;"` // shop id от яндекс кассы

	// URL для уведомлений со стороны Я.Кассы.
	// URL		string 	`json:"url" gorm:"type:varchar(255);"`
	EnabledIncomingNotifications	bool	`json:"enabled_incoming_notifications" gorm:"type:bool;default:true"` // обрабатывать ли уведомления от Я.Кассы.

	// ####### Внутренние данные ####### //
	// Возврат после платежа или отмена для пользователя
	ReturnUrl		string 	`json:"return_url" gorm:"type:varchar(255);"`

	// Включен ли данный способ оплаты
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // обрабатывать ли вебхук
	// Description 		string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание что к чему)
	InstantDelivery 	bool 	`json:"instant_delivery" gorm:"type:bool;default:false"`

	// Сохранение платежных данных (с их помощью можно проводить повторные безакцептные списания ).
	SavePaymentMethod 	bool 	`json:"save_payment_method" gorm:"type:bool;default:false"`

	// Автоматический прием  поступившего платежа. Со стороны Я.Кассы
	Capture	bool	`json:"capture" gorm:"type:bool;default:true"`

	WebSite	WebSite `json:"web_site" gorm:"preload"`

	// !!! deprecated !!!
	// PaymentOption   PaymentOption `gorm:"polymorphic:Owner;"`

}

func (PaymentYandex) PgSqlCreate() {
	db.Migrator().CreateTable(&PaymentYandex{})
	// db.Model(&PaymentYandex{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&PaymentYandex{}).AddForeignKey("web_site_id", "web_sites(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE payment_yandexes " +
		"ADD CONSTRAINT payment_yandexes_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT payment_yandexes_web_site_id_fkey FOREIGN KEY (web_site_id) REFERENCES web_sites(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
func (paymentYandex *PaymentYandex) BeforeCreate(tx *gorm.DB) error {
	paymentYandex.Id = 0
	paymentYandex.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	return nil
}

// ############# Entity interface #############
func (paymentYandex PaymentYandex) GetId() uint { return paymentYandex.Id }
func (paymentYandex *PaymentYandex) setId(id uint) { paymentYandex.Id = id }
func (paymentYandex *PaymentYandex) setPublicId(id uint) { paymentYandex.Id = id }
func (paymentYandex PaymentYandex) GetAccountId() uint { return paymentYandex.AccountId }
func (paymentYandex *PaymentYandex) setAccountId(id uint) { paymentYandex.AccountId = id }
func (PaymentYandex) SystemEntity() bool { return false }
// ############# END OF Entity interface #############

// ############# Payment Method interface #############
func (paymentYandex PaymentYandex) CreatePaymentByOrder(order Order, mode PaymentMode) (*Payment, error) {

	// 1. Формируем paymentMode (Признак способа расчета): full_prepayment / full_payment / service
	for i := range(order.CartItems) {
		order.CartItems[i].Id = mode.Id
		order.CartItems[i].PaymentMode = mode
		order.CartItems[i].PaymentModeYandex = mode.Code
	}

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
				Email: *order.Customer.Email,
				Phone: *order.Customer.Phone,
				FullName: *order.Customer.Name + " " + *order.Customer.Surname,
			},
			Items: order.CartItems, // <<< Признак предмета расчета(?) & Признак способа расчета (?)
			Email: *order.Customer.Email,
			Phone: *order.Customer.Phone,
		},
		/*Recipient: Recipient{
			AccountId: paymentYandex.ShopId,
		},*/

		// Чтобы понять какой платеж был оплачен!!!
		Metadata: datatypes.JSON(utils.MapToRawJson(map[string]interface{}{
			"orderId":order.Id,
			"accountId":paymentYandex.AccountId,
		})),
		SavePaymentMethod: paymentYandex.SavePaymentMethod,
		OwnerId: paymentYandex.Id,
		OwnerType: "payment_yandexes",
		Capture: paymentYandex.Capture,

		OrderId: order.Id,
		PaymentMethodId: order.PaymentMethodId,
		PaymentMethodType: order.PaymentMethodType,
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
func (paymentYandex PaymentYandex) GetType() string { return "payment_yandexes" }
func (paymentYandex PaymentYandex) GetCode() string { return "payment_yandex" }
func (paymentYandex PaymentYandex) IsInstantDelivery() bool { return paymentYandex.InstantDelivery }
// ############# END OF Payment Method interface #############

// ######### CRUD Functions ############
func (paymentYandex PaymentYandex) create() (Entity, error)  {
	py := paymentYandex
	if err := db.Create(&py).Error; err != nil {
		return nil, err
	}

	if err := py.GetPreloadDb(false,false, true).First(&py,py.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &py

	return entity, nil
}
func (PaymentYandex) get(id uint) (Entity, error) {

	var paymentYandex PaymentYandex

	err := paymentYandex.GetPreloadDb(false,false, true).First(&paymentYandex, id).Error
	if err != nil {
		return nil, err
	}
	return &paymentYandex, nil
}
func (PaymentYandex) getByHashId(hashId string) (*PaymentYandex, error) {
	paymentYandex := PaymentYandex{}

	err := paymentYandex.GetPreloadDb(false,false, true).First(&paymentYandex, "hash_id = ?", hashId).Error
	if err != nil {
		return nil, err
	}
	return &paymentYandex, nil
}
func (paymentYandex *PaymentYandex) load() error {
	if paymentYandex.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить PaymentYandex - не указан  Id"}
	}

	err := paymentYandex.GetPreloadDb(false,true, true).First(paymentYandex,paymentYandex.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (*PaymentYandex) loadByPublicId() error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}
func (PaymentYandex) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return  PaymentYandex{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (PaymentYandex) GetListByWebSiteAndDelivery(delivery Delivery) ([]PaymentYandex, error) {

	methods := make([]PaymentYandex,0)

	err := db.Table("payment_to_delivery").
		Joins("LEFT JOIN payment_yandexes ON payment_yandexes.id = payment_to_delivery.payment_id AND payment_yandexes.type = payment_to_delivery.payment_type").
		Select("payment_to_delivery.account_id, payment_to_delivery.web_site_id, payment_to_delivery.payment_id, payment_to_delivery.payment_type, payment_to_delivery.delivery_id, payment_to_delivery.delivery_type, payment_yandexes.*").
		Where("payment_to_delivery.account_id = ? AND payment_to_delivery.web_site_id = ? " +
			"AND payment_to_delivery.delivery_id = ? AND payment_to_delivery.delivery_type = ? " +
			"AND payment_to_delivery.payment_type = ?",
			delivery.GetAccountId(), delivery.GetWebSiteId(), delivery.GetId(), delivery.GetType(), PaymentYandex{}.GetType()).Find(&methods).Error

	if err != nil { return nil, err }

	return methods,nil
}
func (PaymentYandex) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	paymentYandexs := make([]PaymentYandex,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&PaymentYandex{}).GetPreloadDb(false,false, true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

		err := (&PaymentYandex{}).GetPreloadDb(false,false, true).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
	for i := range paymentYandexs {
		entities[i] = &paymentYandexs[i]
	}

	return entities, total, nil
}
func (paymentYandex *PaymentYandex) update(input map[string]interface{}) error {
	delete(input,"web_site")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"web_site_id","shop_id"}); err != nil {
		return err
	}

	err := paymentYandex.GetPreloadDb(false,false, false).Where("id = ?", paymentYandex.Id).
		Omit("id", "hashId","account_id").Updates(input).Error
	if err != nil { return err}
	_ = paymentYandex.load()
	return nil
}
func (paymentYandex *PaymentYandex) delete () error {
	return paymentYandex.GetPreloadDb(true,true, false).Where("id = ?", paymentYandex.Id).Delete(paymentYandex).Error
}
// ######### END CRUD Functions ############

// ########## Work function ############
func (paymentYandex PaymentYandex) ExternalCreate(payment *Payment) error {

	sendData := struct {
		Amount PaymentAmount `json:"amount"`
		PaymentMethodData PaymentMethodData `json:"payment_method_data"`
		SavePaymentMethod 	bool 	`json:"save_payment_method"`
		Capture	bool	`json:"capture" `
		Confirmation	Confirmation	`json:"confirmation"`
		Description 	string 	`json:"description"`
		Metadata	datatypes.JSON	`json:"metadata"`
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

	// fmt.Println("PaymentSubjectYandex: ", sendData.Receipt.Items[0].PaymentSubjectYandex)
	// return utils.Error{Message: "Не удалось разобрать JSON платежа"}
	
	url := "https://payment.yandex.net/api/v3/payments"

	// Собираем JSON данные
	body, err := json.Marshal(sendData)
	if err != nil {
		return utils.Error{Message: "Не удалось разобрать JSON платежа", Errors: map[string]interface{}{"paymentOption":err.Error()}}
	}

	uuidV4 := uuid.NewV4()

	// crate new request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return utils.Error{Message: "Не удалось создать http-запрос для создания платежа",
			Errors: map[string]interface{}{"paymentOption":err.Error()}}
	}

	request.Header.Set("Idempotence-Key", uuidV4.String())
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(strconv.Itoa(paymentYandex.ShopId), paymentYandex.ApiKey)

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
				Errors: map[string]interface{}{"paymentMode":err.Error()}}
		}
		fmt.Println(responseRequest)
		return utils.Error{Message: fmt.Sprintf("Ответ сервера Яндекс.Кассы: %v", response.StatusCode),
			Errors: map[string]interface{}{"paymentMode":"Проблема с сервером Яндекс.Кассы"}}
	}

	return nil
}

// Создает чек зачета предоплаты
func (paymentYandex PaymentYandex) PrepaymentCheck(payment *Payment, order Order) (*Payment, error) {

	// 1. Формируем paymentMode (Признак способа расчета): full_prepayment / full_payment / service

	// Получаем признак полный оплаты
	mode, err := PaymentMode{}.GetFullPaymentMode()
	if err != nil {
		return nil, utils.Error{Message: "Не удалось обновить статус платежа - не найден признак зачета предоплаты"}
	}


	for i := range(order.CartItems) {
		order.CartItems[i].Id = mode.Id
		order.CartItems[i].PaymentMode = mode
		order.CartItems[i].PaymentModeYandex = mode.Code
	}

	// fmt.Println("payment.Amount: ", payment.Amount)
	// fmt.Println("order.CartItems: ", order.CartItems[1].PaymentSubjectYandex)
	// return payment, nil

	type Settlements struct {
		Type string `json:"type"` //'prepayment'
		Amount PaymentAmount `json:"amount"`
	}

	err = Account{Id: paymentYandex.AccountId}.LoadEntity(payment, payment.Id)
	if err != nil {
		return nil, utils.Error{Message: "Не удалось обновить статус платежа - не найден платеж"}
	}

	sendData := struct {
		Customer Customer `json:"customer"`
		PaymentId string `json:"payment_id"`
		Type string `json:"type"` //'payment'
		Send bool `json:"send"` // true
		Items []CartItem `json:"items"`

		//  !!! Это совершенные предоплаты !!!
		Settlements []Settlements `json:"settlements"`
	}{
		Customer: Customer{
			FullName: *order.Customer.Name,
			Email: *order.Customer.Email,
			Phone: *order.Customer.Phone,
		},
		PaymentId: payment.ExternalId,
		Type: "payment",
		Send: true,
		Items: order.CartItems,
		Settlements: []Settlements{
			{Type: "prepayment", Amount: payment.Amount},
		},
	}

	url := "https://payment.yandex.net/api/v3/receipts"

	// Собираем JSON данные
	body, err := json.Marshal(sendData)
	if err != nil {
		return nil, utils.Error{Message: "Не удалось разобрать JSON платежа", Errors: map[string]interface{}{"prepaymentCheck":err.Error()}}
	}

	uuidV4 := uuid.NewV4()

	// crate new request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, utils.Error{Message: "Не удалось создать http-запрос для создания платежа зачета предоплаты",
			Errors: map[string]interface{}{"prepaymentCheck":err.Error()}}
	}

	request.Header.Set("Idempotence-Key", uuidV4.String())
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(strconv.Itoa(paymentYandex.ShopId), paymentYandex.ApiKey)

	// Делаем вызов
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, utils.Error{Message: fmt.Sprintf("Ошибка запроса для Yandex кассы: %v", err),
			Errors: map[string]interface{}{"paymentOption":err.Error()}}
	}
	defer response.Body.Close()


	if response.StatusCode == 200 {

		// todo: повесить флаг отправки зачетного чека
		// fmt.Println("response.StatusCode: ", response.StatusCode)
		// ставим дату Чека закрытия
		/*if err := payment.update(map[string]interface{}{"ConfirmationUrl":confirmation.ConfirmationUrl}); err != nil {
			return err
		}*/
		// Закрываем order:
		
		/*var responseRequest map[string]interface{}
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
		}*/

	} else {
		fmt.Println("Ошибка зачета чека: ", response.StatusCode)
		
		var responseRequest map[string]interface{}
		if err := json.NewDecoder(response.Body).Decode(&responseRequest); err != nil {
			return nil, utils.Error{Message: fmt.Sprintf("Ошибка разбора ответа от Yandex кассы: %v", err),
				Errors: map[string]interface{}{"paymentMode":err.Error()}}
		}
		fmt.Println(responseRequest)
		return nil, utils.Error{Message: fmt.Sprintf("Ответ сервера Яндекс.Кассы: %v", response.StatusCode),
			Errors: map[string]interface{}{"paymentMode":"Проблема с сервером Яндекс.Кассы"}}
	}

	return payment, nil
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
func (paymentYandex *PaymentYandex) GetPreloadDb(autoUpdateOff bool, getModel bool, preload bool) *gorm.DB {

	_db := db

	if autoUpdateOff {
		_db = _db.Set("gorm:association_autoupdate", false)
	}
	if getModel {
		_db = _db.Model(&paymentYandex)
	} else {
		_db = _db.Model(&PaymentYandex{})
	}

	if preload {
		return _db.Preload("WebSite")
	} else {
		return _db
	}
}
