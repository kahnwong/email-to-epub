package main

import (
	"fmt"
	"log"
	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/joho/godotenv"
)

func loginToIMAPServer() *client.Client {
	imapAddress := fmt.Sprintf("%s:%s", os.Getenv("IMAP_SERVER_HOST"), os.Getenv("IMAP_SERVER_PORT"))
	c, err := client.DialTLS(imapAddress, nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	if err := c.Login(os.Getenv("IMAP_USERNAME"), os.Getenv("IMAP_PASSWORD")); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	return c
}

func isFolderEpubEmpty(c *client.Client) bool {
	mbox, err := c.Select(os.Getenv("DESTINATION_MAILBOX"), false)
	if err != nil {
		log.Fatal(err)
	}

	if mbox.Messages == 0 {
		log.Println("No emails in epub folder")
		return true
	}

	log.Println("EPUB folder not empty")
	return false
}

func getUnreadMessages(c *client.Client, n uint32) *imap.SeqSet {
	mbox, err := c.Select(os.Getenv("SOURCE_MAILBOX"), false)
	if err != nil {
		log.Fatal(err)
	}

	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > (n - 1) {
		// We're using unsigned integers here, only subtract if the result is > 0
		from = mbox.Messages - (n - 1)
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Printf("Last %d messages", n)
	for msg := range messages {
		log.Println("* " + msg.Envelope.Subject)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	return seqset
}

func moveMessagesToDestination(c *client.Client, seqset *imap.SeqSet) {
	log.Println("Moving emails...")
	if err := c.Move(seqset, os.Getenv("DESTINATION_MAILBOX")); err != nil {
		log.Fatalf("Error on move to %s: %v", "epub", err)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	c := loginToIMAPServer()

	isEmptyEmail := isFolderEpubEmpty(c)

	if isEmptyEmail {
		seqset := getUnreadMessages(c, 10)
		moveMessagesToDestination(c, seqset)

		log.Println("Done")
	}
}
