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

var accounts = []models.Account{
	{Name: "RatusMedia",},
	{Name: "Rus-Marketing"},
}

// разворачивает базовые разрешения для всех пользователей
func UserSeeding()  {

	db := base.GetDB()

	db.Unscoped().Delete(&models.Account{})
	db.Unscoped().Delete(&models.User{})

	for i, v := range users {
		if err := v.Create(); err != nil {
			fmt.Println("Cant create Users", err.Error())
			return
		}

		if err := v.CreateAccount(&accounts[i]); err != nil {
			fmt.Println("Cant create Accounts", err.Error())
			return
		}
	}
}
