package models

import (
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/mitchellh/mapstructure"
	"github.com/nkokorev/crm-go/utils"
)

// Прообраз торговой точки
type WebSite struct {
	ID     uint   `json:"id" gorm:"primary_key"`
	AccountID uint `json:"-" gorm:"type:int;index;not null;"`

	Name string `json:"name" gorm:"type:varchar(255);default:'Новый магазин';not null;"` // Внутреннее имя сайта
	Hostname string `json:"hostname" gorm:"type:varchar(255);not_null;"` // ratuscrm.com, airoclimate.ru, vetvent.ru, ..
	URL 	string `json:"url" gorm:"type:varchar(255);not_null;"` // https://ratuscrm.com, https://airoclimate.ru, http://vetvent.ru, ..

	// Email DKIM
	DKIMPublicRSAKey string `json:"dkimPublicRsaKey" gorm:"type:text;"` // публичный ключ
	DKIMPrivateRSAKey string `json:"dkimPrivateRsaKey" gorm:"type:text;"` // приватный ключ
	DKIMSelector string `json:"dkimSelector" gorm:"type:varchar(255);default:'dk1'"` // dk1

	// Контактные данные
	Address string `json:"address" gorm:"type:varchar(255);default:null;"` // Публичный физический адрес
	Email string `json:"email" gorm:"type:varchar(255);default:null;"` // Публичный email магазина
	Phone string `json:"phone" gorm:"type:varchar(255);default:null;"` // Публичный телефон

	Type	string 	`json:"type" gorm:"type:varchar(50);not null;"` // имя типа shop, site, ... хз как это использовать, на будущее
	Description string `json:"description" gorm:"type:text;default:''"` // html-описание магазина

	Deliveries 		[]Delivery  `json:"deliveries" gorm:"-"`// `gorm:"polymorphic:Owner;"`
	ProductGroups 	[]ProductGroup `json:"productGroups"`
	EmailBoxes 		[]EmailBox `json:"emailBoxes"` // доступные почтовые ящики с которых можно отправлять
}

func (WebSite) PgSqlCreate() {
	
	db.CreateTable(&WebSite{})
	db.Model(&WebSite{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

// ############# Entity interface #############
func (webSite WebSite) GetId() uint { return webSite.ID }
func (webSite *WebSite) setId(id uint) { webSite.ID = id }
func (webSite WebSite) GetAccountId() uint { return webSite.AccountID }
func (webSite *WebSite) setAccountId(id uint) { webSite.AccountID = id }
func (webSite WebSite) systemEntity() bool {
	return false
}
// ############# END Of Entity interface #############

func (webSite *WebSite) BeforeCreate(scope *gorm.Scope) error {
	webSite.ID = 0
	return nil
}

func (webSite *WebSite) AfterFind() (err error) {
	
	webSite.Deliveries = webSite.GetDeliveryMethods()
	return nil
}

// ######### CRUD Functions ############
func (webSite WebSite) create() (Entity, error)  {

	wb := webSite
	
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	var newItem Entity = &wb
	return newItem, nil
}

func (WebSite) get(id uint) (Entity, error) {

	var webSite WebSite

	err := db.First(&webSite, id).Error
	if err != nil {
		return nil, err
	}
	return &webSite, nil
}

func (webSite *WebSite) load() error {

	err := db.Preload("EmailBoxes").First(webSite).Error
	if err != nil {
		return err
	}
	return nil
}


func (WebSite) getList(accountId uint, sortBy string) ([]Entity, uint, error) {

	webSites := make([]WebSite,0)
	var total uint

	err := db.Model(&WebSite{}).Limit(1000).Order(sortBy).Where( "account_id = ?", accountId).
		Preload("EmailBoxes").Find(&webSites).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, 0, err
	}

	// Определяем total
	err = db.Model(&WebSite{}).Where("account_id = ?", accountId).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webSites))
	for i,_ := range webSites {
		entities[i] = &webSites[i]
	}

	return entities, total, nil
}

func (WebSite) getPaginationList(accountId uint, offset, limit int, sortBy, search string) ([]Entity, uint, error) {

	webSites := make([]WebSite,0)
	var total uint

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&WebSite{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailBoxes").
			Find(&webSites, "name ILIKE ? OR address ILIKE ? OR email ILIKE ? OR phone ILIKE ?", search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WebSite{}).
			Where("account_id = ? AND name ILIKE ? OR address ILIKE ? OR email ILIKE ? OR phone ILIKE ?", accountId, search,search,search,search).
			Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}

	} else {

		err := db.Model(&WebSite{}).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
			Preload("EmailBoxes").
			Find(&webSites).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

		// Определяем total
		err = db.Model(&WebSite{}).Where("account_id = ?", accountId).Count(&total).Error
		if err != nil {
			return nil, 0, utils.Error{Message: "Ошибка определения объема базы"}
		}
	}

	// Преобразуем полученные данные
	entities := make([]Entity,len(webSites))
	for i,_ := range webSites {
		entities[i] = &webSites[i]
	}

	return entities, total, nil
}

func (webSite *WebSite) update(input map[string]interface{}) error {
	return db.Set("gorm:association_autoupdate", false).Model(webSite).Omit("id", "account_id").Updates(input).Preload("EmailBoxes").First(webSite,webSite.ID).Error
}

func (webSite WebSite) delete () error {
	return db.Model(WebSite{}).Where("id = ?", webSite.ID).Delete(webSite).Error
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############

func (account Account) ExistProductGroups(groupId uint) bool {
	if groupId < 1 {
		return false
	}

	return !db.Model(&ProductGroup{}).Where("account_id = ? AND id = ?", account.ID, groupId).First(&ProductGroup{}).RecordNotFound()
}
// ######### END OF ACCOUNT Functions ############

// ######### SHOP PRODUCT Functions ############
func (webSite WebSite) CreateProduct(input Product, card *ProductCard) (*Product, error) {
	input.AccountID = webSite.AccountID

	if input.ExistSKU() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"sku":"Товар с таким SKU уже есть"}}
	}
	if input.ExistModel() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"model":"Товар с такой моделью уже есть"}}
	}

	// Создаем продукт
	product, err := input.create()
	if err != nil {
		return nil, err
	}

	// Добавляем продукт в карточку товара
	if card != nil {
		if err = card.AppendProduct(product); err != nil {
			return nil, err
		}
	}
	
	return product, nil
}

func (webSite WebSite) GetProduct(productId uint) (*Product, error) {

	// Создаем продукт
	product, err := Product{}.get(productId)
	if err != nil {
		return nil, err
	}

	if product.AccountID != webSite.AccountID {
		return nil, utils.Error{Message: "Продукт с указанным id не найден"}
	}

	return product, nil
}

func (webSite WebSite) CreateProductWithCardAndGroup(input Product, newCard ProductCard, groupId *uint) (*Product, error) {
	input.AccountID = webSite.AccountID

	if input.ExistSKU() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"sku":"Товар с таким SKU уже есть"}}
	}
	if input.ExistModel() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"model":"Товар с такой моделью уже есть"}}
	}

	// Создаем продукт
	product, err := input.create()
	if err != nil {
		return nil, err
	}

	// Создаем карточку товара
	newCard.AccountID = webSite.AccountID
	newCard.WebSiteID = webSite.ID
	if groupId != nil {
		newCard.ProductGroupID = groupId
	}
	card, err := newCard.create()
	if err != nil {
		return product, err
	}

	// Добавляем товар в новую карточку
	if err = card.AppendProduct(product); err != nil {
		return nil, err
	}

	return product, nil
}

/////////////////////////

func (webSite WebSite) AppendDeliveryMethod(entity Entity) error {
	return entity.update(map[string]interface{}{"web_site_id":webSite.ID})
}

func (webSite WebSite) GetDeliveryMethods() []Delivery {
	// Находим все необходимые методы
	var posts []DeliveryRussianPost
	if err := db.Model(&DeliveryRussianPost{}).Find(&posts, "account_id = ? AND web_site_id = ?", webSite.AccountID, webSite.ID).Error; err != nil {
		return nil
	}

	var couriers []DeliveryCourier
	if err := db.Model(&DeliveryCourier{}).Find(&couriers, "account_id = ? AND web_site_id = ?", webSite.AccountID, webSite.ID).Error; err != nil {
		return nil
	}

	var pickups []DeliveryPickup
	if err := db.Model(&DeliveryPickup{}).Find(&pickups, "account_id = ? AND web_site_id = ?", webSite.AccountID, webSite.ID).Error; err != nil {
		return nil
	}


	deliveries := make([]Delivery, len(posts)+len(pickups)+len(couriers))
	for i,_ := range posts {
		deliveries[i] = &posts[i]
	}
	for i,_ := range couriers {
		deliveries[i+len(posts)] = &couriers[i]
	}
	for i,_ := range pickups {
		deliveries[i+len(posts)+len(couriers)] = &pickups[i]
	}

/*	fmt.Println("New list: ")
	for i,v := range deliveries {
		fmt.Printf("[%s], %v\n\r",i, v)
	}*/

	return deliveries
}

func (webSite WebSite) CalculateDelivery(deliveryRequest DeliveryRequest) (*DeliveryData, error) {

	delivery, err := webSite.GetDelivery(deliveryRequest.DeliveryMethod.Code, deliveryRequest.DeliveryMethod.ID)
	if err != nil {
		return nil, err
	}

	// 1. Расчет веса посылки
	if deliveryRequest.DeliveryData.NeedToCalculateWeight {
		var weight float64
		weight = 0

		// проходим циклом по продуктам и складываем их общий вес
		for _,v := range deliveryRequest.Cart {
			// 1. Получаем продукт
			product, err := webSite.GetProduct(v.ProductId)
			if err != nil {
				return nil, err
			}

			_w, err := product.GetAttribute(deliveryRequest.DeliveryData.ProductWeightKey)
			wg, ok := _w.(float64)
			if !ok {
				continue
			}
			weight += wg * float64(v.Count)
		}

		deliveryRequest.DeliveryData.Weight = weight
	}

	// 2. Проверяем максимальный вес
	if err := delivery.checkMaxWeight(deliveryRequest.DeliveryData); err != nil {
		fmt.Println("Ошибка макс веса", err)
		return nil, err
	}

	// 3. Проводим расчет стоимости доставки
	deliveryData, err := delivery.CalculateDelivery(deliveryRequest.DeliveryData)
	if err != nil {
		return nil, err
	}

	return deliveryData, nil
}

func (webSite WebSite) CreateDelivery(input map[string]interface{}) (*Entity, error) {

	var delivery Delivery

	codeStr, ok := input["code"];
	if !ok {
		return nil, utils.Error{Message: "Не хватает параметра code в запросе"}
	}

	code, ok := codeStr.(string)

	switch code {
	case "russianPost":
		var deliveryRussianPost DeliveryRussianPost
		if err := mapstructure.Decode(input, &deliveryRussianPost); err != nil {
			return nil, err
		}
		delivery = &deliveryRussianPost
	case "courier":
		var deliveryCourier DeliveryCourier
		if err := mapstructure.Decode(input, &deliveryCourier); err != nil {
			return nil, err
		}
		delivery = &deliveryCourier
	case "pickup":
		var deliveryPickup DeliveryPickup
		if err := mapstructure.Decode(input, &deliveryPickup); err != nil {
			return nil, err
		}
		delivery = &deliveryPickup
	default:
		return nil, utils.Error{Message: "Ошибка в коде типа создаваемого интерфейса"}
	}

	delivery.setShopId(webSite.ID)
	delivery.setAccountId(webSite.AccountID)

	entity, err := delivery.create()
	if err != nil {
		return nil, err
	}
	return &entity, nil

	// return &delivery, nil
}

func (webSite WebSite) GetDelivery(code string, methodId uint) (Delivery, error) {

	// Получаем все варианты доставки (обычно их мало). Можно через switch, но лень потом исправлять баг с новыми типом доставки
	deliveries := webSite.GetDeliveryMethods()

	// Ищем наш вариант доставки
	var delivery Delivery
	for _,v := range deliveries {
		if v.GetCode() == code && v.GetId() == methodId {
			delivery = v
			break
		}
	}

	// Проверяем, удалось ли найти выбранный вариант доставки
	if delivery == nil {
		return nil, utils.Error{Message: "Не верно указан тип доставки"}
	}

	return delivery, nil
}
// 
func (webSite WebSite) UpdateDelivery(input map[string]interface{}) (Delivery, error) {

	// Парсим тип рассылки и ее ID
	code,ok := input["code"].(string)
	methodId, ok := input["id"].(float64)
	if !ok {
		return nil, utils.Error{Message: "Код или id способа доставки не верен"}
	}

	delivery, err := webSite.GetDelivery(code, uint(methodId))
	if err != nil {
		return nil, err
	}

	if delivery.GetAccountId() != webSite.AccountID {
		return nil, utils.Error{Message: "Метод доставки принадлежит другому аккаунту!"}
	}

	err = delivery.update(input)
	if err != nil {
		return nil, err
	}

	return delivery, nil
}

func (webSite WebSite) DeleteDelivery(input map[string]interface{}) error {

	// Парсим тип рассылки и ее ID
	code,ok := input["code"].(string)
	methodId, ok := input["id"].(float64)
	if !ok {
		return utils.Error{Message: "Код или id способа доставки не верен"}
	}

	delivery, err := webSite.GetDelivery(code, uint(methodId))
	if err != nil {
		return err
	}

	if delivery.GetAccountId() != webSite.AccountID {
		return utils.Error{Message: "Метод доставки принадлежит другому аккаунту!"}
	}

	err = delivery.delete()
	if err != nil {
		return err
	}

	return nil
}

func (webSite WebSite) DeliveryListOptions() map[string]interface{} {
	return map[string]interface{}{
		"russianPost": "Почта России",
		"courier": "Курьерская доставка",
		"pickup": "Самовывоз",
	}
}

// Вспомогательная функция
func GetAccountIdByWebSiteId(webSiteId uint) (uint, error) {
	type Result struct {
		AccountId uint `json:"accountId"`
	}
	var result Result
	if err := db.Model(&WebSite{}).Where("id = ?", webSiteId).Scan(&result).Error; err != nil {
		return 0, err
	}

	return result.AccountId, nil
}

func (webSite WebSite) CreateEmailBox(emailBox EmailBox) (Entity, error) {
	emailBox.AccountID = webSite.AccountID
	emailBox.WebSiteID = webSite.ID
	return emailBox.create()
}

func (webSite WebSite) GetEmailBoxList(sortBy string) ([]EmailBox, error) {
	return EmailBox{}.getListByWebSite(webSite.AccountID, webSite.ID, sortBy)
}