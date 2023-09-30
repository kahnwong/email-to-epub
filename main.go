package main

import (
	"fmt"
	"log"
	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
	// "github.com/emersion/go-imap"
)

func main() {
	// load env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// connect to server
	imap_address := fmt.Sprintf("%s:%s", os.Getenv("IMAP_SERVER_HOST"), os.Getenv("IMAP_SERVER_PORT"))
	c, err := client.DialTLS(imap_address, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout() //nolint:all

	// Login
	if err := c.Login(os.Getenv("IMAP_USERNAME"), os.Getenv("IMAP_PASSWORD")); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}
}
