package main

import (
	"fmt"
	"log"
	"os"

	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
	// "github.com/emersion/go-imap"
)

func loginToIMAPServer() (*client.Client, error) {
	imapAddress := fmt.Sprintf("%s:%s", os.Getenv("IMAP_SERVER_HOST"), os.Getenv("IMAP_SERVER_PORT"))
	c, err := client.DialTLS(imapAddress, nil)
	if err != nil {
		return nil, err
	}
	log.Println("Connected")

	if err := c.Login(os.Getenv("IMAP_USERNAME"), os.Getenv("IMAP_PASSWORD")); err != nil {
		return nil, err
	}
	log.Println("Logged in")

	return c, nil
}

func main() {
	// load env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// login to imap server
	c, err := loginToIMAPServer()
	if err != nil {
		log.Fatal(err)
	}

	// Select folder: epub
	mbox, err := c.Select("epub", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for INBOX:", mbox.Flags)

	// Check if empty
	fmt.Println(mbox.Messages)

	if mbox.Messages == 0 {
		log.Println("No emails in epub folder")
	}
}
