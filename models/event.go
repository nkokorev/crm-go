package models

import (
	"github.com/nkokorev/crm-go/event"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
// При добавлении нового события - внести в список EventItem!!!
type Event struct {}


func (Event) UserCreated(accountID uint, userID uint) event.Event {
	return event.NewBasic("UserCreated", map[string]interface{}{"accountID":accountID, "userID":userID})
}
func (Event) UserUpdated(accountID uint, userID uint) event.Event {
	return event.NewBasic("UserUpdated", map[string]interface{}{"accountID":accountID, "userID":userID})
}
func (Event) UserDeleted(accountID uint, userID uint) event.Event {
	return event.NewBasic("UserDeleted", map[string]interface{}{"accountID":accountID, "userID":userID})
}

func (Event) UserAppendedToAccount(accountID, userID, roleID uint) event.Event {
	return event.NewBasic("UserAppendedToAccount", map[string]interface{}{"accountID":accountID, "userID":userID, "roleID":roleID})
}
func (Event) UserRemovedFromAccount(accountID, userID uint) event.Event {
	return event.NewBasic("UserRemovedFromAccount", map[string]interface{}{"accountID":accountID, "userID":userID})
}

// ######### Product #########
func (Event) ProductCreated(accountID, productID uint) event.Event {
	return event.NewBasic("ProductCreated", map[string]interface{}{"accountID":accountID, "productID":productID})
}
func (Event) ProductUpdated(accountID, productID uint) event.Event {
	return event.NewBasic("ProductUpdated", map[string]interface{}{"accountID":accountID, "productID":productID})
}
func (Event) ProductDeleted(accountID, productID uint) event.Event {
	return event.NewBasic("ProductDeleted", map[string]interface{}{"accountID":accountID, "productID":productID})
}

// ######### ProductCard #########
func (Event) ProductCardCreated(accountID, productCardID uint) event.Event {
	return event.NewBasic("ProductCardCreated", map[string]interface{}{"accountID":accountID, "productCardID":productCardID})
}
func (Event) ProductCardUpdated(accountID, productCardID uint) event.Event {
	return event.NewBasic("ProductCardUpdated", map[string]interface{}{"accountID":accountID, "productCardID":productCardID})
}
func (Event) ProductCardDeleted(accountID, productCardID uint) event.Event {
	return event.NewBasic("ProductCardDeleted", map[string]interface{}{"accountID":accountID, "productCardID":productCardID})
}

// ######### ProductGroup #########
func (Event) ProductGroupCreated(accountID, productGroupID uint) event.Event {
	return event.NewBasic("ProductGroupCreated", map[string]interface{}{"accountID":accountID, "productGroupID":productGroupID})
}
func (Event) ProductGroupUpdated(accountID, productGroupID uint) event.Event {
	return event.NewBasic("ProductGroupUpdated", map[string]interface{}{"accountID":accountID, "productGroupID":productGroupID})
}
func (Event) ProductGroupDeleted(accountID, productGroupID uint) event.Event {
	return event.NewBasic("ProductGroupDeleted", map[string]interface{}{"accountID":accountID, "productGroupID":productGroupID})
}

// ######### Storage #########
func (Event) StorageCreated(accountID, productID uint) event.Event {
	return event.NewBasic("ProductCreated", map[string]interface{}{"accountID":accountID, "productID":productID})
}
func (Event) StorageUpdated(accountID, productID uint) event.Event {
	return event.NewBasic("ProductUpdated", map[string]interface{}{"accountID":accountID, "productID":productID})
}
func (Event) StorageDeleted(accountID, productID uint) event.Event {
	return event.NewBasic("ProductDeleted", map[string]interface{}{"accountID":accountID, "productID":productID})
}

// ######### Article #########
func (Event) ArticleCreated(accountID, articleID uint) event.Event {
	return event.NewBasic("ArticleCreated", map[string]interface{}{"accountID":accountID, "articleID":articleID})
}
func (Event) ArticleUpdated(accountID, articleID uint) event.Event {
	return event.NewBasic("ArticleUpdated", map[string]interface{}{"accountID":accountID, "articleID":articleID})
}
func (Event) ArticleDeleted(accountID, articleID uint) event.Event {
	return event.NewBasic("ArticleDeleted", map[string]interface{}{"accountID":accountID, "articleID":articleID})
}

// ######### WebSite #########
func (Event) WebSiteCreated(accountID, webSiteID uint) event.Event {
	return event.NewBasic("WebSiteCreated", map[string]interface{}{"accountID":accountID, "webSiteID":webSiteID})
}
func (Event) WebSiteUpdated(accountID, webSiteID uint) event.Event {
	return event.NewBasic("WebSiteUpdated", map[string]interface{}{"accountID":accountID, "webSiteID":webSiteID})
}
func (Event) WebSiteDeleted(accountID, webSiteID uint) event.Event {
	return event.NewBasic("WebSiteDeleted", map[string]interface{}{"accountID":accountID, "webSiteID":webSiteID})
}