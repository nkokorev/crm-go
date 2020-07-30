package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"github.com/nkokorev/crm-go/event"
	"github.com/nkokorev/crm-go/utils"
)

type ProductGroup struct {
	Id     uint   `json:"id" gorm:"primary_key"`
	WebSiteId uint `json:"webSiteId" gorm:"type:int;index;not null;"` // магазин, к которому относится данная группа
	// AccountId uint `json:"-" gorm:"type:int;index;not null;"` // хз хз
	// ParentId uint `json:"parentId,omitempty" gorm:"default:NULL"`
	ParentId *uint `json:"parentId" gorm:"default:NULL"`

	Code string `json:"code" gorm:"type:varchar(255);default:null;"` // tea, coffe, china
	URL string `json:"url"gorm:"type:varchar(255);default:null;"`

	Name string `json:"name" gorm:"type:varchar(255);default:null;"` // Чай, кофе, ..
	IconName string `json:"iconName" gorm:"type:varchar(255);default:null;"` // icon name
	RouteName string `json:"routeName" gorm:"type:varchar(255);default:null;"` // Чай, кофе, ..

	Order 	int		`json:"order" gorm:"type:int;default:10;"` // Порядок отображения (часто нужно файлам)
	Breadcrumb string `json:"breadcrumb" gorm:"type:varchar(255);default:null;"`
	ShortDescription string `json:"shortDescription" gorm:"type:varchar(255);default:null;"`
	Description string `json:"description" gorm:"type:text;default:null;"`

	MetaTitle string `json:"metaTitle" gorm:"type:varchar(255);default:null;"`
	MetaKeywords string `json:"metaKeywords" gorm:"type:varchar(255);default:null;"`
	MetaDescription string `json:"metaDescription" gorm:"type:varchar(255);default:null;"`

	WebSite WebSite `json:"webSite" `
	ParentGroup *ProductGroup `json:"-"` // if has parentId
	// ProductCards ProductCard `json:"productCards" gorm:"many2many:product_group_product_cards"`
	ProductCards []ProductCard `json:"productCards"`
	// Products []Product `json:"products" gorm:"many2many:product_group_products"`

}

func (ProductGroup) PgSqlCreate() {
	// 1. Создаем таблицу и настройки в pgSql
	db.CreateTable(&ProductGroup{})
	db.Model(&ProductGroup{}).AddForeignKey("web_site_id", "web_sites(id)", "CASCADE", "CASCADE")

	db.Exec("ALTER TABLE product_groups\n--     ADD CONSTRAINT products_account_id_fkey FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE ON UPDATE CASCADE,\n    alter column parent_id SET DEFAULT NULL;\n--     ADD CONSTRAINT products_parent_id_fkey FOREIGN KEY (parent_id) REFERENCES product_groups(id) ON DELETE CASCADE ON UPDATE CASCADE;\n\n\n-- create unique index uix_products_account_id_sku ON products (account_id,sku);\n-- alter table product_groups alter column parent_id set default NULL;\n")
}

func (productGroup *ProductGroup) BeforeCreate(scope *gorm.Scope) error {
	productGroup.Id = 0
	return nil
}

func (productGroup ProductGroup) getId() uint {
	return productGroup.Id
}

// ######### CRUD Functions ############
func (productGroup ProductGroup) create() (*ProductGroup, error)  {
	
	if productGroup.WebSiteId < 1 {
		return nil, utils.Error{Message: "Необходимо указать Id магазина"}
	}

	var productGroupNew = productGroup
	// if err := db.Create(&productGroupNew).Preload("WebSite").Find(productGroupNew).Error; err != nil {
	if err := db.Create(&productGroupNew).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&productGroupNew).Preload("WebSite").First(&productGroupNew).Error; err != nil {
		return nil, err
	}

	event.AsyncFire(Event{}.ProductGroupCreated(productGroupNew.WebSite.AccountId, productGroupNew.Id))
	
	return &productGroupNew, nil
}

func (ProductGroup) get(id uint) (*ProductGroup, error) {

	group := ProductGroup{}

	if err := db.First(&group, id).Error; err != nil {
		return nil, err
	}

	return &group, nil
}

func (productGroup *ProductGroup) update(input map[string]interface{}) error {
	if err :=  db.Model(productGroup).Omit("id").Updates(input).Preload("WebSite").Find(productGroup).Error; err != nil { return err }

	event.AsyncFire(Event{}.ProductGroupUpdated(productGroup.WebSite.AccountId, productGroup.Id))

	return nil
}

func (productGroup *ProductGroup) delete () error {

	accountId, err2 := GetAccountIdByWebSiteId(productGroup.WebSiteId)

	if err := db.Model(ProductGroup{}).Where("id = ?", productGroup.Id).Delete(productGroup).Error; err != nil { return err }

	if err2 != nil {
		event.AsyncFire(Event{}.ProductGroupDeleted(accountId, productGroup.Id))
	}

	return nil
}
// ######### END CRUD Functions ############


// ######### SHOP Functions ############
func (webSite WebSite) CreateProductGroup(input ProductGroup) (*ProductGroup, error) {
	input.WebSiteId = webSite.Id

	productGroup, err := input.create()
	if err != nil {
		return nil, err
	}

	/*account, err := GetAccount(webSite.AccountId)
	if err == nil && account != nil {
		event.AsyncFire(Event{}.ProductGroupCreated(account.Id, productGroup.Id))
	}*/

	return productGroup, nil
}

func (webSite WebSite) GetProductGroup(groupId uint) (*ProductGroup, error) {
	group, err := ProductGroup{}.get(groupId)
	if err != nil {
		return nil, err
	}

	return group, nil
}

func (webSite WebSite) GetProductGroupList() ([]ProductGroup, error) {

	groups := make([]ProductGroup,0)

	err := db.Model(&webSite).Where("web_site_id = ?", webSite.Id).Association("ProductGroups").Find(&groups).Error
	if err != nil && err != gorm.ErrRecordNotFound{
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	return groups, nil
}
func (webSite WebSite) GetProductGroupsPaginationList(offset, limit int, search string) ([]ProductGroup, uint, error) {

	productGroups := make([]ProductGroup,0)
	//groups := []ProductGroup{}

	if len(search) > 0 {

		// string pattern
		search = "%"+search+"%"

		err := db.Model(&webSite).
			Limit(limit).
			Offset(offset).
			Where("web_site_id = ?", webSite.Id).
			Where("code ILIKE ? OR url ILIKE ? OR name ILIKE ? OR short_description ILIKE ? OR description ILIKE ? OR meta_title ILIKE ? OR meta_keywords ILIKE ? OR meta_description ILIKE ?" , search,search,search,search,search,search,search,search).
			Association("ProductGroups").
			Find(&productGroups).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}

	} else {
		if offset < 0 || limit < 0 {
			return nil, 0, errors.New("Offset or limit is wrong")
		}

		err := db.Model(&webSite).
			Limit(limit).
			Offset(offset).
			Where("web_site_id = ?", webSite.Id).
			Association("ProductGroups").
			Find(&productGroups).Error
		if err != nil && err != gorm.ErrRecordNotFound{
			return nil, 0, err
		}
	}

	total := db.Model(&webSite).Where("web_site_id = ?", webSite.Id).Association("ProductGroups").Count()

	/*if err := db.Model(&webSite).Association("ProductGroups").Find(&groups).Error; err != nil {
		return nil, err
	}*/

	return productGroups, uint(total), nil
}

func (account Account) GetProductGroups() ([]ProductGroup, error) {

	groups := make([]ProductGroup,0)

	err := db.Model(&ProductGroup{}).
		Joins("LEFT JOIN web_sites ON product_groups.web_site_id = web_sites.id").
		Where("account_id = ?", account.Id).
		Find(&groups).Error;
	if err != nil {
		return nil, err
	}

	return groups, nil
}

/*func (account Account) GetProductGroupByRouteName(routeName uint) (*ProductGroup, error) {
	group, err := ProductGroup{}.get(groupId)
	if err != nil {
		return nil, err
	}

	return group, nil
}*/

func (webSite WebSite) UpdateProductGroup(groupId uint, input map[string]interface{}) (*ProductGroup, error) {
	productGroup, err := webSite.GetProductGroup(groupId)
	if err != nil {
		return nil, err
	}

	// Проверим, что новый WebSite принадлежит этому аккаунту
	/*m := structs.Map(input)

	// todo: 
	if m["WebSiteId"] != webSite.Id {
	   acc, err := GetAccount(webSite.AccountId)
	   if err != nil {
		   return nil, utils.Error{Message: "Не удалось получить аккаунт"}
	   }
		if !acc.ExistShop(group.WebSiteId) {
			return nil, utils.Error{Message: "Указанный магазин принадлежит другому аккаунту"}
		}
	}

	// если меняется родитель
	if m["ParentId"] != group.ParentId {
		// todo: 
	}*/

	err = productGroup.update(input)
	if err != nil {
		return nil, err
	}

	/*account, err := GetAccount(webSite.AccountId)
	if err == nil && account != nil {
		go account.CallWebHookIfExist(EventProductGroupUpdated, productGroup)
	}*/

	return productGroup, nil
}

func (webSite WebSite) DeleteProductGroup(groupId uint) error {

	// включает в себя проверку принадлежности к аккаунту
	productGroup, err := webSite.GetProductGroup(groupId)
	if err != nil {
		return err
	}

	if err = productGroup.delete(); err != nil {
		return err
	}

	return nil
}

// ######### END OF SHOP Functions ############


// ######### ProductGroup Functions ############
func (productGroup ProductGroup) CreateChild(input ProductGroup) (*ProductGroup, error) {
	input.ParentId = &productGroup.Id
	input.WebSiteId = productGroup.WebSiteId
	return input.create()
}

// Создает и добавляет продукт в категорию товаров
func (productGroup ProductGroup) CreateProductCard(_c *ProductCard) (*ProductCard, error) {
 	_c.ProductGroupId = &productGroup.Id
	return _c.create()
}
/*func (group ProductGroup) CreateAndAppendProductCard(card *ProductCard) error {
	return db.Model(&group).Association("ProductCards").Append(card).Error
}*/

func (productGroup ProductGroup) GetProductCards() ([]ProductCard, error) {

	cards := make([]ProductCard,0)

	if err := db.Model(&productGroup).Association("ProductCards").Find(&cards).Error; err != nil {
		return nil, err
	}

	return cards, nil
}

// ######### END OF ProductGroup Functions ############




