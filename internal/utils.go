package internal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

func InitEMLDestination() string {
	outputPath := "emails"
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		err = os.Mkdir(outputPath, os.ModeDir|0755)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to create output directory: %s", outputPath)
		}
	}

	return outputPath
}

func StringToInt(s string) uint32 {
	v, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to convert string to integer: %s", s)
	}
	return uint32(v)
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

func RunCommand(command string) {
	cmd := exec.Command("/bin/sh", "-c", command)

	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create stdout pipe")
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to create stderr pipe")
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		log.Fatal().Err(err).Msgf("Failed to start command")
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
		log.Fatal().Err(err).Msgf("Failed to wait for command")
	}
}
