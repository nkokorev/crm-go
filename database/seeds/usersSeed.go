package seeds

import (
	"fmt"
	"github.com/nkokorev/crm-go/database/base"
	"github.com/nkokorev/crm-go/models"
)

var users = []models.User{
	// Пользователи: 2хх |
	{Username:"admin",Email: "kokorevn@gmail.com",Name:"Никита",Surname:"Кокорев",Patronymic:"Романович",Password: "qwerty99#DD"},
	{Username:"mex388",Email: "mex388@gmail.com",Name:"Никита",Surname:"Кокорев",Patronymic:"Романович",Password: "qwerty99#DD"},


}

// разворачивает базовые разрешения для всех пользователей
func UserSeeding()  {

	db := base.GetDB()

	db.Delete(models.User{})

	for _, v := range users {
		err := v.Create()
		if err.HasErrors() {
			fmt.Println("Cant create Users")
		}
	}
}
