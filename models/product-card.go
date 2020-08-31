package models

import (
	"errors"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"log"
)

// Карточка "товара" в магазине в котором могут быть разные торговые предложения
type ProductCard struct {
	// Id     				uint 	`json:"id" gorm:"primaryKey"`
	Id        			uint 	`json:"id" gorm:"primarykey"`
	AccountId 			uint 	`json:"-" gorm:"type:int;index;not null;"` // потребуется, если productGroupId == null
	WebSiteId 			uint 	`json:"web_site_id" gorm:"type:int;index;"` // магазин, к которому относится
	WebPageId 			uint 	`json:"web_page_id" gorm:"type:int;index;"` // группа товаров, категория товаров

	Enabled 			bool 	`json:"enabled" gorm:"type:bool;default:true"` // активна ли карточка товара
	URL 				string 	`json:"url" gorm:"type:varchar(255);"` // идентификатор страницы (products/syao-chzhun )
	Breadcrumb 			string 	`json:"breadcrumb" gorm:"type:varchar(255);default:null;"`
	Label	 			string 	`json:"label" gorm:"type:varchar(255);default:'';"` // что выводить в список товаров

	MetaTitle 			string 	`json:"meta_title" gorm:"type:varchar(255);default:null;"`
	MetaKeywords 		string 	`json:"meta_keywords" gorm:"type:varchar(255);default:null;"`
	MetaDescription 	string 	`json:"meta_description" gorm:"type:varchar(255);default:null;"`

	// Full description нет т.к. в карточке описание берется от офера
	ShortDescription 	string 	`json:"short_description" gorm:"type:varchar(255);default:null;"` // для превью карточки товара
	Description 		string 	`json:"description" gorm:"type:text;default:null;"` // фулл описание товара

	// Хелперы карточки: переключение по цветам, размерам и т.д.
	// SwitchProducts	 	*pq.StringArray `json:"switch_products" sql:"type:varchar(255)[];default:null"` // {color, size} Параметры переключения среди предложений
	//SwitchProducts	 	pq.StringArray `json:"switch_products" sql:"type:varchar(255)[];default:'{}'"` // {color, size} Параметры переключения среди предложений
	//SwitchProducts	 	[]string `json:"switch_products" gorm:"type:JSONB;DEFAULT '{}'::JSONB"` // {color, size} Параметры переключения среди предложений
	SwitchProducts	 	datatypes.JSON `json:"switch_products"` // {color, size} Параметры переключения среди предложений

	// ProductGroup 		ProductGroup 	`json:"-" gorm:"-"`
	// WebPage 			WebPage 	`json:"-" gorm:"-"`
	WebPages 			[]WebPage 	`json:"web_pages" gorm:"many2many:web_page_product_card;"`
	WebSite		 		WebSite 	`json:"-" gorm:"-"`
	// Products 			[]*Product 		`json:"products" gorm:"many2many:product_card_products;ForeignKey:id;References:id;"`
	Products 			[]Product 	`json:"products" gorm:"many2many:product_card_products;ForeignKey:id;References:id;"`
	// Products 			[]Product 		`json:"products" gorm:"many2many:product_card_products;"`
}

func (ProductCard) PgSqlCreate() {
	if err := db.Migrator().AutoMigrate(&ProductCard{}); err != nil { log.Fatal(err) }
	// db.Model(&ProductCard{}).AddForeignKey("account_id", "accounts(id)", "CASCADE", "CASCADE")
	err := db.Exec("ALTER TABLE product_cards ADD CONSTRAINT product_cards_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;").Error
	if err != nil {
		log.Fatal("Error: ", err)
	}

	err = db.SetupJoinTable(&ProductCard{}, "Products", &ProductCardProduct{})
	if err != nil {
		log.Fatal(err)
	}
}

func (productCard *ProductCard) BeforeCreate(tx *gorm.DB) error {
	productCard.Id = 0
	return nil
}


func (productCard ProductCard) getId() uint {
	return productCard.Id
}

// ######### CRUD Functions ############
func (productCard ProductCard) create() (*ProductCard, error) {
	var productCardNew = productCard
	if err := db.Create(&productCardNew).Error; err != nil {
		return nil, err
	}
	err := db.Where("id = ?", productCardNew.Id).Find(&productCardNew).Error
	if err != nil {
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
/*func (webSite WebSite) CreateProductCard(input ProductCard, group *ProductGroup) (*ProductCard, error) {

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

}*/

func (webSite WebSite) GetProductCard(cardId uint) (*ProductCard, error) {
	return ProductCard{}.get(cardId)
}

func (webSite WebSite) GetProductCardList(offset, limit int, search string, products bool) ([]ProductCard, int64, error) {
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
	var total int64
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

	if product.Id > 0 {
		if err := db.Model(&ProductCardProduct{}).Create(&ProductCardProduct{ProductId: product.Id, ProductCardId: productCard.Id}).Error; err != nil {
			return err
		}
	} else {
		// if err := db.Model(&productCard).Association("Products").Append(product); err != nil {
		// if err := db.Debug().Model(product).Association("ProductCards").Replace(&productCard); err != nil {
		if err := db.Debug().Set("gorm:saveAssociations", false).Model(&productCard).Association("Products").Append(product); err != nil {
			return err
		}
	}

	account, err := GetAccount(productCard.AccountId)
	if err == nil && account != nil {
		event.AsyncFire(Event{}.ProductCardUpdated(account.Id, productCard.Id))
	}

	return nil
}

type ProductCardProduct struct {
	ProductId  uint
	ProductCardId uint
}
func (ProductCardProduct) BeforeCreate(db *gorm.DB) error {
	// ...
	return nil
}

func (productCard ProductCard) ExistProduct(product *Product) bool {
	if product.Id < 1 {
		return false
	}
	var count int64
	db.Model(&ProductCardProduct{}).Where("product_id = ? AND product_card_id = ?", product.Id, productCard.Id).Count(&count)
	if count > 0 {
	 return true
	}
	return false
}
