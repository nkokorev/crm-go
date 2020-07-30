package models

import (
	"github.com/nkokorev/crm-go/event"
)

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
// При добавлении нового события - внести в список EventItem!!!
type Event struct {}

// ######### User #########
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
func (Event) UserRemovedFromAccount(accountId, userId uint) event.Event {
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
func (Event) StorageCreated(accountId, storageId uint) event.Event {
	return event.NewBasic("StorageCreated", map[string]interface{}{"accountId":accountId, "storageId":storageId})
}
func (Event) StorageUpdated(accountId, storageId uint) event.Event {
	return event.NewBasic("StorageUpdated", map[string]interface{}{"accountId":accountId, "storageId":storageId})
}
func (Event) StorageDeleted(accountId, storageId uint) event.Event {
	return event.NewBasic("StorageDeleted", map[string]interface{}{"accountId":accountId, "storageId":storageId})
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

// ######### WebSite #########
func (Event) WebSiteCreated(accountId, webSiteId uint) event.Event {
	return event.NewBasic("WebSiteCreated", map[string]interface{}{"accountId":accountId, "webSiteId":webSiteId})
}
func (Event) WebSiteUpdated(accountId, webSiteId uint) event.Event {
	return event.NewBasic("WebSiteUpdated", map[string]interface{}{"accountId":accountId, "webSiteId":webSiteId})
}
func (Event) WebSiteDeleted(accountId, webSiteId uint) event.Event {
	return event.NewBasic("WebSiteDeleted", map[string]interface{}{"accountId":accountId, "webSiteId":webSiteId})
}


// ######### Email Marketing #########
func (Event) UserSubscribed(accountId uint, userId uint) event.Event {
	return event.NewBasic("UserSubscribed", map[string]interface{}{"accountId":accountId, "userId":userId})
}
func (Event) UserUnsubscribed(accountId uint, userId uint) event.Event {
	return event.NewBasic("UserUnsubscribed", map[string]interface{}{"accountId":accountId, "userId":userId})
}
func (Event) UserUpdateSubscribeStatus(accountId uint, userId uint) event.Event {
	return event.NewBasic("UserUpdateSubscribeStatus", map[string]interface{}{"accountId":accountId, "userId":userId})
}

// ######### Order #########
func (Event) OrderCreated(accountId uint, orderId uint) event.Event {
	return event.NewBasic("OrderCreated", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}
func (Event) OrderUpdated(accountId uint, orderId uint) event.Event {
	return event.NewBasic("OrderUpdated", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}
func (Event) OrderDeleted(accountId uint, orderId uint) event.Event {
	return event.NewBasic("OrderDeleted", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}
func (Event) OrderCompleted(accountId uint, orderId uint) event.Event {
	return event.NewBasic("OrderCompleted", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}
func (Event) OrderCanceled(accountId uint, orderId uint) event.Event {
	return event.NewBasic("OrderCanceled", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}

// ######### DeliveryOrder #########
func (Event) DeliveryOrderCreated(accountId uint, deliveryOrderId uint) event.Event {
	return event.NewBasic("DeliveryOrderCreated", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderUpdated(accountId uint, deliveryOrderId uint) event.Event {
	return event.NewBasic("DeliveryOrderUpdated", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderDeleted(accountId uint, deliveryOrderId uint) event.Event {
	return event.NewBasic("DeliveryOrderDeleted", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderCompleted(accountId uint, deliveryOrderId uint) event.Event {
	return event.NewBasic("DeliveryOrderCompleted", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderCanceled(accountId uint, deliveryOrderId uint) event.Event {
	return event.NewBasic("DeliveryOrderCanceled", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}

// ######### Payment #########
func (Event) PaymentCreated(accountId uint, paymentId uint) event.Event {
	return event.NewBasic("PaymentCreated", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}
func (Event) PaymentUpdated(accountId uint, paymentId uint) event.Event {
	return event.NewBasic("PaymentUpdated", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}
func (Event) PaymentDeleted(accountId uint, paymentId uint) event.Event {
	return event.NewBasic("PaymentDeleted", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}
func (Event) PaymentCompleted(accountId uint, paymentId uint) event.Event {
	return event.NewBasic("PaymentCompleted", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}
func (Event) PaymentCanceled(accountId uint, paymentId uint) event.Event {
	return event.NewBasic("PaymentCanceled", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}

