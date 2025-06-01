package container

import (
	"fmt"
	"io"
	"strings"
)

// streamLogsFromPipes handles the common log streaming logic for both Docker and Podman
// This eliminates duplicate code between the two implementations
func streamLogsFromPipes(stdout, stderr io.ReadCloser, logChan chan<- string, stopChan <-chan struct{}) {
	// Start goroutine to read from stdout
	go func() {
		defer stdout.Close()
		streamFromReader(stdout, logChan, stopChan)
	}()

	// Start goroutine to read from stderr
	go func() {
		defer stderr.Close()
		streamFromReader(stderr, logChan, stopChan)
	}()
}

// streamFromReader reads from an io.Reader and sends lines to the log channel
func streamFromReader(reader io.Reader, logChan chan<- string, stopChan <-chan struct{}) {
	buf := make([]byte, 1024)
	for {
		select {
		case <-stopChan:
			return
		default:
			n, err := reader.Read(buf)
			if n > 0 {
				// Split by lines and send each line
				lines := strings.Split(strings.TrimSpace(string(buf[:n])), "\n")
				for _, line := range lines {
					if line != "" {
						select {
						case logChan <- line:
						case <-stopChan:
							return
						}
					}
				}
			}
			if err != nil {
				return
			}
		}
	}
}

// parseLogOutput is a shared utility for parsing container log output
func parseLogOutput(output []byte) []string {
	logLines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(logLines) == 1 && logLines[0] == "" {
		return []string{} // Return empty slice for empty logs
	}
	return logLines
}

// buildLogCommand creates the command arguments for getting container logs
func buildLogCommand(containerName string, tailLines int, follow bool) []string {
	args := []string{"logs"}
	if follow {
		args = append(args, "--follow")
		args = append(args, "--tail", "50") // Limit tail for streaming
	} else {
		args = append(args, "--tail", fmt.Sprintf("%d", tailLines))
	}
	args = append(args, containerName)
	return args
}
