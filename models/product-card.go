package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/nkokorev/crm-go/utils"
)

// Карточка "товара" в магазине в котором могут быть разные торговые предложения
type ProductCard struct {
	ID     				uint `json:"id" gorm:"primary_key"`
	AccountID 			uint `json:"-" gorm:"type:int;index;not null;"` // потребуется, если productGroupId == null
	ShopID 				uint `json:"shopId" gorm:"type:int;index;default:NULL;"` // магазин, к которому относится
	ProductGroupID 		*uint `json:"productGroupId" gorm:"type:int;index;default:NULL;"` // группа товаров, категория товаров

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
	Shop		 		Shop `json:"-"` // к какому магазину относится
	Products 			[]Product `json:"products" gorm:"many2many:product_card_products"` // можно заводить два схожих продукта, разных по цвету
}

func (ProductCard) PgSqlCreate() {
	db.CreateTable(&ProductCard{})
	db.Exec("ALTER TABLE product_cards\n    ADD CONSTRAINT product_cards_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE;\n--     ADD CONSTRAINT product_cards_product_group_id_fkey FOREIGN KEY (product_group_id) REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE;\n")
}

func (productCard *ProductCard) BeforeCreate(scope *gorm.Scope) error {
	productCard.ID = 0
	return nil
}

func (ProductCard) TableName() string {
	return "product_cards"
}

func (productCard ProductCard) getId() uint {
	return productCard.ID
}

// ######### CRUD Functions ############
func (productCard ProductCard) create() (*ProductCard, error) {
	var productCardNew = productCard
	err := db.Create(&productCardNew).First(&productCardNew).Error
	return &productCardNew, err
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

	if err := db.First(&card, "id = ? AND account_id = ?", id, accountId).Error; err != nil {
		return nil, err
	}

	return &card, nil
}

func (ProductCard) getListByShop(shopId uint) ([]ProductCard, error) {

	cards := make([]ProductCard,0)

	err := db.Find(&cards, "shopId = ?", shopId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return cards, nil
	
}

func (ProductCard) getListByAccount(accountId uint) ([]ProductCard, error) {

	cards := make([]ProductCard,0)

	err := db.Find(&cards, "account_id = ?", accountId).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return cards, nil
}

func (productCard *ProductCard) update(input map[string]interface{}) error {
	// fmt.Println(input)
	// return db.Model(card).Omit("id", "account_id").Update(input).Error
	//return db.Model(productCard).Omit("id", "account_id").Updates(structs.Map(input)).Error
	return db.Model(productCard).Omit("id", "account_id").Updates(input).Error

}

func (productCard ProductCard) delete () error {
	return db.Model(ProductCard{}).Where("id = ?", productCard.ID).Delete(productCard).Error
}
// ######### END CRUD Functions ############

// ######### SHOP PRODUCT Functions ############
func (shop Shop) CreateProductCard(input ProductCard, group *ProductGroup) (*ProductCard, error) {

	if shop.ID < 1 {
		return nil, utils.Error{Message: "Не верно указан id магазина"}
	}
	
	input.ShopID = shop.ID
	input.AccountID = shop.AccountID

	if group != nil {
		return group.CreateProductCard(&input);
	} else {
		return input.create()
	}

}

func (shop Shop) GetProductCard(cardId uint) (*ProductCard, error) {
	return ProductCard{}.get(cardId)
}

func (shop Shop) GetProductCardList(offset, limit int, search string, products bool) ([]ProductCard, uint, error) {
	cards := make([]ProductCard,0)

	// if need to search
	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&ProductCard{}).Preload("Products").Limit(limit).Offset(offset).Where("account_id = ?", shop.AccountID).
			Find(&cards, "url ILIKE ? OR breadcrumb ILIKE ? OR meta_title ILIKE ? OR meta_keywords ILIKE ? OR meta_description ILIKE ? OR short_description ILIKE ?" , search,search,search,search,search,search).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		if products {
			err := db.Model(&ProductCard{}).Preload("Products").Limit(limit).Offset(offset).Find(&cards, "account_id = ? AND shop_id = ?", shop.AccountID, shop.ID).Error
			if err != nil && err != gorm.ErrRecordNotFound{
				return nil, 0, err
			}
		} else {
			err := db.Model(&ProductCard{}).Limit(limit).Offset(offset).Find(&cards, "account_id = ? AND shop_id = ?", shop.AccountID, shop.ID).Error
			if err != nil && err != gorm.ErrRecordNotFound{
				return nil, 0, err
			}
		}



	}

	// len(cards) != всему списку!
	var total uint
	err := db.Model(&ProductCard{}).Where("account_id = ? AND shop_id = ?", shop.AccountID, shop.ID).Count(&total).Error
	if err != nil {
		return nil, 0, utils.Error{Message: "Ошибка определения объема"}
	}

	return cards, total, nil
}

func (shop Shop) DeleteProductCard(cardId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	card, err := shop.GetProductCard(cardId)
	if err != nil {
		return err
	}

	return card.delete()
}

// #####################

func (account Account) CreateProductCard(input ProductCard) (*ProductCard, error) {

	if account.ID < 1 {
		return nil, utils.Error{Message: "Не верно указан id аккаунта"}
	}

	input.AccountID = account.ID

	productCard, err := input.create()
	if err != nil {
		return nil, err
	}

	// todo: костыль вместо евента
	go account.CallWebHookIfExist(EventProductCardCreated, productCard)

	return productCard, nil
}

func (account Account) GetProductCard(cardId uint) (*ProductCard, error) {
	return ProductCard{}.getByAccount(cardId, account.ID)
}

func (account Account) GetProductCards() ([]ProductCard, error) {
	return ProductCard{}.getListByAccount(account.ID)
}

func (account Account) UpdateProductCard(cardId uint, input map[string]interface{}) (*ProductCard, error) {
	productCard, err := account.GetProductCard(cardId)
	if err != nil {
		return nil, err
	}

	if productCard.AccountID != account.ID {
		return nil, utils.Error{Message: "Карточка товара принадлежит другому аккаунту"}
	}

	err = productCard.update(input)
	if err != nil {
		return nil, err
	}

	// todo: костыль вместо евента
	go account.CallWebHookIfExist(EventProductCardUpdated, productCard)

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


	go account.CallWebHookIfExist(EventProductCardDeleted, productCard)

	return nil
}
// ######### END IF SHOP PRODUCT Functions ############

////// ########

func (productCard ProductCard) AppendProduct(product *Product) error {
	if err := db.Model(&productCard).Association("Products").Append(product).Error; err != nil {
		return err
	}

	account, err := GetAccount(productCard.AccountID)
	if err == nil && account != nil {
		go account.CallWebHookIfExist(EventProductCardUpdated, productCard)
	}

	return nil
}
