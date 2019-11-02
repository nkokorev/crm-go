package seeds

import (
	"fmt"
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

	//db := base.GetDB()

	//db.DropTableIfExists("account_users")
	//db.Unscoped().Delete(&models.Account{})
	//db.Unscoped().Delete(&models.User{})

	for i, _ := range users {
		if err := users[i].Create(); err != nil {
			fmt.Println("Cant create User: ", users[i], err.Error())
			return
		} else {
			//fmt.Println("User created: ", users[i])
		}

		if err := users[i].CreateAccount(&accounts[i]); err != nil {
			fmt.Println("Cant create Accounts: ", accounts[i], "\r\n", err.Error())
			return
		}
	}

	// добавим первого пользователя во второй аккаунт
	fmt.Println("Добавим пользователя во второй аккаунт")

	//aUser := &models.AccountUser{}
	aUser, err := accounts[1].AppendUser(&users[0]);
	if err != nil {
		fmt.Println("Cant add user to Accounts, хз почему")
	}

	if err := aUser.SetRoleAdmin(); err != nil {
		fmt.Println("Неудалось добавить роль...")
	}

}


