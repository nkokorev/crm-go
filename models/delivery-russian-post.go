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

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

func (DeliveryRussianPost) PgSqlCreate() {
	db.CreateTable(&DeliveryRussianPost{})
	
	db.Model(&DeliveryRussianPost{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	
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

	err := db.First(&deliveryRussianPost, id).Error
	if err != nil {
		return nil, err
	}
	return &deliveryRussianPost, nil
}

func (deliveryRussianPost *DeliveryRussianPost) load() error {

	err := db.First(deliveryRussianPost, deliveryRussianPost.Id).Error
	if err != nil {
		return err
	}
	return nil
}

func (DeliveryRussianPost) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	deliveryRussianPosts := make([]DeliveryRussianPost,0)
	var total uint

	// if need to search
	err := db.Model(&DeliveryRussianPost{}).Limit(100).Order(sortBy).Where( "account_id = ?", accountId).
		Find(&deliveryRussianPosts).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&DeliveryRussianPost{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryRussianPosts))
	for i,_ := range deliveryRussianPosts {
		entities[i] = &deliveryRussianPosts[i]
	}

	return entities, total, nil
}

func (DeliveryRussianPost) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	deliveryRussianPosts := make([]DeliveryRussianPost,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&DeliveryRussianPost{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

		err := db.Model(&DeliveryRussianPost{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
	return db.Set("gorm:association_autoupdate", false).Model(deliveryRussianPost).Omit("id", "account_id").Updates(input).Error
}

func (deliveryRussianPost DeliveryRussianPost) delete () error {
	return db.Model(DeliveryRussianPost{}).Where("id = ?", deliveryRussianPost.Id).Delete(deliveryRussianPost).Error
}

// ########## End of CRUD Entity interface ###########
func (deliveryRussianPost DeliveryRussianPost) CalculateDelivery(deliveryData DeliveryData) (*DeliveryData, error) {

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
		"mass": deliveryData.Weight * float64(1000), // масса в граммах (*1000)
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
		return nil, utils.Error{Message: "Ошибка связи с сервисом Почты России"}
	}

	request.Header.Set("Authorization", Authorization)
	request.Header.Set("X-User-Authorization", XUserAuthorization)
	request.Header.Set("Content-Type", "application/json")

	response, err := client.Do(request)
	if err != nil {
		return nil, utils.Error{Message: "Ошибка связи с сервисом Почты России"}
	}
	defer response.Body.Close()

	// 1. Сначала узнаем статус запроса
	if response.Status == "400 Bad Request" {

		var input struct {
			Desc string `json:"desc"`
		}

		if err := json.NewDecoder(response.Body).Decode(&input); err != nil {
			return nil, err
		}

		return nil, utils.Error{Message: input.Desc}

	} else {

		var input struct {
			TotalRate float64 `json:"total-rate"`
			TotalVat float64 `json:"total-vat"`
		}

		if err := json.NewDecoder(response.Body).Decode(&input); err != nil {
			return nil, utils.Error{Message: "Ошибка данных со стороны Почты России"}
		}

		deliveryData.TotalCost = input.TotalRate / 100 // т.е. в копейках
		return &deliveryData, nil
	}

	return nil, utils.Error{Message: "Ошибка расчета стоимости"}
}

func (deliveryRussianPost DeliveryRussianPost) checkMaxWeight(deliveryData DeliveryData) error {
	// проверяем максимальную массу:
	fmt.Println("deliveryData.Weight: ", deliveryData.Weight)
	fmt.Println("deliveryRussianPost.MaxWeight: ", deliveryRussianPost.MaxWeight)
	if deliveryData.Weight > deliveryRussianPost.MaxWeight {
		return utils.Error{Message: fmt.Sprintf("Превышен максимальный вес посылки в %vкг.", deliveryRussianPost.MaxWeight)}
	}

	return nil
}
