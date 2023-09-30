package main

import (
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"os"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
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

func getUnreadMessages(c *client.Client, mailbox string, n uint32) *imap.SeqSet {
	mbox, err := c.Select(mailbox, false)
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

	messages := make(chan *imap.Message, n)
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

func sanitizeFilename(filename string) string {
	// Remove non-alphanumeric characters, spaces, and replace with underscores
	re := regexp.MustCompile(`[^\w\s-]`)
	filename = re.ReplaceAllString(filename, "_")

	// Replace multiple spaces with a single underscore
	filename = regexp.MustCompile(`\s+`).ReplaceAllString(filename, "_")

	// Trim leading and trailing spaces and underscores
	filename = strings.Trim(filename, "_")

	// Truncate to a reasonable length (e.g., 255 characters)
	if len(filename) > 255 {
		filename = filename[:255]
	}

	return filename
}

func main() {
	// init
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	outputPath := "emails"
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		err := os.Mkdir(outputPath, os.ModeDir|0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	// app
	c := loginToIMAPServer()

	isEmptyEmail := isFolderEpubEmpty(c)

	n := uint32(10) // debug

	if isEmptyEmail {
		seqset := getUnreadMessages(c, os.Getenv("SOURCE_MAILBOX"), n)
		moveMessagesToDestination(c, seqset)

		// write messages to file
		seqsetToFile := getUnreadMessages(c, os.Getenv("DESTINATION_MAILBOX"), n)

		var section imap.BodySectionName
		items := []imap.FetchItem{section.FetchItem()}

		messages := make(chan *imap.Message, n)
		go func() {
			if err := c.Fetch(seqsetToFile, items, messages); err != nil {
				log.Fatal(err)
			}
		}()

		for msg := range messages {
			fmt.Println(msg.Body)

			////////////////////////////////////////////////////////////////
			r := msg.GetBody(&section)

			mr, err := mail.CreateReader(r)
			if err != nil {
				log.Fatal(err)
			}

			body := ""

			header := mr.Header
			subject, err := header.Subject()
			if err != nil {
				log.Fatal(err)
			}

			for {
				p, err := mr.NextPart()
				if err == io.EOF {
					break
				} else if err != nil {
					log.Fatal(err)
				}

				switch p.Header.(type) {
				case *mail.InlineHeader:
					b, _ := io.ReadAll(p.Body)
					body = string(b)
				}
			}

			fmt.Println(body)
			////////////////////////////////////////////////////////////////

			// Create a new file
			file, err := os.Create(fmt.Sprintf("%s/%s.eml", outputPath, sanitizeFilename(subject)))
			if err != nil {
				log.Fatal(err)
			}
			defer file.Close()

			var h mail.Header
			h.SetSubject(subject)

			mw, err := mail.CreateWriter(file, h)
			if err != nil {
				log.Fatal(err)
			}

			// Create a text part
			tw, err := mw.CreateInline()
			if err != nil {
				log.Fatal(err)
			}
			var th mail.InlineHeader
			th.Set("Content-Type", "text/html")
			w, err := tw.CreatePart(th)
			if err != nil {
				log.Fatal(err)
			}

			_, err = io.WriteString(w, body)
			if err != nil {
				log.Fatal(err)
			}

			w.Close()
			tw.Close()

			mw.Close()

			log.Println("Done")
		}
	}
}
