package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/google/uuid"
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

//func isFolderEPUBEmpty(c *client.Client) bool {
//	mbox, err := c.Select(os.Getenv("DESTINATION_MAILBOX"), false)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	if mbox.Messages == 0 {
//		log.Println("No emails in epub folder")
//		return true
//	}
//
//	log.Println("EPUB folder not empty")
//	return false
//}

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

func getMessagesBody(c *client.Client, n uint32, seqset *imap.SeqSet) (imap.BodySectionName, chan *imap.Message) {
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, n)
	go func() {
		if err := c.Fetch(seqset, items, messages); err != nil {
			log.Fatal(err)
		}
	}()

	return section, messages
}

func getMessageContent(section imap.BodySectionName, msg *imap.Message) (string, string) {
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
	log.Printf("Fetching: %s", subject)

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

	return subject, body
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
		filename = filename[:65]
	}

	return filename
}

func writeEMLFile(outputPath string, subject string, body string) {
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
}

func runCommand(command string) {
	cmd := exec.Command("/bin/sh", "-c", command)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		panic(err)
	}

	// Create scanners to read from stdout and stderr
	stdoutScanner := bufio.NewScanner(stdout)
	stderrScanner := bufio.NewScanner(stderr)

	// Start goroutines to read from stdout and stderr
	go func() {
		for stdoutScanner.Scan() {
			fmt.Println(stdoutScanner.Text())
		}
	}()

	go func() {
		for stderrScanner.Scan() {
			fmt.Println(stderrScanner.Text())
		}
	}()

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}

func main() {
	// init
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Loading env from env var instead...")
	}

	outputPath := "emails"
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		err := os.Mkdir(outputPath, os.ModeDir|0755)
		if err != nil {
			log.Fatal(err)
		}
	}

	outputFile := fmt.Sprintf("output-%s.epub", uuid.New())

	// app
	c := loginToIMAPServer()
	//isEmptyEmail := isFolderEPUBEmpty(c)

	n_raw, err := strconv.Atoi(os.Getenv("N"))
	if err != nil {
		log.Println("Error:", err)
		return
	}
	n := uint32(n_raw)

	//if isEmptyEmail {
	seqset := getUnreadMessages(c, os.Getenv("SOURCE_MAILBOX"), n)
	moveMessagesToDestination(c, seqset)

	seqsetToFile := getUnreadMessages(c, os.Getenv("DESTINATION_MAILBOX"), n)
	section, messages := getMessagesBody(c, n, seqsetToFile)

	for msg := range messages {
		subject, body := getMessageContent(section, msg)
		writeEMLFile(outputPath, subject, body)
	}

	// convert to epub
	log.Println("Converting to epub...")
	runCommand(fmt.Sprintf("email-to-epub emails/*.eml -o %s", outputFile))

	// upload to dropbox
	log.Println("Uploading to dropbox...")
	runCommand(fmt.Sprintf("rclone copy %s \"dropbox:%s\"", outputFile, os.Getenv("DROPBOX_UPLOAD_PATH")))
	//}

	log.Println("Done")
}
