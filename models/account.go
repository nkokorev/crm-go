package models

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	_ "github.com/nkokorev/auth-server/locales"
	"github.com/nkokorev/crm-go/database/base"
	e "github.com/nkokorev/crm-go/errors"
	t "github.com/nkokorev/crm-go/locales"
	u "github.com/nkokorev/crm-go/utils"
	"log"
	"reflect"
	"time"
)

//a struct to rep user account
type Account struct {
	ID        uint `gorm:"primary_key;unique_index;" json:"-"`
	HashID string `json:"hash_id" gorm:"unique_index"`
	Name string `json:"name"` // СтанПроф / ООО ПК ВТВ-Инжинеринг / Ратус Медия / X6-Band (должно ли быть уникальным??)
	//Label string `json:"label"` // accountId : stan-prof / vtvent / ratus-media / x6-band
	UserID uint `json:"-"` // владелец
	AUsers 		[]AccountUser `json:"-"`
	Users       []User `json:"users" gorm:"many2many:account_users;"`
	ApiKeys		[]ApiKey `json:"-"`
	Roles 		[]Role	`json:"roles"`
	Stores      []Store `json:"stores"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

// Создает новый аккаунт от имени пользователя, добавляя его в список пользователей аккаунта и назначая роль owner.
func (account *Account) create(owner *User) (err error) {

	// проверим, что пользовательский ID есть
	if owner.ID < 1 {
		return e.AccountFailedToCreate
	}

	// если аккаунт существует, то не надо его создавать повторно
	if account.ID > 0 {
		return e.AccountFailedToCreate
	}

	// проверим валидацию данных аккаунта
	myErr := account.ValidateCreate(owner)
	if myErr.HasErrors() {
		return e.AccountFailedToCreate
	}

	account.HashID, err = u.CreateHashID(account)
	if err != nil {
		return err
	}

	// set ID account Owner. Нужна обработка удаления пользователя и его аккаунтов
	account.UserID = owner.ID
	if err = base.GetDB().Create(account).Error; err != nil {
		return err
	}
	if err = base.GetDB().Save(account).Error; err != nil {
		return err
	}

	// Добавляем пользователя в созданный им аккаунт
	aUser, err := account.AppendUser(owner)
	if err != nil {
		// вообще это фаталити, т.к. аккаунт создан, но пользователь не добавился в аккаунт, хотя права доступа к нему он имеет
		// пользователь не увидит созданный аккаунт, в списке доступных аккаунтов, следовательно не сможет зайти или удалить этот аккаунт
		if err = account.Delete(); err != nil {
			return errors.New("Can't append user in his account. And account user not deleted!")
		}
		return errors.New("Can't append user to your account")
	}

	// назначаем роль владельца владельцу аккаунта
	if err = aUser.SetRoleOwner(); err != nil {
		// это фаталити, бесхозный акк, нужно его удалить
		if err = account.Delete(); err != nil {
			return err
		}
		return err
	}

	return
}

func (account *Account) ValidateCreate(owner *User) (error u.Error) {

	// проверяем, что пользователь, который создает аккаунт действительно создан
	if reflect.TypeOf(owner.ID).String() != "uint" || owner.ID < 1 || base.GetDB().First(&User{}, owner.ID).RecordNotFound() {
		error.Message = t.Trans(t.AccountFailedToCreate)
		return
	}

	// проверим заполнение имени аккаунта
	if len([]rune(account.Name)) < 1 {
		error.AddErrors("name", t.Trans(t.InputIsRequired) )
	}

	// проверим заполнение имени аккаунта (не слишком ли оно короткое)
	if len([]rune(account.Name)) < 3 {
		error.AddErrors("name", t.Trans(t.InputIsTooShort) )
	}

	// проверим заполнение имени аккаунта, не слишком ли оно длинное
	if len([]rune(account.Name)) > 25 {
		error.AddErrors("name", t.Trans(t.InputIsTooLong) )
	}

	if error.HasErrors() {
		error.Message = t.Trans(t.AccountFailedToCreate)
		return
	}

	return
}

// Мягкое удаление аккаунта (до 30 дней)
func (account *Account) SoftDelete() error {

	if reflect.TypeOf(account.ID).String() != "uint" {
		return e.AccountDeletionError
	}

	if err := base.GetDB().Delete(&account).Error; err != nil {
		return err
	}
	return nil
}

// Удаление аккаунта и всех связанных данных
func (account *Account) Delete() error {
	if reflect.TypeOf(account.ID).String() != "uint" {
		log.Println("Переданный аккаунт для удаления не содержит ID: ", account)
		return e.AccountDeletionError
	}

	// удаляем вручную все не системные роли
	/*if err := account.RemoveAllRoles(); err != nil {
		return err
	}*/

	if err := base.GetDB().Unscoped().Delete(&account).Error; err != nil {
		return err
	}
	return nil
}

// Добавление к аккаунту пользователя, возвращает aUser пользователя
// todo: а где роли то?
func (account *Account) AppendUser(user *User) (aUser AccountUser, err error) {

	err = errors.New("Неудалось добавить пользователя")
	// проверяем, что пользователь, от имени которого происходит создание аккаунта действительно существует
	if user.ID < 1 {
		//return nil, errors.New("Нельзя добавить пользователя, который не существует")
		return
	}

	// проверим, что аккаунт тоже существует
	if account.ID < 1 {
		//return nil, errors.New("Нельзя добавить пользователя в несуществующимй аккаунт")
		return
	}

	// проверим, что пользователь уже не в аккаунте
	if base.GetDB().Model(user).Where("account_id = ?", account.ID).Association("Accounts").Count() > 0 {
		//return nil, errors.New("Невозможно добавить в аккаунт пользователя, т.к. он уже там")
		return
	}

	// добавляем пользователя в аккаунт путем ассоциации (many-to-many)
	if err = base.GetDB().Model(account).Association("Users").Append(user).Error; err != nil {
		//return nil, err
		return
	}

	// найдем нового пользователя aUser
	if err = aUser.GetAccountUser(user.ID, account.ID); err != nil {
		//return nil, err
		return
	}

	return
}

// исключение пользователя из аккаунта
func (account *Account) RemoveUser(user *User) error {

	if err := base.GetDB().Model(&account).Association("Users").Delete(user).Error; err != nil {
		return err
	}
	return nil
}

// создает индивидуальную для системы роль в аккаунте
func (account *Account) CreateRole(role *Role, codes []int) error {

	// проверим, что аккаунт создан (есть действительный его ID)
	if account.ID < 1 {
		return  errors.New("Cant create role, not found account ID")
	}

	// привязываем к роли аккаунт
	role.AccountID = account.ID

	// указываем, что роль НЕ системная
	role.System = false

	return role.create(codes)
}

// удаляет роль, проверяя ее на системность и права владения аккаунтом
func (account *Account) DeleteRole(role *Role) error {

	// проверим, что аккаунт создан (есть действительный его ID)
	if account.ID < 1 {
		return errors.New("Неудалось удалить роль, т.к. не найден аккаунт")
	}

	if role.ID < 1 {
		fmt.Println(role)
		return errors.New("Нельзя удалить роль, которой не существует")
	}

	// проверим, что роль не системная
	if role.System {
		return errors.New("Нельзя удалить системную роль")
	}

	// проверим, что роль принадлежит текущему аккаунту
	if role.AccountID != account.ID {
		return errors.New("Указанная роль не принадлежит текущему аккаунту")
	}

	if err := role.delete();err != nil {
		return err
	}
	return nil
}

// удаление всех ролей аккаунта, не являющимися системными
func (account *Account) DeleteAllRoles() error {
	roles := []Role{}
	if err := base.GetDB().Delete(&roles,"account_id = ? AND system = FALSE", account.ID).Error; err != nil {
		return err
	}
	return nil
}

// создает API ключ и привязывает его к аккаунту
func (account *Account) CreateApiKey(key *ApiKey, role *Role) error {

	// проверяем, что контекст действителен
	if reflect.TypeOf(account.ID).String() != "uint" {
		return errors.New("Can't create new api-key: account ID not specified")
	}

	// привязывает ключ к аккаунту
	key.AccountID = account.ID

	// назначаем роль
	key.RoleID = role.ID

	// создаем api-key
	if err := key.create(); err != nil {
		return err
	}

	return nil
}

// Удаляет API ключ, принадлежащий аккаунту
func (account *Account) DeleteApiKey(key *ApiKey) error {

	// проверяем, что контекст действителен
	if reflect.TypeOf(account.ID).String() != "uint" {
		return errors.New("Can't create new api-key: account ID not specified")
	}

	// проверяем, что ключ принадлежит этому аккаунту
	if key.AccountID != account.ID {
		return errors.New("Невозможно удалить api-key, т.к. он не принадлежит текущему аккаунту")
	}

	// удаляем api-key
	if err := key.delete(); err != nil {
		return err
	}

	return nil
}

// функция проверяет существенная ли модель
/*func (account *Account) isExists() bool {
	if reflect.TypeOf(account.ID).String() != "uint" || account.ID < 1 || base.GetDB().First(&Account{}, account.ID).RecordNotFound() {
		return false
	}
	return true
}
// обратная к isExists функция
func (account *Account) isNotExists() bool {
	return !account.isExists()
}*/

// ######### ниже не проверенные фукнции ###########
// Создает голую роль в аккаунте с разрешениями (Permission / []Permissions)


// создает новый токен
var CreateAccountToken = func (userId, accountId uint) (cryptToken string, error error) {

	expiresAt := time.Now().Add(time.Minute * 45).Unix()
	claims := Token{
		userId,
		accountId,
		jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Issuer:    "AuthServer",
		},
	}
	cryptToken, error = claims.CreateToken()
	return
}

func GetAccount(u uint) *Account {

	account := &Account{}
	//acc := &Account{}
	db := base.GetDB()
	//db.GetDB().Model(&acc).Association("Accounts")
	//db.GetDB().Preload( "Accounts" ).First (&acc)

	//db.Model(&user).Related("Accounts")

	//db.Preload("Accounts").Where("id = ?", u).First(&user)

	db.Where("id = ?", u).First(account)

	return account
}

func GetAccountByHashID(hash string) *Account {

	account := &Account{}
	//acc := &Account{}
	db := base.GetDB()
	//db.GetDB().Model(&acc).Association("Accounts")
	//db.GetDB().Preload( "Accounts" ).First (&acc)

	//db.Model(&user).Related("Accounts")

	//db.Preload("Accounts").Where("id = ?", u).First(&user)

	db.Where("hash_id = ?", hash).First(account)

	return account
}



