package models

import (
	"fmt"
	"time"
)

type EmailPkg struct {
	Account       Account
	EmailBox      EmailBox      // отправитель with *Domain
	User          User          // получатель
	EmailTemplate EmailTemplate // шаблон письма
	Subject       string               // тема сообщения
}

var SmtpCh chan EmailPkg

func init() {
	SmtpCh = make(chan EmailPkg, 50)
	go mtaServer(SmtpCh) // start MTA server
}


// внутренняя функция, читающая канал
func mtaServer(c <-chan EmailPkg) {
	for {
		// time.Sleep(time.Second * 2)
		// получаем сообщение
		// pkg, more := <- c
		pkg := <- c
		fmt.Printf("Принял сообщение: %s \n", pkg.Subject)
		fmt.Printf("В очереди: %d\n", len(c))
		fmt.Printf("Макс. длина: %d\n", cap(c))

		// имитируем его отправку
		time.Sleep(time.Second * 2)
	}
}

// Асинхронная функция в одну сторону
func SendEmailPkg(pkg EmailPkg)  {
	// fmt.Println("Отправляем сообщение")
	SmtpCh <- pkg
}
