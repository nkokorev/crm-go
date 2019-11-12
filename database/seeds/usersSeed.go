package seeds

import (
	"fmt"
	"github.com/nkokorev/crm-go/models"
)

var users = []models.User{
	// Пользователи: 2хх |
	{Username:"admin",Email: "kokorevn@gmail.com",Name:"Никита",Surname:"Кокорев",Patronymic:"Романович",Password: "qwerty99#DD"},
	{Username:"mex388",Email: "mex388@gmail.com",Name:"Никита",Surname:"Кокорев",Patronymic:"Романович",Password: "qwerty99#DD"},
	{Username:"vasiliy",Email: "vasiliy1985@gmail.com",Name:"Василий",Surname:"Попов",Patronymic:"Николаевич",Password: "qwerty99#DD"},
}

var accounts = []models.Account{
	{Name: "RatusMedia",},
	{Name: "Rus-Marketing"},
	{Name: "357 грамм"},
}

var keys = []models.ApiKey{
	{Name: "Integration with site"},
	{Name: "Integration with site"},
	{Name: "Integration with site"},
}

var products = []models.Entity {
	&models.Product{Name:"Дянь Хун Мао Фэн"},
	&models.Product{Name:"Сяо Чжун"},
	&models.Product{Name:"Дянь Хун Цзинь Хао"},
	&models.Product{Name:"Хун Сун Чжень"},
	&models.Product{Name:"Цзинь Цзюнь Мэй"},
	&models.Product{Name:"Чжун Го Хун"},
	&models.Product{Name:"Ли Чжи Хун Ча"},
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

		// добавляем ключи API с ролями полного доступа
		if err := accounts[i].CreateEntity(&keys[i]); err != nil {
			fmt.Println("Key cant be created:", err, keys[i])
			return
		}
		if err := keys[i].SetRoleFullAccess(); err != nil {
			fmt.Println("Cant set Role Full Access", err)
			return
		}

	}

	// Добавим первого пользователя во второй аккаунт
	//aUser := &models.AccountUser{}
	aUser, _ := accounts[1].AppendUser(&users[0]);
	if err := aUser.SetRoleAdmin(); err != nil {
		fmt.Println("Неудалось добавить роль...")
	}

	aUser2, _ := accounts[2].AppendUser(&users[0]);
	if err := aUser2.SetRoleAdmin(); err != nil {
		fmt.Println("Неудалось добавить роль...")
	}


	// Добавляем продукты в 357 гр
	if err := accounts[2].CreateEntities(products); err != nil {
		fmt.Println("Неудалось добавить продукты в 357 гр...")
	}



}


