package internal

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/go-shiori/go-readability"
	"github.com/rs/zerolog/log"
)

func LoginToIMAPServer() *client.Client {
	imapAddress := fmt.Sprintf("%s:%s", os.Getenv("IMAP_SERVER_HOST"), os.Getenv("IMAP_SERVER_PORT"))
	c, err := client.DialTLS(imapAddress, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to IMAP server")
	}
	log.Info().Msg("Connected to IMAP server")

	if err := c.Login(os.Getenv("IMAP_USERNAME"), os.Getenv("IMAP_PASSWORD")); err != nil {
		log.Fatal().Err(err).Msg("Failed to login")
	}
	log.Info().Msg("Logged in")

	return c
}

func GetUnreadMessages(c *client.Client, mailbox string, n uint32) *imap.SeqSet {
	mbox, err := c.Select(mailbox, false)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to select mailbox")
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
		log.Fatal().Err(err).Msg("Failed to fetch messages")
	}

	return seqset
}

func GetMessagesBody(c *client.Client, n uint32, seqset *imap.SeqSet) (imap.BodySectionName, chan *imap.Message) {
	var section imap.BodySectionName
	items := []imap.FetchItem{section.FetchItem()}

	messages := make(chan *imap.Message, n)
	go func() {
		if err := c.Fetch(seqset, items, messages); err != nil {
			log.Fatal().Err(err).Msg("Failed to get message body")
		}
	}()

	return section, messages
}

func GetMessageContent(section imap.BodySectionName, msg *imap.Message) (string, string) {
	r := msg.GetBody(&section)

	mr, err := mail.CreateReader(r)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to parse: %s", r)
	}

	body := ""

	header := mr.Header
	subject, err := header.Subject()
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to get subject")
	}
	log.Info().Msgf("Fetching: %s", subject)

	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal().Err(err).Msgf("Failed to get %s", subject)
		}

		switch p.Header.(type) {
		case *mail.InlineHeader:
			b, _ := io.ReadAll(p.Body)
			bodyRaw := string(b)

			// strip formatting
			bodyReader := strings.NewReader(bodyRaw)

			article, err := readability.FromReader(bodyReader, nil)
			if err != nil {
				log.Fatal().Err(err).Msgf("Failed to parse %s", subject)
			}

			body = article.Content
		}
	}

	return subject, body
}

func WriteEMLFile(outputPath string, subject string, body string) {
	// Create a new file
	file, err := os.Create(fmt.Sprintf("%s/%s.eml", outputPath, sanitizeFilename(subject)))
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create file")
	}
	defer file.Close()

	var h mail.Header
	h.SetSubject(subject)

	mw, err := mail.CreateWriter(file, h)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create writer")
	}

	// Create a text part
	tw, err := mw.CreateInline()
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create inline")
	}
	var th mail.InlineHeader
	th.Set("Content-Type", "text/html")
	w, err := tw.CreatePart(th)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create part")
	}

	_, err = io.WriteString(w, body)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to write body")
	}

	w.Close()
	tw.Close()

	mw.Close()
}

func MoveMessagesToDestination(c *client.Client, seqset *imap.SeqSet) {
	log.Info().Msg("Moving messages to destination")
	if err := c.Move(seqset, os.Getenv("DESTINATION_MAILBOX")); err != nil {
		log.Fatal().Err(err).Msg("Error on move to epub")
	}
}
