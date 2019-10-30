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

	for i, v := range users {
		if err := users[i].Create(); err != nil {
			fmt.Println("Cant create Users", err.Error())
			return
		} else {
			fmt.Println("User created: ", v)
		}

		if err := users[i].CreateAccount(&accounts[i]); err != nil {
			fmt.Println("Cant create Accounts", err.Error())
			return
		}
	}

	// добавим первого пользователя во второй аккаунт
	fmt.Println("Добавим пользователя во второй аккаунт")
/*	fmt.Println(accounts[0])
	fmt.Println(accounts[1])
	fmt.Println(users[0])
	fmt.Println(users[1])
*/

	//aUser := &models.AccountUser{}
	aUser, err := accounts[1].AppendUser(&users[0]);
	if err != nil || aUser == nil {
		fmt.Println("Cant add user to Accounts", err.Error())
	}
	if err := aUser.SetRoleAdmin(); err != nil {
		fmt.Println("")
	}
}
