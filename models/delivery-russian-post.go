package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
	"net/http"
	"time"
)

type DeliveryRussianPost struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	AccountId 	uint	`json:"-" gorm:"index;not null"` // аккаунт-владелец ключа
	WebSiteId	uint 	`json:"web_site_id" gorm:"type:int;index;"` // магазин, к которому относится
	Code 		string	`json:"code" gorm:"type:varchar(16);default:'russianPost';"` // Для идентификации во фронтенде
	Type 		string	`json:"type" gorm:"type:varchar(32);default:'delivery_russian_posts';"` // Для идентификации

	Enabled 	bool 	`json:"enabled" gorm:"type:bool;default:true"` // активен ли способ доставки
	Name 		string `json:"name" gorm:"type:varchar(255);"` // "Курьерская доставка", "Почта России", "Самовывоз"

	AccessToken 		string 	`json:"access_token" gorm:"type:varchar(255);"` // accessToken
	XUserAuthorization 	string 	`json:"x_user_authorization" gorm:"type:varchar(255);"` // XUserAuthorization в base64
	MaxWeight 			float64 `json:"max_weight" gorm:"type:numeric;default:20"` // максимальная масса в кг

	PostalCodeFrom	string	`json:"postal_code_from" gorm:"type:varchar(255);"` // индекс отправки с почты России
	MailCategory 	string	`json:"mail_category" gorm:"type:varchar(50);"` // https://otpravka.pochta.ru/specification#/enums-base-mail-category
	MailType 		string	`json:"mail_type" gorm:"type:varchar(50);"` // https://otpravka.pochta.ru/specification#/enums-base-mail-type
	Fragile 		bool	`json:"fragile" gorm:"type:bool;default:false"`  // отметка "Осторожно хрупкое"
	WithElectronicNotice	bool	`json:"with_electronic_notice"" gorm:"type:bool;default:true"`  // отметка "Осторожно хрупкое"
	WithOrderOfNotice		bool	`json:"with_order_of_notice" gorm:"type:bool;default:true"`  // отметка "Осторожно хрупкое"
	WithSimpleNotice		bool	`json:"with_simple_notice" gorm:"type:bool;default:false"`  // отметка "Осторожно хрупкое"

	AddressRequired		bool	`json:"address_required" gorm:"type:bool;default:true"` // Требуется ли адрес доставки
	PostalCodeRequired	bool	`json:"postal_code_required" gorm:"type:bool;default:true"` // Требуется ли индекс в адресе доставки

	// Признак предмета расчета
	PaymentSubjectId	*uint	`json:"payment_subject_id" gorm:"type:int;default:1;"`//
	PaymentSubject 		PaymentSubject `json:"payment_subject"`

	VatCodeId	uint	`json:"vat_code_id" gorm:"type:int;not null;default:1;"`// товар или услуга ? [вид номенклатуры]
	VatCode		VatCode	`json:"vat_code"`

	// загружаемый интерфейс
	PaymentMethods		[]PaymentMethod `json:"payment_methods" gorm:"-"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (DeliveryRussianPost) PgSqlCreate() {
	db.Migrator().CreateTable(&DeliveryRussianPost{})

	// db.Model(&DeliveryRussianPost{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	// db.Model(&DeliveryRussianPost{}).AddForeignKey("payment_subject_id", "payment_subjects(id)", "RESTRICT", "CASCADE")
	// db.Model(&DeliveryRussianPost{}).AddForeignKey("vat_code_id", "vat_codes(id)", "RESTRICT", "CASCADE")
	err := db.Exec("ALTER TABLE delivery_russian_posts " +
		"ADD CONSTRAINT delivery_russian_posts_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE," +
		"ADD CONSTRAINT delivery_russian_posts_payment_subject_id_fkey FOREIGN KEY (payment_subject_id) REFERENCES payment_subjects(id) ON DELETE RESTRICT ON UPDATE CASCADE," +
		"ADD CONSTRAINT delivery_russian_posts_vat_code_id_fkey FOREIGN KEY (vat_code_id) REFERENCES vat_codes(id) ON DELETE RESTRICT ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

// ############# Entity interface #############
func (deliveryRussianPost DeliveryRussianPost) GetId() uint { return deliveryRussianPost.Id }
func (deliveryRussianPost *DeliveryRussianPost) setId(id uint) { deliveryRussianPost.Id = id }
func (deliveryRussianPost *DeliveryRussianPost) setPublicId(id uint) { }
func (deliveryRussianPost DeliveryRussianPost) GetAccountId() uint { return deliveryRussianPost.AccountId }
func (deliveryRussianPost DeliveryRussianPost) GetWebSiteId() uint { return deliveryRussianPost.WebSiteId }
func (deliveryRussianPost *DeliveryRussianPost) setAccountId(id uint) { deliveryRussianPost.AccountId = id }
func (deliveryRussianPost *DeliveryRussianPost) setWebSiteId(webSiteId uint) { deliveryRussianPost.WebSiteId = webSiteId }
func (deliveryRussianPost DeliveryRussianPost) SystemEntity() bool { return false }

func (deliveryRussianPost DeliveryRussianPost) GetCode() string {
	return deliveryRussianPost.Code
}
func (deliveryRussianPost DeliveryRussianPost) GetType() string {
	return deliveryRussianPost.Type
}
func (deliveryRussianPost DeliveryRussianPost) GetPaymentSubject() PaymentSubject {
	return deliveryRussianPost.PaymentSubject
}
// ############# Entity interface #############

// ###### GORM Functional #######
func (deliveryRussianPost *DeliveryRussianPost) BeforeCreate(tx *gorm.DB) error {
	deliveryRussianPost.Id = 0
	return nil
}
func (deliveryRussianPost *DeliveryRussianPost) AfterFind(tx *gorm.DB) (err error) {

	methods, err := GetPaymentMethodsByDelivery(deliveryRussianPost)
	if err != nil { return err }
	deliveryRussianPost.PaymentMethods = methods

	return nil
}
// ###### End of GORM Functional #######

// ############# CRUD Entity interface #############
func (deliveryRussianPost DeliveryRussianPost) create() (Entity, error)  {

	_item := deliveryRussianPost
	if err := db.Create(&_item).Error; err != nil {
		return nil, err
	}

	if err := _item.GetPreloadDb(false,false, nil).First(&_item,_item.Id).Error; err != nil {
		return nil, err
	}

	var entity Entity = &_item

	return entity, nil
}

func (DeliveryRussianPost) get(id uint, preloads []string) (Entity, error) {

	var item DeliveryRussianPost

	err := item.GetPreloadDb(false, false, preloads).First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (deliveryRussianPost *DeliveryRussianPost) load(preloads []string) error {
	if deliveryRussianPost.Id < 1 {
		return utils.Error{Message: "Невозможно загрузить DeliveryRussianPost - не указан  Id"}
	}

	err := deliveryRussianPost.GetPreloadDb(false, false, preloads).First(deliveryRussianPost, deliveryRussianPost.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (deliveryRussianPost *DeliveryRussianPost) loadByPublicId(preloads []string) error {
	return errors.New("Нет возможности загрузить объект по Public Id")
}

func (DeliveryRussianPost) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return DeliveryRussianPost{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (DeliveryRussianPost) getListByShop(accountId, websiteId uint) ([]DeliveryRussianPost, error) {

	deliveryRussianPosts := make([]DeliveryRussianPost,0)

	err := (&DeliveryRussianPost{}).GetPreloadDb(false,false, []string{"PaymentMethods"}).
		Limit(100).Where( "account_id = ? AND web_site_id = ?", accountId, websiteId).
		Find(&deliveryRussianPosts).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}

	return deliveryRussianPosts, nil
}

func (DeliveryRussianPost) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	deliveryRussianPosts := make([]DeliveryRussianPost,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&DeliveryRussianPost{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryRussianPosts, "name ILIKE ? OR code ILIKE ? OR postal_code_from ILIKE ?", search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&DeliveryRussianPost{}).GetPreloadDb(false,false, nil).
			Where("account_id = ? AND name ILIKE ? OR code ILIKE ? OR postal_code_from ILIKE ?", accountId, search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := (&DeliveryRussianPost{}).GetPreloadDb(false,false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Find(&deliveryRussianPosts).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = (&DeliveryRussianPost{}).GetPreloadDb(false,false, nil).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(deliveryRussianPosts))
	for i := range deliveryRussianPosts {
		entities[i] = &deliveryRussianPosts[i]
	}

	return entities, total, nil
}

func (deliveryRussianPost *DeliveryRussianPost) update(input map[string]interface{}, preloads []string) error {
	delete(input,"payment_subject")
	delete(input,"vat_code")
	delete(input,"payment_methods")
	utils.FixInputHiddenVars(&input)
	if err := utils.ConvertMapVarsToUINT(&input, []string{"web_site_id","payment_subject_id","vat_code_id"}); err != nil {
		return err
	}

	if err := deliveryRussianPost.GetPreloadDb(false, false, nil).Where("id = ?", deliveryRussianPost.Id).Omit("id", "account_id").Updates(input).
		Error; err != nil {return err}

	err := deliveryRussianPost.GetPreloadDb(false,false, preloads).First(deliveryRussianPost, deliveryRussianPost.Id).Error
	if err != nil {
		return err
	}

	return nil

}

func (deliveryRussianPost *DeliveryRussianPost) delete () error {
	return deliveryRussianPost.GetPreloadDb(true,false,nil).Where("id = ?", deliveryRussianPost.Id).Delete(deliveryRussianPost).Error
}

// ########## End of CRUD Entity interface ###########
func (deliveryRussianPost DeliveryRussianPost) CalculateDelivery(deliveryData DeliveryData, weight float64) (float64, error) {

	if weight == 0 {
		// Грязный хак, в кг
		weight = 1
		// return 0, utils.Error{Message: "Ошибка расчета стоимости доставки: отсутствует вес товара"}
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

func (deliveryRussianPost DeliveryRussianPost) AppendPaymentMethods(paymentMethods []PaymentMethod) error  {
	return nil
}
func (deliveryRussianPost DeliveryRussianPost) RemovePaymentMethods(paymentMethods []PaymentMethod) error  {
	return nil
}
func (deliveryRussianPost DeliveryRussianPost) ExistPaymentMethod(paymentMethod PaymentMethod) bool  {
	return true
}

func (deliveryRussianPost DeliveryRussianPost) CreateDeliveryOrder(deliveryData DeliveryData, amount PaymentAmount,order Order) (Entity, error)  {

	status, err := DeliveryStatus{}.GetStatusNew()
	if err != nil { return nil, err}

	deliveryOrder := DeliveryOrder{
		AccountId: deliveryRussianPost.AccountId,
		OrderId:   &order.Id,
		CustomerId: order.CustomerId,
		WebSiteId: order.WebSiteId,
		Code:  deliveryRussianPost.Code,
		MethodId: deliveryRussianPost.Id,
		Address: deliveryData.Address,
		PostalCode: deliveryData.PostalCode,
		Amount: amount,
		StatusId: status.Id,
	}

	return deliveryOrder.create()

}

func (deliveryRussianPost *DeliveryRussianPost) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(deliveryRussianPost)
	} else {
		_db = _db.Model(&DeliveryRussianPost{})
	}

	if autoPreload {
		return _db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"PaymentSubject","VatCode"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}