package main

import (
	"email-to-epub/internal"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func main() {
	// init
	outputPath := internal.InitEMLDestination()
	outputFile := fmt.Sprintf("output-%s.epub", uuid.New())
	limit := internal.StringToInt(os.Getenv("N"))

	// app
	c := internal.LoginToIMAPServer()

	seqsetToFile := internal.GetUnreadMessages(c, os.Getenv("SOURCE_MAILBOX"), limit)
	section, messages := internal.GetMessagesBody(c, limit, seqsetToFile)
	for msg := range messages {
		subject, body := internal.GetMessageContent(section, msg)
		internal.WriteEMLFile(outputPath, subject, body)
	}

	internal.MoveMessagesToDestination(c, seqsetToFile)

	// convert to epub
	log.Info().Msg("Converting to epub...")
	internal.RunCommand(fmt.Sprintf("email-to-epub emails/*.eml -o %s", outputFile))
	log.Info().Msg("Done")
}
