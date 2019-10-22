package models

import (
	"github.com/nkokorev/crm-go/database/base"
	"reflect"
	"testing"
)

func TestExistUserTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&User{}) {
		tableName := db.NewScope(&User{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

func (user *User) TestCreate(t *testing.T) {

	user.Username = "admin"
	user.Email = "kokorevn@gmail.com"
	user.Name = "Никита"
	user.Surname = "Кокорев"
	user.Patronymic = "Романович"
	user.Password = "qwerty99#DD"

	err := user.Create()
	if err.HasErrors() {
		t.Errorf("Cant create user: %v", err.Message)
	}
}

func TestExistAccountUserTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&AccountUser{}) {
		tableName := db.NewScope(&AccountUser{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}
}

func TestUser_Create(t *testing.T) {

	user := User{
		Username:"admin_test",
		Email: "testmail@ratus-dev.ru",
		//Name:"Nikita",
		Name:"РеальноеИмя",
		//Surname:"Kokorev",
		Surname:"РеальнаяФамилия",
		//Patronymic:"Romanovich",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}

	myErr := user.Create()
	if myErr.HasErrors() {
		t.Error(myErr.Message)
		for k, r := range myErr.GetErrors() {
			t.Errorf("%s | %s", k,r)
		}
		return
	} else {
		defer user.Delete()
	}

	// убеждаемся, что новый пользователь создан
	if user.ID == 0 {
		t.Errorf("User ID == 0, expected > 0")
	}

	temp_user := User{}
	if err := base.GetDB().First(&temp_user, user.ID).Error; err != nil {
		t.Errorf("Cant find created user: %v", err.Error())
	}

	if !reflect.DeepEqual(user, temp_user) {
		t.Error("Данные пользователей созданного и найденого пользователей несовпадают:")
		t.Error(user)
		t.Error(temp_user)
	}

	// тестируем невалидные данные пользователей. Эти пользователи не должны быть созданы
	users := []User{
		{}, // ничего нет %)
		{Username:"", Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // нет username
		{Username:"admin_test", Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // повторяющийся Username
		{Username:"Почтальон!",Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // not ANCII
		{Username:"Почтальон Печкин!",Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // not ANCII
		{Username:"test_user_2", Email: "mail-test@ratus-dev.ru"}, // отсутствует пароль
		{Username:"test_user_3", Email: "mail-test@ratus-dev.ru", Password: ""}, // пустой пароль
		{Username:"test_user_4", Email: "mail-test@ratus-dev.ru", Password: "123"}, // слишком простой пароль
		{Username:"test_user_5", Email: "mail-test@ratus-dev.ru", Password: "dfsjfdsjffdkfdsk"}, // слишком простой пароль
		{Username:"test_user_6", Email: "mail-test@ratus-dev.ru", Password: "dfsjfdsjfd8734udjdjskds"}, // слишком длинный пароль (26)
		{Username:"test_user_7", Email: "info@localhost", Password: "qwerty123#Aa"}, // не валидный email
		{Username:"test_user_8", Email: "localhost", Password: "qwerty123#Aa"}, // не валидный email
		{Username:"test_user_9", Email: "testmail@ratus-dev.ru", Password: "qwerty123#Aa"}, // Этот еmail-адрес уже используется
		{Username:"test_user_10", Password: "qwerty123#Aa"}, // Нет email'a
	}

	for i, _ := range users {
		myErr := users[i].Create()
		if !myErr.HasErrors() {
			t.Errorf("создан пользователь с невалидными данными: %v", users[i].Username)
			defer  users[i].Delete()
		} else {
			/*t.Error(myErr.Message)
			for k, r := range myErr.GetErrors() {
				t.Errorf("%s | %s", k,r)
			}*/
		}
	}

	// проверим ограничения
	test_user_2 := User{
		Username:	"test_user_2",
		Email: 		"mail-test@ratus-dev.ru",
		Name:		"РеальноеИмяРеальноеИмяйва", // 25 simbols
		Surname:	"РеальнаяФамилияРеальнаяФа", // 25 simbols
		Patronymic:	"РеальноеОтчествоРеальноФв", // -- /// --
		Password: 	"qwerty123#Aa",
	}

	myErr = test_user_2.Create()
	if myErr.HasErrors() {
		t.Error(myErr.Message)
		for k, r := range myErr.GetErrors() {
			t.Errorf("%s | %s", k,r)
		}
		return
	} else {
		test_user_2.Delete()
	}

}