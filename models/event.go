package models

import (
	"github.com/nkokorev/crm-go/event"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
// При добавлении нового события - внести в список EventItem!!!
type Event struct {}


func (Event) UserCreated(accountId uint, userId uint) event.Event {
	return event.NewBasic("UserCreated", map[string]interface{}{"accountId":accountId, "userId":userId})
}
func (Event) UserUpdated(accountId uint, userId uint) event.Event {
	return event.NewBasic("UserUpdated", map[string]interface{}{"accountId":accountId, "userId":userId})
}
func (Event) UserDeleted(accountId uint, userId uint) event.Event {
	return event.NewBasic("UserDeleted", map[string]interface{}{"accountId":accountId, "userId":userId})
}

func (Event) UserAppendedToAccount(accountId, userId, roleId uint) event.Event {
	return event.NewBasic("UserAppendedToAccount", map[string]interface{}{"accountId":accountId, "userId":userId, "roleId":roleId})
}
func (Event) UserRemovedFromAccount(accountId, userId, roleId uint) event.Event {
	return event.NewBasic("UserRemovedFromAccount", map[string]interface{}{"accountId":accountId, "userId":userId})
}

// ######### Product #########
func (Event) ProductCreated(accountId, productId uint) event.Event {
	return event.NewBasic("ProductCreated", map[string]interface{}{"accountId":accountId, "productId":productId})
}
func (Event) ProductUpdated(accountId, productId uint) event.Event {
	return event.NewBasic("ProductUpdated", map[string]interface{}{"accountId":accountId, "productId":productId})
}
func (Event) ProductDeleted(accountId, productId uint) event.Event {
	return event.NewBasic("ProductDeleted", map[string]interface{}{"accountId":accountId, "productId":productId})
}

// ######### ProductCard #########
func (Event) ProductCardCreated(accountId, productCardId uint) event.Event {
	return event.NewBasic("ProductCardCreated", map[string]interface{}{"accountId":accountId, "productCardId":productCardId})
}
func (Event) ProductCardUpdated(accountId, productCardId uint) event.Event {
	return event.NewBasic("ProductCardUpdated", map[string]interface{}{"accountId":accountId, "productCardId":productCardId})
}
func (Event) ProductCardDeleted(accountId, productCardId uint) event.Event {
	return event.NewBasic("ProductCardDeleted", map[string]interface{}{"accountId":accountId, "productCardId":productCardId})
}

// ######### ProductGroup #########
func (Event) ProductGroupCreated(accountId, productGroupId uint) event.Event {
	return event.NewBasic("ProductGroupCreated", map[string]interface{}{"accountId":accountId, "productGroupId":productGroupId})
}
func (Event) ProductGroupUpdated(accountId, productGroupId uint) event.Event {
	return event.NewBasic("ProductGroupUpdated", map[string]interface{}{"accountId":accountId, "productGroupId":productGroupId})
}
func (Event) ProductGroupDeleted(accountId, productGroupId uint) event.Event {
	return event.NewBasic("ProductGroupDeleted", map[string]interface{}{"accountId":accountId, "productGroupId":productGroupId})
}

// ######### Storage #########
func (Event) StorageCreated(accountId, productId uint) event.Event {
	return event.NewBasic("ProductCreated", map[string]interface{}{"accountId":accountId, "productId":productId})
}
func (Event) StorageUpdated(accountId, productId uint) event.Event {
	return event.NewBasic("ProductUpdated", map[string]interface{}{"accountId":accountId, "productId":productId})
}
func (Event) StorageDeleted(accountId, productId uint) event.Event {
	return event.NewBasic("ProductDeleted", map[string]interface{}{"accountId":accountId, "productId":productId})
}

// ######### Article #########
func (Event) ArticleCreated(accountId, articleId uint) event.Event {
	return event.NewBasic("ArticleCreated", map[string]interface{}{"accountId":accountId, "articleId":articleId})
}
func (Event) ArticleUpdated(accountId, articleId uint) event.Event {
	return event.NewBasic("ArticleUpdated", map[string]interface{}{"accountId":accountId, "articleId":articleId})
}
func (Event) ArticleDeleted(accountId, articleId uint) event.Event {
	return event.NewBasic("ArticleDeleted", map[string]interface{}{"accountId":accountId, "articleId":articleId})
}

// ######### Shop #########
func (Event) ShopCreated(accountId, shopId uint) event.Event {
	return event.NewBasic("ShopCreated", map[string]interface{}{"accountId":accountId, "shopId":shopId})
}
func (Event) ShopUpdated(accountId, shopId uint) event.Event {
	return event.NewBasic("ShopUpdated", map[string]interface{}{"accountId":accountId, "shopId":shopId})
}
func (Event) ShopDeleted(accountId, shopId uint) event.Event {
	return event.NewBasic("ShopDeleted", map[string]interface{}{"accountId":accountId, "shopId":shopId})
}