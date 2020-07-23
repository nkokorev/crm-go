package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
	"github.com/satori/go.uuid"
	"net/http"
	"strings"
)

type YandexPayment struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	HashId 		string `json:"hashId" gorm:"type:varchar(12);unique_index;not null;"` // публичный Id для защиты от спама/парсинга
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name 		string 	`json:"name" gorm:"type:varchar(128);default:''"` // Имя интеграции магазина "<name>"

	// ####### API YandexPayment ####### //

	// Для авторизации Basic Auth: username:password
	ApiKey	string	`json:"apiKey" gorm:"type:varchar(128);"` // ApiKey от яндекс кассы
	ShopId	string	`json:"shopId" gorm:"type:int;"` // shop id от яндекс кассы

	// URL для уведомлений со стороны Я.Кассы.
	URL		string 	`json:"url" gorm:"type:varchar(255);"`
	EnabledIncomingNotifications 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // обрабатывать ли уведомления от Я.Кассы.
	// ####### Внутренние данные ####### //

	// Возврат после платежа или отмена для пользователя
	ReturnUrl		string 	`json:"returnUrl" gorm:"type:varchar(255);"`

	// Включен ли данный способ оплаты
	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // обрабатывать ли вебхук
	Description 		string 	`json:"description" gorm:"type:varchar(255);default:''"` // Описание что к чему)

	// Сохранение платежных данных (с их помощью можно проводить повторные безакцептные списания ).
	SavePaymentMethod 	bool 	`json:"savePaymentMethod" gorm:"type:bool;default:false"`

	// Автоматический прием  поступившего платежа. Со стороны Я.Кассы
	Capture	bool	`json:"capture" gorm:"type:bool;default:true"`
}

// ############# Entity interface #############
func (yandexPayment YandexPayment) GetId() uint { return yandexPayment.Id }
func (yandexPayment *YandexPayment) setId(id uint) { yandexPayment.Id = id }
func (yandexPayment YandexPayment) GetAccountId() uint { return yandexPayment.AccountId }
func (yandexPayment *YandexPayment) setAccountId(id uint) { yandexPayment.AccountId = id }
func (YandexPayment) SystemEntity() bool { return false }

// ############# Entity interface #############

func (YandexPayment) PgSqlCreate() {
	db.CreateTable(&YandexPayment{})
	db.Model(&YandexPayment{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}
func (yandexPayment *YandexPayment) BeforeCreate(scope *gorm.Scope) error {
	yandexPayment.Id = 0
	yandexPayment.HashId = strings.ToLower(utils.RandStringBytesMaskImprSrcUnsafe(12, true))
	return nil
}


// ######### CRUD Functions ############
func (yandexPayment YandexPayment) create() (Entity, error)  {
	wb := yandexPayment
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var entity Entity = &wb

	return entity, nil
}

func (YandexPayment) get(id uint) (Entity, error) {

	var yandexPayment YandexPayment

	err := db.First(&yandexPayment, id).Error
	if err != nil {
		return nil, err
	}
	return &yandexPayment, nil
}
func (yandexPayment *YandexPayment) load() error {
	if yandexPayment.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить YandexPayment - не указан  Id"}
	}

	err := db.First(yandexPayment,yandexPayment.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (YandexPayment) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return  YandexPayment{}.getPaginationList(accountId, 0, 100, sortBy, "")
}

func (YandexPayment) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	webHooks := make([]YandexPayment,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&YandexPayment{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks, "name ILIKE ? OR code ILIKE ? OR description ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&YandexPayment{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR description ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&YandexPayment{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&webHooks).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&YandexPayment{}).Where("account_id = ?", accountId).Count(&total).Error
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

func (YandexPayment) getByEvent(eventName string) (*YandexPayment, error) {

	wh := YandexPayment{}

	if err := db.First(&wh, "event_type = ?", eventName).Error; err != nil {
		return nil, err
	}

	return &wh, nil
}

func (yandexPayment *YandexPayment) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).
		Model(yandexPayment).Where("id", yandexPayment.Id).Omit("id", "account_id").Updates(input).Error
}

func (yandexPayment YandexPayment) delete () error {
	return db.Model(YandexPayment{}).Where("id = ?", yandexPayment.Id).Delete(yandexPayment).Error
}
// ######### END CRUD Functions ############


// ########## Work function ############

func (yandexPayment YandexPayment) CreatePaymentByOrder(order Order) (*Payment, error) {

	_p := Payment {
		AccountId: yandexPayment.AccountId,
		Paid: false,
		Amount: Amount{Value: float64(12),Currency: "RUB"},
		Description:  fmt.Sprintf("Заказ №%v в магазине AiroCliamte", order.Id),  // Видит клиент
		PaymentMethod: PaymentMethod{Type: "bank_card"},
		Confirmation: Confirmation{Type: "redirect", ReturnUrl: yandexPayment.ReturnUrl},

		// Чтобы понять какой платеж был оплачен!!!
		Metadata: postgres.Jsonb{ RawMessage: utils.MapToRawJson(map[string]interface{}{
			"orderId":order.Id,
			"accountId":yandexPayment.AccountId,
		})},
		SavePaymentMethod: yandexPayment.SavePaymentMethod,
		OwnerId: yandexPayment.Id,
		Capture: yandexPayment.Capture,
		OwnerType: "yandex_payment",
		OrderId: order.Id,
	}

	// создаем внутри платеж
	entity, err := _p.create()
	if err != nil {
		return nil, err
	}
	payment := entity.(*Payment)

	// Вызываем Yandex для созданного платежа
	err = yandexPayment.ExternalCreate(payment)
	if err != nil {
		return nil, err
	}

	return payment, nil
}

func (yandexPayment YandexPayment) ExternalCreate(payment *Payment) error {

	fmt.Println("Вызываем yandex payment")
	
	url := "https://payment.yandex.net/api/v3/payments"

	// Собираем JSON данные
	body, err := json.Marshal(payment)
	if err != nil {
		return err;
	}
	fmt.Println("Request: ", string(body))

	uuidV4, err := uuid.NewV4()
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return utils.Error{Message: "Не удалось создать UUID для создания платежа"}
	}

	// crate new request
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return utils.Error{Message: "Не удалось создать http-запрос для создания платежа"}
	}

	request.Header.Set("Idempotence-Key", uuidV4.String())
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth(yandexPayment.ShopId, yandexPayment.ApiKey)

	// Делаем вызов
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return utils.Error{Message: fmt.Sprintf("Ошибка запроса для Yandex кассы: %v", err)}
	}
	defer response.Body.Close()

	// fmt.Println("======= Код ответа: ", response.Status)
	// fmt.Println("======= Запрос Body: ", response.Body)
	fmt.Println("========================")

	if response.StatusCode == 200 {
		// var responseRequest Payment
		var responseRequest map[string]interface{}
		if err := json.NewDecoder(response.Body).Decode(&responseRequest); err != nil {
			return utils.Error{Message: fmt.Sprintf("Ошибка разбора ответа от Yandex кассы: %v", err)}
		}

		/*	if err = payment.update(responseRequest); err != nil {
			fmt.Println(err)
			return utils.Error{Message: "Ошибка сохранения responseRequest от Яндекс кассы"}
		}*/
		// fmt.Println(responseRequest)

		var confirmation struct {
			Type 	string `json:"type" gorm:"type:varchar(32);"` // embedded, redirect, external, qr
			ReturnUrl 	string `json:"return_url" gorm:"type:varchar(255);"`
			ConfirmationUrl 	string `json:"confirmation_url" gorm:"type:varchar(255);"`
		}

		jsonString, err := json.Marshal(responseRequest["confirmation"])
		if err != nil {
			return utils.Error{Message: fmt.Sprintf("Ошибка разбора ответа от Yandex кассы 1: %v", err)}
		}

		// fmt.Println("jsonString: ",string(jsonString))

		if err := json.NewDecoder(bytes.NewBuffer(jsonString)).Decode(&confirmation); err != nil {
			return utils.Error{Message: fmt.Sprintf("Ошибка разбора ответа от Yandex кассы: %v", err)}
		}

		fmt.Println("confirmation: ", confirmation.ConfirmationUrl)
/*
		_confirmation_url, ok := responseRequest["confirmation_url"]; if !ok {
			return utils.Error{Message: "Ответ от Яндекса не содержит confirmation_url."}
		}

		confirmation_url, ok := _confirmation_url.(string); if !ok {
			return utils.Error{Message: "Ответ от Яндекса не содержит confirmation_url в нужном формате."}
		}

		if err := payment.update(map[string]interface{}{"confirmation_url":confirmation_url}); err != nil {
			return utils.Error{Message: "Ошибка сохранения confirmation_url"}
		}

		fmt.Println("confirmation_url: ", confirmation_url)*/
		// fmt.Println("Обработанный ответ: ", responseRequest)
		// fmt.Println("Обновленный payment: ", payment)
	} else {
		return utils.Error{Message: fmt.Sprintf("Ответ сервера Яндекс.Кассы: %v", response.StatusCode)}
	}




	// todo: тут мы создаем payment, если все хорошо

	return nil
}