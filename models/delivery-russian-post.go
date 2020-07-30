package models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/utils"
	"net/http"
	"time"
)

type DeliveryRussianPost struct {
	Id     		uint   	`json:"id" gorm:"primary_key"`
	AccountId 	uint	`json:"-" gorm:"index;not null"` // аккаунт-владелец ключа
	WebSiteId		uint 	`json:"webSiteId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится
	Code 		string	`json:"code" gorm:"type:varchar(16);default:'russianPost';"` // Для идентификации во фронтенде

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // активен ли способ доставки
	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Курьерская доставка", "Почта России", "Самовывоз"

	AccessToken 		string `json:"accessToken" gorm:"type:varchar(255);"` // accessToken
	XUserAuthorization 	string `json:"xUserAuthorization" gorm:"type:varchar(255);"` // XUserAuthorization в base64
	MaxWeight 	float64 `json:"maxWeight" gorm:"type:int;default:20"` // максимальная масса в кг
	
	PostalCodeFrom	string	`json:"postalCodeFrom" gorm:"type:varchar(255);"` // индекс отправки с почты России
	MailCategory 	string	`json:"mailCategory" gorm:"type:varchar(50);"` // https://otpravka.pochta.ru/specification#/enums-base-mail-category
	MailType 		string	`json:"mailType" gorm:"type:varchar(50);"` // https://otpravka.pochta.ru/specification#/enums-base-mail-type
	Fragile 		bool	`json:"fragile" gorm:"type:bool;default:false"`  // отметка "Осторожно хрупкое"
	WithElectronicNotice	bool	`json:"withElectronicNotice" gorm:"type:bool;default:true"`  // отметка "Осторожно хрупкое"
	WithOrderOfNotice		bool	`json:"withOrderOfNotice" gorm:"type:bool;default:true"`  // отметка "Осторожно хрупкое"
	WithSimpleNotice		bool	`json:"withSimpleNotice" gorm:"type:bool;default:false"`  // отметка "Осторожно хрупкое"

	AddressRequired	bool	`json:"addressRequired" gorm:"type:bool;default:true"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postalCodeRequired" gorm:"type:bool;default:true"` // Требуется ли индекс в адресе доставки

	// Признак предмета расчета
	PaymentSubjectId	uint	`json:"paymentSubjectId" gorm:"type:int;not null;"`//
	PaymentSubject 		PaymentSubject `json:"paymentSubject"`

	VatCodeId	uint	`json:"vatCodeId" gorm:"type:int;not null;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vatCode"`

	// Разрешенные методы оплаты для данного типа доставки
	PaymentOptions	[]PaymentOption `json:"paymentOptions" gorm:"many2many:payment_options_delivery_russian_posts;preload"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (DeliveryRussianPost) PgSqlCreate() {
	db.CreateTable(&DeliveryRussianPost{})
	
	db.Model(&DeliveryRussianPost{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	db.Model(&DeliveryRussianPost{}).AddForeignKey("payment_subject_id", "payment_subjects(id)", "CASCADE", "CASCADE")
	db.Model(&DeliveryRussianPost{}).AddForeignKey("vat_code_id", "vat_codes(id)", "CASCADE", "CASCADE")
}

// ############# Entity interface #############
func (deliveryRussianPost DeliveryRussianPost) GetId() uint { return deliveryRussianPost.Id }
func (deliveryRussianPost *DeliveryRussianPost) setId(id uint) { deliveryRussianPost.Id = id }
func (deliveryRussianPost DeliveryRussianPost) GetAccountId() uint { return deliveryRussianPost.AccountId }
func (deliveryRussianPost *DeliveryRussianPost) setAccountId(id uint) { deliveryRussianPost.AccountId = id }
func (deliveryRussianPost *DeliveryRussianPost) setShopId(webSiteId uint) { deliveryRussianPost.WebSiteId = webSiteId }
func (deliveryRussianPost DeliveryRussianPost) SystemEntity() bool { return false }

func (deliveryRussianPost DeliveryRussianPost) GetCode() string {
	return deliveryRussianPost.Code
}
// ############# Entity interface #############

// ###### GORM Functional #######
func (deliveryRussianPost *DeliveryRussianPost) BeforeCreate(scope *gorm.Scope) error {
	deliveryRussianPost.Id = 0
	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############
func (deliveryRussianPost DeliveryRussianPost) create() (Entity, error)  {

	drp :=  deliveryRussianPost
	if err := db.Create(&drp).Error; err != nil {
		return nil, err
	}
	var entity Entity = &drp

	return entity, nil
}

func (DeliveryRussianPost) get(id uint) (Entity, error) {

	var deliveryRussianPost DeliveryRussianPost

	err := db.Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").First(&deliveryRussianPost, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryRussianPost, nil
}

func (deliveryRussianPost *DeliveryRussianPost) load() error {

	err := db.Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").First(deliveryRussianPost, deliveryRussianPost.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (DeliveryRussianPost) getList(accountId uint, sortBy string) ([]Entity, uint, error) {
	return DeliveryRussianPost{}.getPaginationList(accountId, 0, 100, sortBy, "")
}

func (DeliveryRussianPost) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	deliveryRussianPosts := make([]DeliveryRussianPost,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&DeliveryRussianPost{}).Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryRussianPosts, "name ILIKE ? OR code ILIKE ? OR postal_code_from ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryRussianPost{}).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR postal_code_from ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&DeliveryRussianPost{}).Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryRussianPosts).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&DeliveryRussianPost{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryRussianPosts))
	for i,_ := range deliveryRussianPosts {
		entities[i] = &deliveryRussianPosts[i]
	}

	return entities, total, nil
}

func (deliveryRussianPost *DeliveryRussianPost) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(deliveryRussianPost).
		Preload("PaymentOptions").Preload("PaymentSubject").Preload("VatCode").Omit("id", "account_id").Updates(input).Error
}

func (deliveryRussianPost *DeliveryRussianPost) delete () error {
	return db.Model(DeliveryRussianPost{}).Where("id = ?", deliveryRussianPost.Id).Delete(deliveryRussianPost).Error
}

// ########## End of CRUD Entity interface ###########
func (deliveryRussianPost DeliveryRussianPost) CalculateDelivery(deliveryData DeliveryData, weight float64) (float64, error) {

	if weight == 0 {
		return 0, utils.Error{Message: "Ошибка расчета стоимости доставки: отсутствует вес товара"}
	}
	// базовые данные для запроса в api почта россиии
	url := "https://otpravka-api.pochta.ru/1.0/tariff"
	Authorization := "AccessToken " + deliveryRussianPost.AccessToken
	XUserAuthorization := "Basic " + deliveryRussianPost.XUserAuthorization

	// Формируем json для запроса
	rawJson := utils.MapToRawJson(map[string]interface{}{
		"index-from":	deliveryRussianPost.PostalCodeFrom,
		"index-to": 	deliveryData.PostalCode,
		"mail-category":deliveryRussianPost.MailCategory,
		"mail-type":deliveryRussianPost.MailType,
		"mass": weight * float64(1000), // масса в граммах (*1000)
		/*"dimension": map[string]interface{}{
			"height": 90, // в см.
			"length": 30, // в см.
			"width": 30, // в см.
		},*/
		"fragile": deliveryRussianPost.Fragile, // отметка "Осторожно хрупкое"
		"with-electronic-notice": deliveryRussianPost.WithElectronicNotice, // уведомление на емейл
		"with-order-of-notice": deliveryRussianPost.WithOrderOfNotice, // уведомление заказное
		"with-simple-notice": deliveryRussianPost.WithSimpleNotice, // уведомление заказное
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

	// 1. Сначала узнаем статус запроса
	if response.StatusCode == 200 {
		var input struct {
			TotalRate float64 `json:"total-rate"`
			TotalVat float64 `json:"total-vat"`
		}

		if err := json.NewDecoder(response.Body).Decode(&input); err != nil {
			return 0, utils.Error{Message: "Ошибка данных со стороны Почты России"}
		}

		return float64(input.TotalRate / 100), nil // т.е. в копейках
		// deliveryData.TotalCost = input.TotalRate / 100 // т.е. в копейках
		// return &deliveryData, nil

	} else {
		var input struct {
			Desc string `json:"desc"`
		}

		if err := json.NewDecoder(response.Body).Decode(&input); err != nil {
			return 0, err
		}

		return 0, utils.Error{Message: input.Desc}
	}


	return 0, utils.Error{Message: "Ошибка расчета стоимости"}
}

func (deliveryRussianPost DeliveryRussianPost) checkMaxWeight(weight float64) error {
	// проверяем максимальную массу:
	if weight > deliveryRussianPost.MaxWeight {
		// return utils.Error{Message: fmt.Sprintf("Превышен максимальный вес посылки в %vкг.", deliveryRussianPost.MaxWeight)}
		return utils.Error{Message: fmt.Sprintf("Превышен максимальный вес посылки в %vкг.", deliveryRussianPost.MaxWeight),
			Errors: map[string]interface{}{"delivery":fmt.Sprintf("Превышен максимальный вес посылки в %vкг.", deliveryRussianPost.MaxWeight)}}
	}

	return nil
}

func (deliveryRussianPost DeliveryRussianPost) GetName () string {
	return deliveryRussianPost.Name
}
func (deliveryRussianPost DeliveryRussianPost) GetVatCode () VatCode {
	return deliveryRussianPost.VatCode
}

func (deliveryRussianPost DeliveryRussianPost) AppendPaymentOptions(paymentOptions []PaymentOption) error  {
	if err := db.Model(&deliveryRussianPost).Association("PaymentOptions").Append(paymentOptions).Error; err != nil {
		return err
	}

	return nil
}
func (deliveryRussianPost DeliveryRussianPost) RemovePaymentOptions(paymentOptions []PaymentOption) error  {
	if err := db.Model(&deliveryRussianPost).Association("PaymentOptions").Delete(paymentOptions).Error; err != nil {
		return err
	}

	return nil
}
func (deliveryRussianPost DeliveryRussianPost) ExistPaymentOption(paymentOptions PaymentOption) bool  {
	return db.Model(&deliveryRussianPost).Where("payment_options.id = ?", paymentOptions.Id).Association("PaymentOptions").Find(&PaymentOption{}).Count() > 0
}

func (deliveryRussianPost DeliveryRussianPost) CreateDeliveryOrder(deliveryData DeliveryData, amount PaymentAmount,order Order) (Entity, error)  {
	deliveryOrder := DeliveryOrder{
		AccountId: deliveryRussianPost.AccountId,
		OrderId:   order.Id,
		CustomerId: order.CustomerId,
		WebSiteId: order.WebSiteId,
		Code:  deliveryRussianPost.Code,
		MethodId: deliveryRussianPost.Id,
		Address: deliveryData.Address,
		PostalCode: deliveryData.PostalCode,
		Amount: amount,
	}

	return deliveryOrder.create()

}