package models

import (
	u "github.com/nkokorev/crm-go/utils"
	"golang.org/x/crypto/bcrypt"
	//"gopkg.in/guregu/null.v3"
	"time"
)

type User struct {
	ID        	uint `json:"id"`
	//HashID 		string `json:"hash_id" gorm:"type:varchar(10);unique_index;not null;"`
	Username 	string `json:"username" `
	Email 		string `json:"email"`
	Password 	string `json:"password"` // json:"-"

	Name 		string `json:"name"`
	Surname 	string `json:"surname"`
	Patronymic 	string `json:"patronymic"`


	DefaultAccountID int `json:"default_account_id"` // указывает какой аккаунт по дефолту загружать

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}


// ### CRUD FUNC ###

// Создает нового пользователя с новым ID
func (u *User) Create () error {

	// проверим входящие сообщения
	if err := u.ValidateCreate(); err != nil {
		return err
	}
	// Создаем пароль
	password, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(password)

	return db.Create(u).Error
}

// осуществляет поиск по ID
func (u *User) Get () error {
	return db.First(u.ID).Error
}

// сохраняет все поля в модели, кроме id, deleted_at
func (u *User) Save () error {
	return db.Omit("id", "deleted_at").Save(u).Error
}

// обновляет все схожие с интерфейсом поля, кроме id, deleted_at
func (u *User) Update (i interface{}) error {
	return db.Model(u).Where("id = ?", u.ID).Omit("id", "deleted_at").Update(i).Find(u, "id = ?", u.ID).Error
}

// удаляет пользователя по ID
func (u *User) Delete () error {
	return db.Model(u).Where("id = ?", u.ID).Delete(u).Error
}

// ### HELPERS FUNC ###

func (u *User) Exist() bool {
	return db.First(u, "ip = ?", u.ID).RecordNotFound()
}


// Проверка входящих полей
func (user *User) ValidateCreate() error {

	e := u.Error{"Не верно указанные данные", map[interface{}]interface{}{} }
	e.AddErrors("email", "Данный email уже используется")
	e.AddErrors("vals", 12)


	if true {
		return e
	}

	/*// проверка на попытку создать дубль пользователя, который уже был создан
	if reflect.TypeOf(user.ID).String() == "uint" {
		if user.ID > 0 && !base.GetDB().First(&User{}, user.ID).RecordNotFound() {
			error.Message = t.Trans(t.UserCreateInvalidCredentials)
		}
	}*/

	// проверка почты должна быть полной, но при тестах часто используется сокращенная проверка
	/*err := u.VerifyEmail(user.Email, !(os.Getenv("http_dev") == "true"))
	if err != nil {
		error.AddErrors("email", err.Error())
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	// проверка на уникальность email-адреса
	if !base.GetDB().First(&User{}, "email = ?", user.Email).RecordNotFound() {
		error.AddErrors("email", "этот адрес уже используется" )
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	err = u.VerifyPassword(user.Password)
	if err != nil {
		error.AddErrors("email", err.Error())
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	// проверим корректность заполнения имени пользователя
	err = u.VerifyUsername(user.Username)
	if err != nil {
		error.AddErrors("username", err.Error() )
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	// проверка на уникальность username
	if !base.GetDB().First(&User{}, "username = ?", user.Username).RecordNotFound() {
		error.AddErrors("username", "этот username уже используется" )
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	if len([]rune(user.Name)) > 25 {
		error.AddErrors("name", t.Trans(t.InputIsTooLong) )
	}
	if len([]rune(user.Surname)) > 25 {
		error.AddErrors("surname", t.Trans(t.InputIsTooLong) )
	}
	if len([]rune(user.Patronymic)) > 25 {
		error.AddErrors("patronymic", t.Trans(t.InputIsTooLong) )
	}

	if error.HasErrors() {
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
		return
	}

	//Email must be unique
	temp := &User{}

	//check for errors and duplicate emails
	err = base.GetDB().Unscoped().Model(&User{}).Where("email = ?", user.Email).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		error.Message = t.Trans(t.UserFailedToCreate)
		return
	}
	if temp.Email != "" {
		error.AddErrors("email", t.Trans(t.EmailAlreadyUse) )
	}

	temp = &User{} // set to empty

	err = base.GetDB().Unscoped().Model(&User{}).Where("username = ?", user.Username).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		error.Message = t.Trans(t.UserFailedToCreate)
		return
	}
	if temp.Username != "" {
		error.AddErrors("username", t.Trans(t.UsernameAlreadyUse) )
	}

	if error.HasErrors() {
		error.Message = t.Trans(t.UserCreateInvalidCredentials)
	}*/

	//return u.Error{"email", pairs }
	return nil
}
