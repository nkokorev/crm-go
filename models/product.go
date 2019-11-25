package models

import (
	"fmt"
	"github.com/nkokorev/crm-go/database/base"
	"github.com/nkokorev/crm-go/utils"
	"log"
	"strings"
	"time"
)

type Product struct {
	ID uint	`json:"id"`
	AccountID uint `json:"-"`
	SKU string `json:"sku"`
	Name string `json:"name"`
}

func (p *Product) create() error {
	return nil
}

func (p *Product) save() error {
	return nil
}

func (p *Product) delete() error {
	return nil
}

func (p *Product) get() error {
	return nil
}



func (p *Product) Create() error {

	//utils.TimeTrack(time.Now(), "Create product: ")

	cols := "(sku, name)"
	sqlStr := "insert into products " + cols + " values (?,?) "

	pool := base.GetPool()
	res, err := pool.Exec(sqlStr, p.SKU, p.Name)
	if err != nil {
		fmt.Println(err)
	}

	id, _ := res.LastInsertId()
	//fmt.Println("ID: ", id)

	row := pool.QueryRow("select * from products where id = ?", id)
	if err := row.Scan(&p.ID, &p.AccountID, &p.SKU, &p.Name); err != nil {
		return err
	}


	//fmt.Println(id)

	//pool.QueryRow()

	/*v := reflect.TypeOf(*p)

	// проверка типа
	for i := 0; i < v.NumField(); i++ {
		fmt.Println(v.Field(i).Name)
	}*/


	/*val := reflect.Indirect(reflect.ValueOf(p))

	for i:=0; i < val.Type().NumField(); i++ {
		fmt.Println(val.Type().Field(i).Name)
	}

	val2 := reflect.ValueOf(p).Elem()
	for i:=0; i<val2.NumField();i++{
		fmt.Println(val2.Type().Field(i).Name)
	}*/
	//fmt.Println(val.Type().Field(0).Name)

	return nil
}

// Создает кучу продуктов
func CreateProducts(count int)  {

	utils.TimeTrack(time.Now(), "Create products: ")
	pool := base.GetPool()

	// очистим таблицу продуктов
	_, err := pool.Exec("delete from products;")
	if err != nil {
		log.Fatal(err)
	}

	// clear eav_product_attributes
	_, err = pool.Exec("delete from eav_product_attributes;")
	if err != nil {
		log.Fatal(err)
	}



	sqlStr := ""

	for i:= 0; i < count; i++ {
		sqlStr += fmt.Sprintf("('sku-%v','name-%v'),", i, i)
	}
	sqlStr = strings.TrimSuffix(sqlStr, ",")

	_, err = pool.Exec("insert into products (sku, name) values " + sqlStr)
	if err != nil {
		log.Fatal(err)
	}

	// каждому продукту по каждому атрибуту в подарок
	_, err = pool.Exec("INSERT INTO eav_product_attributes (product_id, eav_attributes_id)\nSELECT DISTINCT products.id, eav_attributes.id FROM products, eav_attributes;")
	if err != nil {
		log.Fatal(err)
	}

	// каждому продукту
}