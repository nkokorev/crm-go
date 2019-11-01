package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/database/base"
	"testing"
)



func TestExistProductTable(t *testing.T) {
	db := base.GetDB()
	if !db.HasTable(&Product{}) {
		tableName := db.NewScope(&Account{}).GetModelStruct().TableName(db)
		t.Errorf("%v table is not exist!",tableName)
	}

}

func TestProduct_Create(t *testing.T){

	// находим один из существующих аккаунтов
	account := Account{}
	if err := base.GetDB().Model(&Account{}).First(&account,1).Error; err != nil {
		fmt.Println(account)
		t.Error("Cant find account", err)
		return
	}

	// список продуктов для теста - они все должны проходить валидацию
	test_products := []Product{
		{Name:"Дянь хун-1", Account:account},
		{Name:"Дянь мун-2", AccountID: 1},
	}

	// 1. Эти продукты должны создаться
	for i,v := range test_products{

		if err := test_products[i].create(); err != nil {
			t.Errorf("Неудалось создать тестовые продукты: %v %v", v.HashID, err.Error())
		} else {
			defer func(i int) {
				if err := test_products[i].delete(); err != nil {
					t.Errorf("Неудалось удалить тестовый продукт %v %v", v.HashID, err.Error())
				}
			}(i)
		}
	}

	// 2. доп. продукт, тестируем дубль
	test_product_1 := Product{Name: "Product Test", AccountID: 1}
	if err := test_product_1.create(); err != nil {
		t.Errorf("Неудалось создать тестовый продукт: %v %v", test_product_1.HashID, err.Error())
	} else {
		defer func() {
			if err := test_product_1.delete();err!=nil {
				t.Errorf("Неудалось удалить тестовый продукт: %v %v", test_product_1.HashID, err.Error())
			}
		}()
	}

	// 2. Создаем дубль
	if err := test_product_1.create(); err == nil {
		t.Errorf("Удалось создать дубль тестового продукта: %v", test_product_1.HashID)
		if err := test_product_1.delete();err!=nil {
			t.Errorf("Неудалось удалить дубль тестового продукта: %v %v", test_product_1.HashID, err.Error())
		}
	}

	// 3. доп. продукт, тестируем отсуттвие аккаунта
	test_product_2 := Product{Name: "Product Test"}
	if err := test_product_2.create(); err == nil {
		t.Errorf("Создан тестовый продукт без привязки к аккаунта: %v", test_product_2.HashID)
		if err := test_product_2.delete();err!=nil {
			t.Errorf("Неудалось удалить тестовый продукт: %v %v", test_product_2.HashID, err.Error())
		}
	}
	if err := account.CreateProduct(&test_product_2); err != nil {
		t.Errorf("Неудалось создать тестовый продукт: %v", test_product_2.HashID)
	} else {
		if err := test_product_2.delete();err!=nil {
			t.Errorf("Неудалось удалить тестовый продукт: %v %v", test_product_2.HashID, err.Error())
		}
	}
}

func TestProduct_Delete(t *testing.T)  {

	// найдем тестовый аккаунт
	test_account := Account{}
	if err := base.GetDB().Find(&test_account,1).Error;err != nil {
		t.Error("Cant find account", err.Error())
		return
	}

	// продуем с продуктом 1 с аккаунтом с id
	test_product_1 := Product{Name: "Product Test", AccountID: 1}
	if err := test_product_1.create(); err != nil {
		t.Errorf("Неудалось создать тестовый продукт: %v %v", test_product_1.HashID, err.Error())
	} else {
		if err := test_product_1.delete();err!=nil {
			t.Errorf("Неудалось удалить тестовый продукт: %v %v", test_product_1.HashID, err.Error())
		}
	}

	// продуем второй продукт с переданным аккаунтом
	test_product_2 := Product{Name: "Product Test", Account: test_account}
	if err := test_product_2.create(); err != nil {
		t.Errorf("Неудалось создать тестовый продукт: %v %v", test_product_2.HashID, err.Error())
	} else {
		if err := test_product_2.delete();err != nil {
			t.Errorf("Неудалось удалить тестовый продукт: %v %v", test_product_2.HashID, err.Error())
		}
	}


}

// крутая функция: проверяем, что обновляются только нужные поля
func TestProduct_Update(t *testing.T) {

	// находим аккаунт, в котором будем создавать продукт
	test_account := Account{}
	if err := base.GetDB().Model(&Account{}).First(&test_account,1).Error; err != nil {
		t.Error("Cant find account", err)
		return
	}

	// создаем продукт для теста обновления полей
	test_product := Product{Name: "ProductName_1", AccountID: 1}
	if err := test_product.create(); err != nil {
		t.Errorf("Неудалось создать тестовый продукт: %v %v", test_product.HashID, err.Error())
	} else {
		// откладываем удление, т.к. будем использовать тестовый продукт для апдейта
		defer func() {
			if err := test_product.delete();err!=nil {
				t.Errorf("Неудалось удалить тестовый продукт: %v %v", test_product.HashID, err.Error())
			}
		}()
	}

	// 1. Тестируем изменение имени продукта

	// Запоминаем хешпродукта, чтобы потом мы могли найти продукт по его хешу
	hash_id := test_product.HashID

	// Вносим изменения
	test_product.Name = "ProductName_2"
	if err := test_product.update(); err != nil {
		t.Error("Неудалось обновить данные продукта")
		return
	}

	// проверяем изменения
	test_product_2 := Product{}
	if err:= test_product_2.getByHashID(hash_id);err != nil {
		t.Error("Неудалось найти нужный продукт по хешу")
	}

	// проверяем изменилось ли имя продукта
	if test_product_2.Name != test_product.Name {
		t.Errorf("Неудалось внести изменения в данные продукта %v %v", hash_id, test_product.Name)
	}

	// 2. Попробуем внести изменения (hashID) и убеждаемся что их нет
	test_product.Name = "TestHash"
	test_product_2.HashID = "novalid12"
	test_product_2.AccountID = 2
	if err := test_product_2.update(); err != nil {
		t.Error("Ошибка обновления продукта")
		return
	}
	// теперь проверяем, изменился ли хеш или ID
	if test_product.HashID == "novalid12" {
		t.Error("Удалось поменять хеш продукта!")
	}
	if test_product.AccountID == 2 {
		t.Error("Удалось поменять ID продукта!")
	}

	// попробуем другим способом
	test_account_2 := Account{}
	if err := base.GetDB().Model(&Account{}).First(&test_account_2,2).Error; err != nil {
		t.Error("Cant find account 2", err)
		return
	}
	test_product_2.Account = test_account_2
	if err := test_product_2.update(); err != nil {
		t.Error("Ошибка обновления продукта")
		return
	}
	// теперь проверяем, изменился ли хеш или ID
	if test_product.AccountID != 1 {
		t.Error("Удалось поменять ID продукта!")
	}
}
