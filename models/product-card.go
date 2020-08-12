package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
)

// Карточка "товара" в магазине в котором могут быть разные торговые предложения
type ProductCard struct {
	Id     				uint `json:"id" gorm:"primary_key"`
	AccountId 			uint `json:"-" gorm:"type:int;index;not null;"` // потребуется, если productGroupId == null
	WebSiteId 			uint `json:"webSiteId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится
	ProductGroupId 		*uint `json:"productGroupId" gorm:"type:int;index;default:NULL;"` // группа товаров, категория товаров

	Enabled 			bool 	`json:"enabled" gorm:"type:bool;default:true"` // активна ли карточка товара
	URL 				string `json:"url" gorm:"type:varchar(255);"` // идентификатор страницы (products/syao-chzhun )
	Breadcrumb 			string `json:"breadcrumb" gorm:"type:varchar(255);default:null;"`
	Label	 			string `json:"label" gorm:"type:varchar(255);default:'';"` // что выводить в список товаров

	MetaTitle 			string `json:"metaTitle" gorm:"type:varchar(255);default:null;"`
	MetaKeywords 		string `json:"metaKeywords" gorm:"type:varchar(255);default:null;"`
	MetaDescription 	string `json:"metaDescription" gorm:"type:varchar(255);default:null;"`

	// Full description нет т.к. в карточке описание берется от офера
	ShortDescription 	string `json:"shortDescription" gorm:"type:varchar(255);default:null;"` // для превью карточки товара
	Description 		string `json:"description" gorm:"type:text;default:null;"` // фулл описание товара

	// Хелперы карточки: переключение по цветам, размерам и т.д.
	// SwitchProducts	 	*pq.StringArray `json:"switchProducts" sql:"type:varchar(255)[];default:null"` // {color, size} Параметры переключения среди предложений
	//SwitchProducts	 	pq.StringArray `json:"switchProducts" sql:"type:varchar(255)[];default:'{}'"` // {color, size} Параметры переключения среди предложений
	//SwitchProducts	 	[]string `json:"switchProducts" gorm:"type:JSONB;DEFAULT '{}'::JSONB"` // {color, size} Параметры переключения среди предложений
	SwitchProducts	 	postgres.Jsonb `json:"switchProducts" gorm:"type:JSONB;DEFAULT '{}'::JSONB"` // {color, size} Параметры переключения среди предложений

	// ProductGroups 		[]ProductGroup `json:"productGroups" gorm:"many2many:product_group_product_cards"` // для разделов new и т.д.
	ProductGroup 		ProductGroup `json:"-"` // для разделов new и т.д.
	WebSite		 		WebSite `json:"-"` // к какому магазину относится
	Products 			[]Product `json:"products" gorm:"many2many:product_card_products"` // можно заводить два схожих продукта, разных по цвету
}

func (ProductCard) PgSqlCreate() {
	db.CreateTable(&ProductCard{})
	db.Model(&ProductCard{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
}

func (productCard *ProductCard) BeforeCreate(scope *gorm.Scope) error {
	productCard.Id = 0
	return nil
}


func (productCard ProductCard) getId() uint {
	return productCard.Id
}

// ######### CRUD Functions ############
func (productCard ProductCard) create() (*ProductCard, error) {
	var productCardNew = productCard
	if err := db.Create(&productCardNew).First(&productCardNew).Error; err != nil {
		return nil, err
	}

	event.AsyncFire(Event{}.ProductCardCreated(productCardNew.AccountId, productCardNew.Id))

	return &productCardNew, nil
}

func (ProductCard) get(id uint) (*ProductCard, error) {

	card := ProductCard{}

	if err := db.Preload("Products").First(&card, id).Error; err != nil {
		return nil, err
	}

	return &card, nil
}

func (ProductCard) getByAccount(id, accountId uint) (*ProductCard, error) {

	card := ProductCard{}

	if err := db.Preload("Products").First(&card, "id = ? AND account_id = ?", id, accountId).Error; err != nil {
		return nil, err
	}

	return &card, nil
}

func (ProductCard) getListByShop(webSiteId uint) ([]ProductCard, error) {

	cards := make([]ProductCard,0)

	err := db.Preload("Products").Find(&cards, "webSiteId = ?", webSiteId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return cards, nil
	
}

func (ProductCard) getListByAccount(accountId uint) ([]ProductCard, error) {

	cards := make([]ProductCard,0)

	err := db.Preload("Products").Find(&cards, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return cards, nil
}

func (productCard *ProductCard) update(input map[string]interface{}) error {
	if err := db.Model(productCard).Omit("id", "account_id").Updates(input).Find(productCard).Error; err != nil {
		return err
	}

	event.AsyncFire(Event{}.ProductCardUpdated(productCard.AccountId, productCard.Id))

	return nil
}

func (productCard *ProductCard) delete () error {
	if err := db.Model(ProductCard{}).Where("id = ?", productCard.Id).Delete(productCard).Error; err != nil { return err }

	event.AsyncFire(Event{}.ProductCardDeleted(productCard.AccountId, productCard.Id))

	return nil
}
// ######### END CRUD Functions ############

// ######### SHOP PRODUCT Functions ############
func (webSite WebSite) CreateProductCard(input ProductCard, group *ProductGroup) (*ProductCard, error) {

	if webSite.Id < 1 {
		return nil, utils.Error{Message: "Не верно указан id магазина"}
	}
	
	input.WebSiteId = webSite.Id
	input.AccountId = webSite.AccountId

	if group != nil {
		return group.CreateProductCard(&input);
	} else {
		return input.create()
	}

}

func (webSite WebSite) GetProductCard(cardId uint) (*ProductCard, error) {
	return ProductCard{}.get(cardId)
}

func (webSite WebSite) GetProductCardList(offset, limit int, search string, products bool) ([]ProductCard, uint, error) {
	cards := make([]ProductCard,0)

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&ProductCard{}).Preload("Products").Limit(limit).Offset(offset).Where("account_id = ?", webSite.AccountId).Order("id").
			Find(&cards, "url ILIKE ? OR breadcrumb ILIKE ? OR meta_title ILIKE ? OR meta_keywords ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ?" , search,search,search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		if products {
			err := db.Model(&ProductCard{}).Preload("Products").Limit(limit).Offset(offset).Order("id").Find(&cards, "account_id = ? AND web_site_id = ?", webSite.AccountId, webSite.Id).Error
			if err != nil && err != gorm.ErrRecordNotFound{
				return nil, 0, err
			}
		} else {
			err := db.Model(&ProductCard{}).Preload("Products").Limit(limit).Offset(offset).Find(&cards, "account_id = ? AND web_site_id = ?", webSite.AccountId, webSite.Id).Error
			if err != nil && err != gorm.ErrRecordNotFound{
				return nil, 0, err
			}
		}



	}

	// len(cards) != всему списку!
	var total uint
	err := db.Model(&ProductCard{}).Where("account_id = ? AND web_site_id = ?", webSite.AccountId, webSite.Id).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема"}
	}

	return cards, total, nil
}

func (webSite WebSite) DeleteProductCard(cardId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	card, err := webSite.GetProductCard(cardId)
	if err != nil {
		return err
	}

	return card.delete()
}

// #####################

func (account Account) CreateProductCard(input ProductCard) (*ProductCard, error) {

	if account.Id < 1 {
		return nil, utils.Error{Message: "Не верно указан id аккаунта"}
	}

	input.AccountId = account.Id

	productCard, err := input.create()
	if err != nil {
		return nil, err
	}

	return productCard, nil
}

func (account Account) GetProductCard(cardId uint) (*ProductCard, error) {
	return ProductCard{}.getByAccount(cardId, account.Id)
}

func (account Account) GetProductCards() ([]ProductCard, error) {
	return ProductCard{}.getListByAccount(account.Id)
}

func (account Account) UpdateProductCard(cardId uint, input map[string]interface{}) (*ProductCard, error) {
	productCard, err := account.GetProductCard(cardId)
	if err != nil {
		return nil, err
	}

	if productCard.AccountId != account.Id {
		return nil, utils.Error{Message: "Карточка товара принадлежит другому аккаунту"}
	}

	err = productCard.update(input)
	if err != nil {
		return nil, err
	}

	return productCard, nil
}

func (account Account) DeleteProductCard(cardId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	productCard, err := account.GetProductCard(cardId)
	if err != nil {
		return err
	}

	if err = productCard.delete(); err != nil {
		return err
	}

	return nil
}
// ######### END IF SHOP PRODUCT Functions ############

////// ########

func (productCard ProductCard) AppendProduct(product *Product) error {
	if err := db.Model(&productCard).Association("Products").Append(product).Error; err != nil {
		return err
	}

	account, err := GetAccount(productCard.AccountId)
	if err == nil && account != nil {
		event.AsyncFire(Event{}.ProductCardUpdated(account.Id, productCard.Id))
	}

	return nil
}
