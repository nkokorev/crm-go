package models

import (
	"github.com/nkokorev/crm-go/database/base"
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

	test_user_1 := User{
		Username:"admin_test",
		Email: "testmail@ratus-dev.ru",
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}

	myErr := test_user_1.Create()
	if myErr.HasErrors() {
		t.Error(myErr.Message)
		for k, r := range myErr.GetErrors() {
			t.Errorf("%s | %s", k,r)
		}
		return
	} else {
		defer test_user_1.Delete()
	}

	// убеждаемся, что в нового пользователя загружены новые данные после создания
	if test_user_1.ID == 0 {
		t.Errorf("User ID == 0, expected > 0")
		return
	}
	temp_user := User{}
	if base.GetDB().First(&temp_user, test_user_1.ID).RecordNotFound() {
		t.Errorf("Cant find created user: %v", test_user_1.Username)
		return
	}

	// тестируем невалидные данные пользователей. Эти пользователи не должны быть созданы
	test_users := []User{
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

	for i, _ := range test_users {
		myErr := test_users[i].Create()
		if !myErr.HasErrors() {
			t.Errorf("создан пользователь с невалидными данными: %v", test_users[i].Username)
			defer  test_users[i].Delete()
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
		defer test_user_2.Delete()
	}

}

func TestUser_Delete(t *testing.T) {
	test_user := User{Username:"test_user", Email:"testmail@ratus-dev.ru", Password:"qwerty1#R"}
	e := test_user.Create()
	if e.HasErrors() {
		if e.HasErrors() {
			t.Error(e.Message)
			for k, r := range e.GetErrors() {
				t.Errorf("%s | %s", k,r)
			}
			return
		}
	} else {
		defer test_user.Delete()
	}

	// убеждаемся, что новый пользователь создан
	if test_user.ID == 0 {
		t.Errorf("User ID == 0, expected > 0")
		return
	}

	userID := test_user.ID

	e = test_user.Delete()
	if e.HasErrors() {
		if e.HasErrors() {
			t.Error(e.Message)
			for k, r := range e.GetErrors() {
				t.Errorf("%s | %s", k,r)
			}
			return
		}
	}

	temp_user := User{}
	base.GetDB().First(&temp_user,userID)

	if temp_user.ID == userID {
		t.Errorf("найден пользователь, который должен быть удален ID: %v", userID)
	}

}

func TestUser_SoftDelete(t *testing.T) {
	test_user := User{Username:"test_user", Email:"testmail@ratus-dev.ru", Password:"qwerty1#R"}
	e := test_user.Create()
	if e.HasErrors() {
		if e.HasErrors() {
			t.Error(e.Message)
			for k, r := range e.GetErrors() {
				t.Errorf("%s | %s", k,r)
			}
			return
		}
	} else {
		defer test_user.Delete()
	}

	// убеждаемся, что новый пользователь создан
	if test_user.ID == 0 {
		t.Errorf("User ID == 0, expected > 0")
		return
	}

	userID := test_user.ID

	e = test_user.SoftDelete()
	if e.HasErrors() {
		if e.HasErrors() {
			t.Error(e.Message)
			for k, r := range e.GetErrors() {
				t.Errorf("%s | %s", k,r)
			}
			return
		}
	}

	temp_user := User{}

	// без специального условия мы его не должны найти
	base.GetDB().First(&temp_user,userID)
	if temp_user.ID == userID {
		t.Errorf("найден пользователь, который должен быть удален ID: %v", userID)
	}

	// теперь мы его должны найти т.к. ищем все, без учета deleted_at
	base.GetDB().Unscoped().First(&temp_user,userID)
	if temp_user.ID != userID {
		t.Errorf("не найден пользователь, который должен быть мягко удален ID: %v", userID)
	}
}