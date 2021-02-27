package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// JSONMessage type of message received from channel
type JSONMessage struct {
	ID          uuid.UUID `json:"id"`
	URL         string    `json:"url"`
	Price       string    `json:"price"`
	Title       string    `json:"title"`
	Beds        int       `json:"beds"`
	Baths       int       `json:"baths"`
	Provider    string    `json:"provider"`
	Eircode     string    `json:"eircode"`
	DateRenewed time.Time `json:"date_renewed"`
	FirstListed time.Time `json:"first_listed"`
	Propertyid  string    `json:"property_id"`
	Photo       string    `json:"photo"`

}

func main(){
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv(("DB_HOST"))

	db := pg.Connect(&pg.Options{
		Addr:      dbHost,
		User:      dbUser,
		Password:  dbPass,
		Database:  dbName,
		TLSConfig: &tls.Config{InsecureSkipVerify: true},
	})

	defer db.Close()
	
	fmt.Println("Listing changes...");
	ln := db.Listen(db.Context(),"houseinserted")
	defer ln.Close()

	ch := ln.Channel()
	for val := range ch {
		processMessage(val.Payload)
	}
}

func processMessage(message string) {
	log.Println("received message:", message)
	sendEmail(message)
}

func sendEmail(message string ){
	jMessage:= JSONMessage{}
	err := json.Unmarshal([]byte(message), &jMessage)
	if err != nil {
		log.Println("Error converting message")
	}
	
	from := os.Getenv("EMAIL_FROM")
	pass := os.Getenv("EMAIL_PASS")
	to := os.Getenv("EMAIL_TO")

	msg := "From: " + from + "\n" +
		"To: " + to + "\n" +
		"Subject: New house detected\n\n" +
		"Title: " + jMessage.Title +"\n\n" +
		"URL: " + jMessage.URL +"\n\n"


	err = smtp.SendMail("smtp.gmail.com:587",
		smtp.PlainAuth("", from, pass, "smtp.gmail.com"),
		from, []string{to}, []byte(msg))

	if err != nil {
		log.Printf("smtp error: %s", err)
	}
}
