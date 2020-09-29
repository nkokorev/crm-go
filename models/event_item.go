package models

// Список всех событий, которы могут быть вызваны // !!! Не добавлять другие функции под этим интерфейсом !!!
// При добавлении нового события - внести в список EventItem!!!
// type Event struct {}
/*
// ######### User #########
func (Event) UserCreated(accountId uint, userId uint) *Event {
	return NewEvent("UserCreated", map[string]interface{}{"accountId":accountId, "userId":userId})
}
func (Event) UserUpdated(accountId uint, userId uint) *Event {
	return NewEvent("UserUpdated", map[string]interface{}{"accountId":accountId, "userId":userId})
}
func (Event) UserDeleted(accountId uint, userId uint) *Event {
	return NewEvent("UserDeleted", map[string]interface{}{"accountId":accountId, "userId":userId})
}

func (Event) UserAppendedToAccount(accountId, userId, roleId uint) *Event {
	return NewEvent("UserAppendedToAccount", map[string]interface{}{"accountId":accountId, "userId":userId, "roleId":roleId})
}
func (Event) UserRemovedFromAccount(accountId, userId uint) *Event {
	return NewEvent("UserRemovedFromAccount", map[string]interface{}{"accountId":accountId, "userId":userId})
}

// ######### Product #########
func (Event) ProductCreated(accountId, productId uint) *Event {
	return NewEvent("ProductCreated", map[string]interface{}{"accountId":accountId, "productId":productId})
}
func (Event) ProductUpdated(accountId, productId uint) *Event {
	return NewEvent("ProductUpdated", map[string]interface{}{"accountId":accountId, "productId":productId})
}
func (Event) ProductDeleted(accountId, productId uint) *Event {
	return NewEvent("ProductDeleted", map[string]interface{}{"accountId":accountId, "productId":productId})
}

// ######### ProductCard #########
func (Event) ProductCardCreated(accountId, productCardId uint) *Event {
	return NewEvent("ProductCardCreated", map[string]interface{}{"accountId":accountId, "productCardId":productCardId})
}
func (Event) ProductCardUpdated(accountId, productCardId uint) *Event {
	return NewEvent("ProductCardUpdated", map[string]interface{}{"accountId":accountId, "productCardId":productCardId})
}
func (Event) ProductCardDeleted(accountId, productCardId uint) *Event {
	return NewEvent("ProductCardDeleted", map[string]interface{}{"accountId":accountId, "productCardId":productCardId})
}

// ######### ProductCategory #########
func (Event) ProductCategoryCreated(accountId, productCategoryId uint) *Event {
	return NewEvent("ProductCategoryCreated", map[string]interface{}{"accountId":accountId, "productCategoryId":productCategoryId})
}
func (Event) ProductCategoryUpdated(accountId, productCategoryId uint) *Event {
	return NewEvent("ProductCategoryUpdated", map[string]interface{}{"accountId":accountId, "productCategoryId":productCategoryId})
}
func (Event) ProductCategoryDeleted(accountId, productCategoryId uint) *Event {
	return NewEvent("ProductCategoryDeleted", map[string]interface{}{"accountId":accountId, "productCategoryId":productCategoryId})
}

// ######### Storage #########
func (Event) StorageCreated(accountId, storageId uint) *Event {
	return NewEvent("StorageCreated", map[string]interface{}{"accountId":accountId, "storageId":storageId})
}
func (Event) StorageUpdated(accountId, storageId uint) *Event {
	return NewEvent("StorageUpdated", map[string]interface{}{"accountId":accountId, "storageId":storageId})
}
func (Event) StorageDeleted(accountId, storageId uint) *Event {
	return NewEvent("StorageDeleted", map[string]interface{}{"accountId":accountId, "storageId":storageId})
}

// ######### Article #########
func (Event) ArticleCreated(accountId, articleId uint) *Event {
	return NewEvent("ArticleCreated", map[string]interface{}{"accountId":accountId, "articleId":articleId})
}
func (Event) ArticleUpdated(accountId, articleId uint) *Event {
	return NewEvent("ArticleUpdated", map[string]interface{}{"accountId":accountId, "articleId":articleId})
}
func (Event) ArticleDeleted(accountId, articleId uint) *Event {
	return NewEvent("ArticleDeleted", map[string]interface{}{"accountId":accountId, "articleId":articleId})
}

// ######### WebSite #########
func (Event) WebSiteCreated(accountId, webSiteId uint) *Event {
	return NewEvent("WebSiteCreated", map[string]interface{}{"accountId":accountId, "webSiteId":webSiteId})
}
func (Event) WebSiteUpdated(accountId, webSiteId uint) *Event {
	return NewEvent("WebSiteUpdated", map[string]interface{}{"accountId":accountId, "webSiteId":webSiteId})
}
func (Event) WebSiteDeleted(accountId, webSiteId uint) *Event {
	return NewEvent("WebSiteDeleted", map[string]interface{}{"accountId":accountId, "webSiteId":webSiteId})
}

// ######### WebPage #########
func (Event) WebPageCreated(accountId, webPageId uint) *Event {
	return NewEvent("WebPageCreated", map[string]interface{}{"accountId":accountId, "webPageId":webPageId})
}
func (Event) WebPageUpdated(accountId, webPageId uint) *Event {
	return NewEvent("WebPageUpdated", map[string]interface{}{"accountId":accountId, "webPageId":webPageId})
}
func (Event) WebPageDeleted(accountId, webPageId uint) *Event {
	return NewEvent("WebPageDeleted", map[string]interface{}{"accountId":accountId, "webPageId":webPageId})
}

// ######### WarehouseItem #########
func (Event) WarehouseItemProductAppended(accountId, warehouseId, productId uint) *Event {
	return NewEvent("ProductCardCreated", map[string]interface{}{"accountId":accountId, "warehouseId":warehouseId, "productId":productId})
}

// ######### InventoryItem #########
func (Event) InventoryItemProductAppended(accountId, inventoryId, productId uint) *Event {
	return NewEvent("InventoryItemProductAppended", map[string]interface{}{"accountId":accountId, "inventoryId":inventoryId, "productId":productId})
}
func (Event) InventoryItemProductRemoved(accountId, inventoryId, productId uint) *Event {
	return NewEvent("InventoryItemProductRemoved", map[string]interface{}{"accountId":accountId, "inventoryId":inventoryId, "productId":productId})
}


// ######### Email Marketing #########
func (Event) UserSubscribed(accountId uint, userId uint) *Event {
	return NewEvent("UserSubscribed", map[string]interface{}{"accountId":accountId, "userId":userId})
}
func (Event) UserUnsubscribed(accountId uint, userId uint) *Event {
	return NewEvent("UserUnsubscribed", map[string]interface{}{"accountId":accountId, "userId":userId})
}
func (Event) UserUpdateSubscribeStatus(accountId uint, userId uint) *Event {
	return NewEvent("UserUpdateSubscribeStatus", map[string]interface{}{"accountId":accountId, "userId":userId})
}

// ######### Order #########
func (Event) OrderCreated(accountId uint, orderId uint) *Event {
	return NewEvent("OrderCreated", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}
func (Event) OrderUpdated(accountId uint, orderId uint) *Event {
	return NewEvent("OrderUpdated", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}
func (Event) OrderDeleted(accountId uint, orderId uint) *Event {
	return NewEvent("OrderDeleted", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}
func (Event) OrderCompleted(accountId uint, orderId uint) *Event {
	return NewEvent("OrderCompleted", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}
func (Event) OrderCanceled(accountId uint, orderId uint) *Event {
	return NewEvent("OrderCanceled", map[string]interface{}{"accountId":accountId, "orderId":orderId})
}

// ######### DeliveryOrder #########
func (Event) DeliveryOrderCreated(accountId uint, deliveryOrderId uint) *Event {
	return NewEvent("DeliveryOrderCreated", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderUpdated(accountId uint, deliveryOrderId uint) *Event {
	return NewEvent("DeliveryOrderUpdated", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderAlimented(accountId uint, deliveryOrderId uint) *Event {
	return NewEvent("DeliveryOrderAlimented", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderInProcess(accountId uint, deliveryOrderId uint) *Event {
	return NewEvent("DeliveryOrderInProcess", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderCompleted(accountId uint, deliveryOrderId uint) *Event {
	return NewEvent("DeliveryOrderCompleted", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderCanceled(accountId uint, deliveryOrderId uint) *Event {
	return NewEvent("DeliveryOrderCanceled", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderStatusUpdated(accountId uint, deliveryOrderId uint) *Event {
	return NewEvent("DeliveryOrderStatusUpdated", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}
func (Event) DeliveryOrderDeleted(accountId uint, deliveryOrderId uint) *Event {
	return NewEvent("DeliveryOrderDeleted", map[string]interface{}{"accountId":accountId, "deliveryOrderId":deliveryOrderId})
}

// ######### Payment #########
func (Event) PaymentCreated(accountId uint, paymentId uint) *Event {
	return NewEvent("PaymentCreated", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}
func (Event) PaymentUpdated(accountId uint, paymentId uint) *Event {
	return NewEvent("PaymentUpdated", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}
func (Event) PaymentDeleted(accountId uint, paymentId uint) *Event {
	return NewEvent("PaymentDeleted", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}
func (Event) PaymentCompleted(accountId uint, paymentId uint) *Event {
	return NewEvent("PaymentCompleted", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}
func (Event) PaymentCanceled(accountId uint, paymentId uint) *Event {
	return NewEvent("PaymentCanceled", map[string]interface{}{"accountId":accountId, "paymentId":paymentId})
}

// ######### Manufacturer #########
func (Event) ManufacturerCreated(accountId, manufacturerId uint) *Event {
	return NewEvent("ManufacturerCreated", map[string]interface{}{"accountId":accountId, "manufacturerId":manufacturerId})
}
func (Event) ManufacturerUpdated(accountId, manufacturerId uint) *Event {
	return NewEvent("ManufacturerUpdated", map[string]interface{}{"accountId":accountId, "manufacturerId":manufacturerId})
}
func (Event) ManufacturerDeleted(accountId, manufacturerId uint) *Event {
	return NewEvent("ManufacturerDeleted", map[string]interface{}{"accountId":accountId, "manufacturerId":manufacturerId})
}

*/