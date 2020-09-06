package models

import (
	"database/sql"
	"errors"
	"github.com/mitchellh/mapstructure"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

// Прообраз торговой точки
type WebSite struct {
	Id     		uint   	`json:"id" gorm:"primaryKey"`
	PublicId	uint   	`json:"public_id" gorm:"index;precision:0;"`
	AccountId 	uint 	`json:"-" gorm:"type:int;index;not null;"`

	Name 		string `json:"name" gorm:"type:varchar(255);default:'Новый магазин';not null;"` // Внутреннее имя сайта
	Hostname 	string `json:"hostname" gorm:"type:varchar(255);not_null;"` // ratuscrm.com, airoclimate.ru, vetvent.ru, ..
	URL 		string `json:"url" gorm:"type:varchar(255);not_null;"` // https://ratuscrm.com, https://airoclimate.ru, http://vetvent.ru, ..

	// Email DKIM
	DKIMPublicRSAKey 	string `json:"dkim_public_rsa_key" gorm:"type:text;"` // публичный ключ
	DKIMPrivateRSAKey 	string `json:"dkim_private_rsa_key" gorm:"type:text;"` // приватный ключ
	DKIMSelector 		string `json:"dkim_selector" gorm:"type:varchar(255);default:'dk1'"` // dk1

	// Контактные данные
	Address 	*string `json:"address" gorm:"type:varchar(255);"` // Публичный физический адрес
	Email 		*string `json:"email" gorm:"type:varchar(255);"` // Публичный email магазина
	Phone 		*string `json:"phone" gorm:"type:varchar(255);"` // Публичный телефон

	Type		string 	`json:"type" gorm:"type:varchar(50);not null;"` // имя типа shop, site, ... хз как это использовать, на будущее
	Description *string `json:"description" gorm:"type:text;"` // html-описание магазина

	Deliveries 		[]Delivery  `json:"deliveries" gorm:"-"`// `gorm:"polymorphic:Owner;"`
	// ProductGroups 	[]ProductGroup `json:"productGroups" gorm:"-"`
	WebPages 		[]WebPage 	`json:"web_pages"`
	EmailBoxes 		[]EmailBox `json:"email_boxes" gorm:"preload"` // доступные почтовые ящики с которых можно отправлять
	// PaymentOptions 	[]PaymentOption `json:"paymentOptions" gorm:"many2many:payment_options_web_sites;preload"` // доступные почтовые ящики с которых можно отправлять
}

func (WebSite) PgSqlCreate() {
	
	if err := db.Migrator().CreateTable(&WebSite{}); err != nil {
		log.Fatal(err)
	}
	// db.Model(&WebSite{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE web_sites ADD CONSTRAINT web_sites_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n-- alter table web_sites alter column default_account_id set default null;\n-- alter table web_sites alter column invited_user_id set default null;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

}

// ############# Entity interface #############
func (webSite WebSite) GetId() uint { return webSite.Id }
func (webSite *WebSite) setId(id uint) { webSite.Id = id }
func (webSite *WebSite) setPublicId(id uint) { webSite.PublicId = id }
func (webSite WebSite) GetAccountId() uint { return webSite.AccountId }
func (webSite *WebSite) setAccountId(id uint) { webSite.AccountId = id }
func (webSite WebSite) SystemEntity() bool { return false }
// ############# END Of Entity interface #############

func (webSite *WebSite) GetPreloadDb(getModel bool, autoPreload bool, preloads []string) *gorm.DB {

	_db := db

	if getModel {
		_db = _db.Model(&webSite)
	} else {
		_db = _db.Model(&WebSite{})
	}

	if autoPreload {
		return db.Preload(clause.Associations)
	} else {

		allowed := utils.FilterAllowedKeySTRArray(preloads,[]string{"EmailBoxes","WebPages"})

		for _,v := range allowed {
			_db.Preload(v)
		}
		return _db
	}

}


func (webSite *WebSite) BeforeCreate(tx *gorm.DB) error {
	webSite.Id = 0

	var lastIdx sql.NullInt64

	row := db.Model(&WebSite{}).Where("account_id = ?",  webSite.AccountId).
		Select("max(public_id)").Row()
	if row != nil {
		err := row.Scan(&lastIdx)
		if err != nil && err != gorm.ErrRecordNotFound { return err }
	}

	webSite.PublicId = 1 + uint(lastIdx.Int64)
	return nil
}
func (webSite *WebSite) BeforeUpdate(tx *gorm.DB) (err error) {
	// fmt.Println(tx.Statement.Select("public_id"))
	// fmt.Println(tx.Statement.Clauses)
	// fmt.Println(tx.Statement.Context.Value("public_id"))
	/*if tx.Statement.Changed("public_id") {
		tx.Statement.SetColumn("public_id", webSite.PublicId)
	}*/
	return
}
func (webSite *WebSite) AfterFind(tx *gorm.DB) (err error) {
	
	webSite.Deliveries = webSite.GetDeliveryMethods()
	return nil
}

func (webSite *WebSite) AfterCreate(tx *gorm.DB) error {
	event.AsyncFire(Event{}.WebSiteCreated(webSite.AccountId, webSite.Id))
	return nil
}
func (webSite *WebSite) AfterUpdate(tx *gorm.DB) (err error) {
	event.AsyncFire(Event{}.WebSiteUpdated(webSite.AccountId, webSite.Id))
	return nil
}
func (webSite *WebSite) AfterDelete(tx *gorm.DB) (err error) {
	event.AsyncFire(Event{}.WebSiteDeleted(webSite.AccountId, webSite.Id))
	return nil
}

// ######### CRUD Functions ############
func (webSite WebSite) create() (Entity, error)  {

	wb := webSite
	
	if err := db.Create(&wb).Error; err != nil {
		return nil, err
	}

	if err := wb.GetPreloadDb(false,true,nil).First(&wb,wb.Id).Error; err != nil {
		return nil, err
	}

	var newItem Entity = &wb
	return newItem, nil
}
func (WebSite) get(id uint, preloads []string) (Entity, error) {

	var webSite WebSite

	err := (&WebSite{}).GetPreloadDb(false,false,nil).First(&webSite, id).Error
	if err != nil {
		return nil, err
	}
	return &webSite, nil
}
func (webSite *WebSite) load(preloads []string) error {

	err := webSite.GetPreloadDb(false, false, nil).First(webSite,webSite.Id).Error
	if err != nil {
		return err
	}
	return nil
}
func (webSite *WebSite) loadByPublicId(preloads []string) error {


	if webSite.PublicId < 1 {
		return utils.Error{Message: "Невозможно загрузить Payment - не указан  Id"}
	}

	if err := webSite.GetPreloadDb(false,false,nil).
		First(webSite, "account_id = ? AND public_id = ?", webSite.AccountId, webSite.PublicId).Error; err != nil {
		return err
	}

	return nil
}
func (WebSite) getList(accountId uint, sortBy string, preload []string) ([]Entity, int64, error) {
	return WebSite{}.getPaginationList(accountId, 0, 100, sortBy, "",nil, preload)
}
func (WebSite) getPaginationList(accountId uint, offset, limit int, sortBy, search string, filter map[string]interface{},preloads []string) ([]Entity, int64, error) {

	webSites := make([]WebSite,0)
	var total int64

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := (&WebSite{}).GetPreloadDb(false, false, nil).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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

		err := (&WebSite{}).GetPreloadDb(false, false, preloads).Limit(limit).Offset(offset).Order(sortBy).Where( "account_id = ?", accountId).
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
	for i := range webSites {
		entities[i] = &webSites[i]
	}

	return entities, total, nil
}
func (webSite *WebSite) update(input map[string]interface{}, preloads []string) error {
	delete(input,"email_boxes")
	delete(input,"web_pages")
	delete(input,"deliveries")

	// fmt.Printf("Type: %T | %v\n", input["public_id"], input["public_id"])
	if err := utils.ConvertMapVarsToUINT(&input, []string{"public_id"}); err != nil {
		return err
	}

	if err := webSite.GetPreloadDb(false, false, nil).Where("id = ?", webSite.Id).Omit("id", "account_id","public_id").Updates(input).
		Error; err != nil {return err}

	err := webSite.GetPreloadDb(false,false, preloads).First(webSite, webSite.Id).Error
	if err != nil {
		return err
	}

	return nil
}
func (webSite *WebSite) delete () error {
	return webSite.GetPreloadDb(true,false, nil).Where("id = ?", webSite.Id).Delete(webSite).Error
}
// ######### END CRUD Functions ############

// ######### ACCOUNT Functions ############

func (webSite WebSite) CreatePage(input WebPage) (*WebPage, error) {
	input.AccountId = webSite.AccountId
	input.WebSiteId = &webSite.Id
	
	pEntity, err := input.create(); if err != nil {
		return nil, err
	}
	page, ok := pEntity.(*WebPage)
	if !ok {
		return nil, utils.Error{Message: "Ошибка преобразования WebPage"}
	}

	if err := db.Model(&webSite).Association("WebPages").Append(page); err != nil {
		return nil, err
	}

	return page, nil
}
// ######### END OF ACCOUNT Functions ############

// ######### SHOP PRODUCT Functions ############
func (webSite WebSite) CreateProduct(input Product, card ProductCard) (*Product, error) {
	input.AccountId = webSite.AccountId

	if input.ExistSKU() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"sku":"Товар с таким SKU уже есть"}}
	}
	if input.ExistModel() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"model":"Товар с такой моделью уже есть"}}
	}

	// Создаем продукт
	_p, err := input.create()
	if err != nil {
		return nil, err
	}

	product, ok := _p.(*Product)
	if !ok {
		return nil, utils.Error{Message: "Ошибка преобразования Product"}
	}

	// Добавляем продукт в карточку товара
	if err := card.AppendProduct(product); err != nil {
		return nil, err
	}
	
	return product, nil
}
func (webSite WebSite) GetProduct(productId uint) (*Product, error) {

	// Создаем продукт
	_p, err := Product{}.get(productId, nil)
	if err != nil {
		return nil, err
	}

	product, ok := _p.(*Product)
	if !ok {
		return nil, utils.Error{Message: "Ошибка преобразования Product"}
	}

	if product.AccountId != webSite.AccountId {
		return nil, utils.Error{Message: "Продукт с указанным id не найден"}
	}

	return product, nil
}
func (webSite WebSite) CreateProductWithProductCard(input Product, newCard ProductCard, webPage WebPage) (*Product, error) {

	input.AccountId = webSite.AccountId

	if input.ExistSKU() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"sku":"Товар с таким SKU уже есть"}}
	}
	if input.ExistModel() {
		return nil, utils.Error{Message: "Повторение данных", Errors: map[string]interface{}{"model":"Товар с такой моделью уже есть"}}
	}

	// Создаем продукт
	_p, err := input.create()
	if err != nil {
		return nil, err
	}

	product, ok := _p.(*Product)
	if !ok {
		return nil, utils.Error{Message: "Ошибка преобразования Product"}
	}

	// Создаем карточку товара
	newCard.AccountId = webSite.AccountId
	newCard.WebSiteId = &webSite.Id
	// newCard.WebPageId = &webPageId
	
	cardE, err := newCard.create()
	if err != nil {
		return nil, err
	}

	card, ok := cardE.(*ProductCard)
	if !ok {
		return nil, errors.New("Ошибка преобразования")
	}

	// Добавляем товар в новую карточку
	if err = card.AppendProduct(product); err != nil {
		return nil, err
	}

	if err = webPage.AppendProductCard(card);err != nil {
		return nil, err
	}

	return product, nil
}

/////////////////////////

func (webSite WebSite) AppendDeliveryMethod(entity Entity) error {
	return entity.update(map[string]interface{}{"web_site_id":webSite.Id}, nil)
}

// todo: пофиксить выпуск публичного ключа через UI / API
func (webSite WebSite) GetDeliveryMethods() []Delivery {

	posts := make([]DeliveryRussianPost,0)
	posts, err := DeliveryRussianPost{}.getListByShop(webSite.AccountId, webSite.Id)
	if err != nil { return nil }

	couriers := make([]DeliveryCourier,0)
	couriers, err = DeliveryCourier{}.getListByShop(webSite.AccountId, webSite.Id)
	if err != nil { return nil }

	pickups := make([]DeliveryPickup,0)
	pickups, err = DeliveryPickup{}.getListByShop(webSite.AccountId, webSite.Id)
	if err != nil { return nil }

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

	return deliveries
}

func (webSite WebSite) CalculateDelivery(deliveryRequest DeliveryRequest) (totalCost float64, weight float64, err error) {

	delivery, err := webSite.GetDelivery(deliveryRequest.DeliveryData.Code, deliveryRequest.DeliveryData.Id)
	if err != nil {
		return 0, 0,err
	}

	// проходим циклом по продуктам и складываем их общий вес
	for _,v := range deliveryRequest.Cart {
		// 1. Получаем продукт
		product, err := webSite.GetProduct(v.ProductId)
		if err != nil {
			return 0, 0,err
		}

		// todo: 'что есть вес?'
		// _w, err := product.GetAttribute(deliveryRequest.DeliveryData.ProductWeightKey)
		_w, err := product.GetAttribute(product.WeightKey)
		if err != nil || _w == nil {
			// log.Println("Ошибка получения веса: ", err)
			continue
		}
		wg, ok := _w.(float64)
		if !ok {
			// fmt.Println("product.WeightKey: ", product.WeightKey)
			// fmt.Println("Не учитываем вес!")
			continue
		}
		weight += wg * float64(v.Quantity)
	}

	// 2. Проверяем максимальный вес
	if err := delivery.checkMaxWeight(weight); err != nil {
		return 0, 0, err
	}

	// 3. Проводим расчет стоимости доставки
	totalCost, err = delivery.CalculateDelivery(deliveryRequest.DeliveryData, weight)
	if err != nil {
		return 0,0, err
	}

	return totalCost, weight, nil
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

	delivery.setWebSiteId(webSite.Id)
	delivery.setAccountId(webSite.AccountId)

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

func (webSite WebSite) UpdateDelivery(input map[string]interface{}) (Delivery, error) {

	// Парсим тип рассылки и ее Id
	code,ok := input["code"].(string)
	methodId, ok := input["id"].(float64)
	if !ok {
		return nil, utils.Error{Message: "Код или id способа доставки не верен"}
	}

	delivery, err := webSite.GetDelivery(code, uint(methodId))
	if err != nil {
		return nil, err
	}

	if delivery.GetAccountId() != webSite.AccountId {
		return nil, utils.Error{Message: "Метод доставки принадлежит другому аккаунту!"}
	}

	err = delivery.update(input,nil)
	if err != nil {
		return nil, err
	}

	return delivery, nil
}

func (webSite WebSite) DeleteDelivery(input map[string]interface{}) error {

	// Парсим тип рассылки и ее Id
	code,ok := input["code"].(string)
	methodId, ok := input["id"].(float64)
	if !ok {
		return utils.Error{Message: "Код или id способа доставки не верен"}
	}

	delivery, err := webSite.GetDelivery(code, uint(methodId))
	if err != nil {
		return err
	}

	if delivery.GetAccountId() != webSite.AccountId {
		return utils.Error{Message: "Метод доставки принадлежит другому аккаунту!"}
	}

	err = delivery.delete()
	if err != nil {
		return err
	}

	return nil
}

func (webSite WebSite) DeliveryCodeList() map[string]interface{} {
	return map[string]interface{}{
		"russianPost": "Почта России",
		"courier": "Курьерская доставка",
		"pickup": "Самовывоз",
	}
}

// Вспомогательная функция
func GetAccountIdByWebSiteId(webSiteId uint) (uint, error) {
	type Result struct {
		AccountId uint `json:"account_id"`
	}
	var result Result
	if err := db.Model(&WebSite{}).Where("id = ?", webSiteId).Scan(&result).Error; err != nil {
		return 0, err
	}

	return result.AccountId, nil
}

func (webSite WebSite) CreateEmailBox(emailBox EmailBox) (Entity, error) {
	emailBox.AccountId = webSite.AccountId
	emailBox.WebSiteId = webSite.Id
	return emailBox.create()
}
func (webSite WebSite) GetEmailBoxList(sortBy string) ([]EmailBox, error) {
	return EmailBox{}.getListByWebSite(webSite.AccountId, webSite.Id, sortBy)
}

func (webSite WebSite) ValidateDKIM() error {
	if len(webSite.DKIMPrivateRSAKey) < 1 {
		return utils.Error{Message: "Необходимо указать валидный DKIM Private Key"}
	}
	if len(webSite.DKIMPublicRSAKey) < 1 {
		return utils.Error{Message: "Необходимо указать валидный DKIM Public Key"}
	}
	if len(webSite.DKIMSelector) < 1 {
		return utils.Error{Message: "Необходимо указать селектор DKIM"}
	}

	return nil
}