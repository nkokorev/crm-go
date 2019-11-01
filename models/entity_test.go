package models

import (
	"testing"
)

// test create and delete entity models
func TestAccount_CreateAndDeleteEntity(t *testing.T) {
	// для теста новых моделей нужно ее добавить сюда.
	var es = []Entity{&Product{Name:"TestProduct"}, &Role {Name:"TestRole", Description:"This is new role! (test)"}}
	var test_account Account

	// Найдем аккаунт для тестов
	if err := test_account.GetByID(1); err != nil {
		t.Error("Неудалось найти аккаунт с ID: 1")
		return
	}

	// создаем циклами и циклами удаляем
	for i, v := range es {
		if err := test_account.CreateEntity(es[i]); err != nil {
			t.Errorf("Неудалось создать продукт!\r\n" +
				"Account: %v \r\n" +
				"Entity: %v,\r\n" +
				"Error: %v", test_account, es[i], err)
			return
		} else {

			// убеждаемся, что нельзя создать дубль
			if err := test_account.CreateEntity(es[i]); err == nil {
				t.Errorf("Удалось создать дубль продукта!\r\n" +
					"Account: %v \r\n" +
					"Entity: %v,\r\n", test_account, es[i])

				if err := test_account.DeleteEntity(es[i]); err != nil {
					t.Error("Неудалось удалить модель: ", v)
				}
			}

			// убеждаемся, что нельзя удалить из-под другого аккаунта
			es[i].setAccountID(2)
			if err := test_account.DeleteEntity(es[i]); err == nil {
				t.Error("Удалось удалить модель из-под чужого аккаунта: ", v)
			} else {

				// теперь удаляем по-настояему
				es[i].setAccountID(1)
				if err := test_account.DeleteEntity(es[i]); err != nil {
					t.Error("Неудалось удалить модель: ", v)
				}

			}
		}
	}
}

// Тестируем не все возможные модели: Product
func TestAccount_GetEntity(t *testing.T) {
	var test_product, temp_product Product
	var test_role, temp_role Role
	var test_account Account

	// Найдем аккаунт для тестов
	if err := test_account.GetByID(1); err != nil {
		t.Error("Неудалось найти аккаунт с ID: 1")
		return
	}

	// создадим тестовый продукт
	test_product.Name = "Test account"
	if err := test_account.CreateProduct(&test_product);err != nil {
		t.Error("Неудалось создать продукт в аккаунте: ", test_product, err)
		return
	} else {
		defer func() {
			if err := test_product.delete(); err != nil {
				t.Error("Неудалось удалить модель: ", test_product)
			}
		}()
	}

	// попробуем найти тестовый продукт по хешу
	if err := test_account.GetEntity(test_product.HashID, &temp_product); err != nil {
		t.Error("Неудалось получить модель из аккаунта: ", err.Error())
		return
	}

	// сверяем хеши
	if test_product.HashID != temp_product.HashID {
		t.Error("Хеши продуктов не совпали, поиск Entity не работает")
		return
	}

	// # Test Role # //

	// добавим в аккаунт тестовую роль
	if err := test_account.CreateRole(&test_role, []int{}); err != nil {
		t.Error("Неудалось добавить роль в аккаунт")
		return
	} else {
		defer func() {
			if err := test_role.delete(); err != nil {
				t.Error("Неудалось удалить роль: ", test_role)
			}
		}()
	}

	// пробуем найти тестовую роль по хешу
	if err := test_account.GetEntity(test_role.HashID, &temp_role); err != nil {
		t.Error("Неудалось получить модель из аккаунта: ", err.Error())
		return
	}

	// сверяем хеши
	if test_role.HashID != temp_role.HashID {
		t.Error("Хеши продуктов не совпали, поиск Entity не работает")
		return
	}

}