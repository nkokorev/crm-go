package seeds

import (
	"fmt"
	"github.com/nkokorev/crm-go/database/base"
	"github.com/nkokorev/crm-go/models"
)

var roles = []models.Role {
	{ Name: "Владелец аккаунта",Tag:"owner", 	System: true, Description: "Доступ ко всем данным и функционалу аккаунта."},
	{ Name: "Администратор", 	Tag:"admin", 	System: true, Description: "Доступ ко всем данным и функционалу аккаунта. Не может менять владельца аккаунта."},
	{ Name: "Менеджер", 		Tag:"manager", 	System: true, Description: "Не может добавлять пользователей, не может менять биллинговую информацию."},
	{ Name: "Маркетолог", 		Tag:"marketer", System: true, Description: "Читает все клиентские данные, может изменять все что касается маркетинга, но не заказы или склады."},
	{ Name: "Автор", 			Tag:"author", 	System: true, Description: "Может создавать контент: писать статьи, письма, описания к товарам и т.д."},
	{ Name: "Наблюдатель", 		Tag:"viewer", 	System: true, Description: "The Viewer can view reports in the account"},
	{ Name: "Full Access", 		Tag:"full-access", 	Type:	"api",	System: true, Description: "Полный доступ к аккаунту через API"},
	{ Name: "Site Access", 		Tag:"site-access", 	Type:	"api",	System: true, Description: "Доступ к аккаунте через API, необходимый для интеграции с сайтом"},
	{ Name: "Read Access", 		Tag:"read-access", 	Type:	"api",	System: true, Description: "Доступ к чтению основной информации об аккаунте."},
}

// разворачивает базовые разрешения для всех пользователей
func RoleSeeding()  {

	db := base.GetDB()

	db.Unscoped().Delete(models.Role{})

	for _, v := range roles {
		err := v.Create()
		if err != nil {
			fmt.Println("Cant create Roles")
		}
	}
}


