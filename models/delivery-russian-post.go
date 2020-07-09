package models

import (
	"bytes"
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"net/http"
)

type DeliveryRussianPost struct {
	ID     		uint   	`json:"id" gorm:"primary_key"`
	AccountID 	uint	`json:"-" gorm:"index,not null"` // аккаунт-владелец ключа
	ShopID		uint 	`json:"shopId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится
	Code 		string	`json:"code" gorm:"type:varchar(16);default:'russianPost';"` // Для идентификации во фронтенде

	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Курьерская доставка", "Почта России", "Самовывоз"
	AccessToken 		string `json:"accessToken" gorm:"type:varchar(255);"` // accessToken
	XUserAuthorization 	string `json:"xUserAuthorization" gorm:"type:varchar(255);"` // XUserAuthorization в base64
	MaxWeight 	uint `json:"maxWeight" gorm:"type:int;default:20000"` // максимальная масса в граммах

	AddressRequired	bool	`json:"addressRequired" gorm:"type:bool;default:true"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postalCodeRequired" gorm:"type:bool;default:true"` // Требуется ли индекс в адресе доставки
}

func (DeliveryRussianPost) PgSqlCreate() {
	db.CreateTable(&DeliveryRussianPost{})
	
	db.Model(&DeliveryRussianPost{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
}

// ############# Entity interface #############
func (deliveryRussianPost DeliveryRussianPost) getId() uint { return deliveryRussianPost.ID }
func (deliveryRussianPost *DeliveryRussianPost) setId(id uint) { deliveryRussianPost.ID = id }
func (deliveryRussianPost DeliveryRussianPost) GetAccountId() uint { return deliveryRussianPost.AccountID }
func (deliveryRussianPost *DeliveryRussianPost) setAccountId(id uint) { deliveryRussianPost.AccountID = id }

func (deliveryRussianPost DeliveryRussianPost) GetCode() string {
	return deliveryRussianPost.Code
}
// ############# Entity interface #############


// ###### GORM Functional #######
func (DeliveryRussianPost) TableName() string { return "delivery_russian_post" }
func (deliveryRussianPost *DeliveryRussianPost) BeforeCreate(scope *gorm.Scope) error {
	deliveryRussianPost.ID = 0
	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############

func (deliveryRussianPost DeliveryRussianPost) create() (Entity, error)  {
	var newItem Entity = &deliveryRussianPost
	
	if err := db.Create(newItem).Error; err != nil {
		return nil, err
	}

	return newItem, nil
}

func (DeliveryRussianPost) get(id uint) (Entity, error) {

	var deliveryRussianPost DeliveryRussianPost

	err := db.First(&deliveryRussianPost, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryRussianPost, nil
}

func (deliveryRussianPost *DeliveryRussianPost) load() error {

	err := db.First(deliveryRussianPost).Error
	if err != nil {
		return err
	}
	return nil
}

func (DeliveryRussianPost) getPaginationList(accountId uint, offset, limit int, order string, search *string) ([]Entity, error) {

	delivers := make([]DeliveryRussianPost,0)

	err := db.Model(&DeliveryRussianPost{}).Limit(limit).Offset(offset).Order(order).Find(&delivers, "account_id = ?", accountId).Error
	if err != nil {
		return nil, err
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(delivers))
	for i, v := range delivers {
		entities[i] = &v
	}

	return entities, nil
}

func (deliveryRussianPost *DeliveryRussianPost) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(deliveryRussianPost).Omit("id", "account_id").Update(input).Error
}

func (deliveryRussianPost DeliveryRussianPost) delete () error {
	return db.Model(DeliveryRussianPost{}).Where("id = ?", deliveryRussianPost.ID).Delete(deliveryRussianPost).Error
}

// ########## End of CRUD Entity interface ###########
func (deliveryRussianPost DeliveryRussianPost) CalculateDelivery(deliveryData DeliveryData, weight uint) (float64, error) {

	// проверяем максимальную массу:
	if weight*uint(1000) > deliveryRussianPost.MaxWeight {
		return 0, utils.Error{Message: "Превышен максимальный вес посылки в 20кг"}
	}
	url := "https://otpravka-api.pochta.ru/1.0/tariff"

	Authorization := "AccessToken " + deliveryRussianPost.AccessToken
	XUserAuthorization := "Basic " + deliveryRussianPost.XUserAuthorization

	rawJson := utils.MapToRawJson(map[string]interface{}{
		"index-from":	"109390",
		"index-to": 	"107078",
		"mail-category":"ORDINARY",
		"mail-type":"POSTAL_PARCEL",
		"mass": weight*uint(1000), // масса в граммах (*1000)
		"dimension": map[string]interface{}{
			"height": 90, // в см.
			"length": 30, // в см.
			"width": 30, // в см.
		},
		"fragile": false, // отметка "Осторожно хрупкое"
		"with-electronic-notice": true, // уведомление на емейл
		"with-order-of-notice": true, // уведомление заказное
	})

	// response, err := http.Post(url, "application/json", strings.NewReader(""))
	client := &http.Client{}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(rawJson))
	if err != nil {
		return 0, utils.Error{Message: "Ошибка связи с сервисом Почты России"}
	}

	request.Header.Set("Authorization", Authorization)
	request.Header.Set("X-User-Authorization", XUserAuthorization)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return 0, utils.Error{Message: "Ошибка связи с сервисом Почты России"}
	}
	defer response.Body.Close()

	// 1. Сначала узнаем статус реквеста
	if response.Status == "400 Bad Request" {

		var input struct {
			Desc string `json:"desc"`
		}

		if err := json.NewDecoder(response.Body).Decode(&input); err != nil {
			return 0, err
		}

		return 0, utils.Error{Message: input.Desc}

	} else {

		var input struct {
			TotalRate float64 `json:"total-rate"`
			TotalVat float64 `json:"total-vat"`
		}

		if err := json.NewDecoder(response.Body).Decode(&input); err != nil {
			return 0, utils.Error{Message: "Ошибка данных со стороны Почты России"}
		}

		return input.TotalRate, nil
	}

	return 0, utils.Error{Message: "Ошибка расчета стоимости"}
}
