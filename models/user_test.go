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

func TestUser_Create(t *testing.T) {

	test_user_owner := User{
		Username:"test_user_owner", // не менять!!
		Email: "testmail@ratus-dev.ru", // не менять!
		Name:"РеальноеИмя",
		Surname:"РеальнаяФамилия",
		Patronymic:"РеальноеОтчество",
		Password: "qwerty123#Aa",
	}
	err := test_user_owner.Create()
	if err != nil {
		t.Error("Неудалось создать пользователя: ", err.Error())
	} else {
		defer func() {
			if err := test_user_owner.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	// 1. убеждаемся, что в нового пользователя загружены новые данные после создания
	if test_user_owner.ID == 0 {
		t.Errorf("User ID == 0, expected > 0")
	}
	temp_user := User{}
	if base.GetDB().First(&temp_user, test_user_owner.ID).RecordNotFound() {
		t.Errorf("Cant find created user: %v", test_user_owner.Username)
	}

	// 2. проверим, что нельзя повторно тут же создать дубль пользователя с тем же ID
	test_user_owner.Username = "Test User Double" // меняем имя пользователя
	test_user_owner.Email = "mail-test@ratus-dev.ru" // меняем почту
	if err := test_user_owner.Create(); err == nil {
		t.Error("Удалось повторно создать пользователя, который был уже создан")
	} else {
		defer func() {
			if err := test_user_owner.Delete(); err != nil {
				t.Error("неудалось удалить пользователя: ", err.Error())
			}
		}()
	}

	// 3. тестируем валидатор встроенный в функцию (валидатор еще тестируется отдельной функцией)
	test_users := []User{
		{}, // ничего нет %)
		{Username:"test_user_owner", Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // повторяющийся Username
		{Username:"test_user_9", Email: "testmail@ratus-dev.ru", Password: "qwerty123#Aa"}, // Этот еmail-адрес уже используется
		{Username:"", Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // нет username
		{Username:"Почтальон!",Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // not ANCII
		{Username:"test_user_10", Password: "qwerty123#Aa"}, // Нет email'a
	}
	for i, _ := range test_users {
		err = test_users[i].Create()
		if err == nil {
			t.Errorf("создан пользователь с невалидными данными: %v", test_users[i].Username)
			defer  test_users[i].Delete()
		}
	}

	// проверим ограничения, что пользователь создаются с длинными данными
	test_user_2 := User{
		Username:	"test_user_2",
		Email: 		"mail-test@ratus-dev.ru",
		Name:		"РеальноеИмяРеальноеИмяйва", // 25 simbols
		Surname:	"РеальнаяФамилияРеальнаяФа", // 25 simbols
		Patronymic:	"РеальноеОтчествоРеальноФв", // -- /// --
		Password: 	"qwerty123#Aa",
	}

	err = test_user_2.Create()
	if err != nil {
		t.Error(err.Error())
	} else {
		defer test_user_2.Delete()
	}

}

func TestUser_Delete(t *testing.T) {
	test_user := User{Username:"test_user", Email:"testmail@ratus-dev.ru", Password:"qwerty1#R"}
	err := test_user.Create()
	if err != nil {
		t.Error(err.Error())
	} else {
		defer test_user.Delete()
	}

	// убеждаемся, что новый пользователь создан
	if test_user.ID == 0 {
		t.Errorf("User ID == 0, expected > 0")
	}

	// проверяем удаление пользователя
	if err = test_user.Delete(); err != nil {
		t.Error(err.Error())
	}

	// проверям, что пользователь удалился
	if !base.GetDB().First(&User{}, "hash_id = ?",test_user.HashID).RecordNotFound() {
		t.Errorf("найден пользователь, который должен быть удален hash_id: %v", test_user.HashID)
	}

}

func TestUser_SoftDelete(t *testing.T) {
	test_user := User{Username:"test_user", Email:"testmail@ratus-dev.ru", Password:"qwerty1#R"}
	err := test_user.Create()
	if err != nil {
		t.Error(err.Error())
	} else {
		defer test_user.Delete()
	}

	// убеждаемся, что новый пользователь создан
	if test_user.ID == 0 {
		t.Errorf("User ID == 0, expected > 0")
	}

	userID := test_user.ID

	err = test_user.SoftDelete()
	if err != nil {
		t.Error(err.Error())
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

func TestUser_ValidateCreate(t *testing.T) {

	// тестируем невалидные данные пользователей. Эти пользователи не должны быть созданы
	test_users := []User{
		{}, // ничего нет %)
		{Username:"", Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // нет username
		{Username:"Почтальон!",Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // not ANCII
		{Username:"Почтальон Печкин!",Email: "mail-test@ratus-dev.ru", Password: "qwerty123#Aa"}, // not ANCII
		{Username:"test_user_2", Email: "mail-test@ratus-dev.ru"}, // отсутствует пароль
		{Username:"test_user_3", Email: "mail-test@ratus-dev.ru", Password: ""}, // пустой пароль
		{Username:"test_user_4", Email: "mail-test@ratus-dev.ru", Password: "123"}, // слишком простой пароль
		{Username:"test_user_5", Email: "mail-test@ratus-dev.ru", Password: "dfsjfdsjffdkfdsk"}, // слишком простой пароль
		{Username:"test_user_6", Email: "mail-test@ratus-dev.ru", Password: "dfsjfdsjfd8734udjdjskds"}, // слишком длинный пароль (26)
		{Username:"test_user_7", Email: "info@localhost", Password: "qwerty123#Aa"}, // не валидный email
		{Username:"test_user_8", Email: "localhost", Password: "qwerty123#Aa"}, // не валидный email
		{Username:"test_user_9", Password: "qwerty123#Aa"}, // Нет email'a
	}

	for i, _ := range test_users {
		myErr := test_users[i].ValidateCreate()
		if !myErr.HasErrors() {
			t.Errorf("создан пользователь с невалидными данными: %v", test_users[i].Username)
			defer  test_users[i].Delete()
		}
	}
}

func TestUser_Update(t *testing.T) {
	// todo: дописать тест
}