package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/nkokorev/crm-go/models"
	"github.com/nkokorev/crm-go/routes"
	"github.com/ttacon/libphonenumber"
	"html/template"
	"log"
	"net/http"
	"net/mail"
	"os"
	"os/signal"
	"time"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error load .env file", err)
	}
}

func main() {

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	// Устанавливаем соединение с БД
	pool := models.Connect()

	// закрываем соединение, когда заканчиваем работу
	defer pool.Close()

	// !!! запускаем миграции
	//base.RefreshTables()

	tpl, err := template.ParseFiles("smtp/mail.html")
	if err != nil {
		log.Fatal(err)
		return
	}

	buf := bytes.Buffer{}
	if err := tpl.Execute(&buf, nil); err != nil {
		log.Fatal(err)
		return
	}

	message := models.Message{
		To: mail.Address{Name: "", Address: "nkokorev@rus-marketing.ru"},
		From: mail.Address{Name: "Nikita", Address: "nk@ratuscrm.com"},
		Subject: "Bounce text!",
		Body: buf.String(),
	}
	if err := message.Send(); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println("Msg sent")
	}

	//models.SendTestMail()
	/*
	code, str, err := models.SendEmailNew()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(code)
	fmt.Println(str)
	fmt.Println("Сообщение отослано")
	*/
	//err := models.SendMail("127.0.0.1:443", (&mail.Address{"Nikita", "nk@ratusmedia.com"}).String(), "Test mail", "Сообщение тут", []string{(&mail.Address{"to name", "nkokorev@rus-marketing.ru"}).String()})
	/*if err != nil {
		fmt.Println(err)
	}*/

	//examplePhone("89251952295")
	//examplePhone("+380(44)234-68-88")

	pool.DB().SetConnMaxLifetime(0)
	pool.DB().SetMaxIdleConns(10)
	pool.DB().SetMaxOpenConns(10)

	srv := &http.Server{
		Addr: "127.0.0.1:8090",
		//Addr:         "localhost:8090",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      routes.Handlers(), // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	log.Println("shutting down")
	os.Exit(0)
}

func examplePhone(numToParse string) {

	//num, err := libphonenumber.Get
	num, err := libphonenumber.Parse(numToParse, "RU")
	if err != nil {
		// Handle error appropriately.
		log.Fatal("Err: ", err)
	}
	formattedNum := libphonenumber.Format(num, libphonenumber.NATIONAL)

	//fmt.Println("Num: ", num)
	fmt.Println("CountryCode: ", *num.CountryCode)
	fmt.Println("National Number: ", *num.NationalNumber)
	fmt.Println("National Formatted: ", formattedNum)
	fmt.Println("RFC3966: ", libphonenumber.Format(num, libphonenumber.RFC3966))
	fmt.Println("INTERNATIONAL: ", libphonenumber.Format(num, libphonenumber.INTERNATIONAL)) // наиболее популярный
	fmt.Println("E164: ", libphonenumber.Format(num, libphonenumber.E164))

	// num is a *libphonenumber.PhoneNumber

}
