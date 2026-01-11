package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// BuildCommand constructs the yt-dlp command from model state
func BuildCommand(m Model) string {
	parts := []string{"yt-dlp"}

	// Base format flags
	parts = append(parts, "-f", `"bv*+ba/b"`)
	parts = append(parts, "--merge-output-format", "mp4")

	// Force newline output for better streaming
	parts = append(parts, "--newline")

	// Concurrent fragments
	concurrent := strings.TrimSpace(m.concurrent)
	if concurrent != "" && concurrent != "0" {
		if n, err := strconv.Atoi(concurrent); err == nil && n > 0 {
			parts = append(parts, "-N", concurrent)
		}
	}

	// Output folder
	folder := strings.TrimSpace(m.outputFolder)
	if folder != "" && folder != "." {
		outputTemplate := fmt.Sprintf(`"%s/%%(title)s.%%(ext)s"`, folder)
		parts = append(parts, "-o", outputTemplate)
	}

	// Subtitles
	if m.subtitles {
		parts = append(parts, "--write-subs", "--write-auto-subs")
	}

	// Playlist mode
	if !m.playlist {
		parts = append(parts, "--no-playlist")
	}

	// Extra flags
	extraFlags := strings.TrimSpace(m.extraFlags)
	if extraFlags != "" {
		parts = append(parts, extraFlags)
	}

	// URL (last)
	url := strings.TrimSpace(m.url)
	parts = append(parts, url)

	return strings.Join(parts, " ")
}

// ExecuteDownload runs the yt-dlp command and streams output
func ExecuteDownload(cmd string) tea.Cmd {
	return func() tea.Msg {
		// Parse command to execute
		cmdParts := parseCommand(cmd)
		if len(cmdParts) == 0 {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		// Create command
		execCmd := exec.Command(cmdParts[0], cmdParts[1:]...)

		// Get stdout and stderr pipes
		stdout, err := execCmd.StdoutPipe()
		if err != nil {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		stderr, err := execCmd.StderrPipe()
		if err != nil {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		// Start command
		if err := execCmd.Start(); err != nil {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		// Create channels for output streaming
		outputChan := make(chan string, 100)
		doneChan := make(chan bool)

		// Stream stdout
		go streamOutput(stdout, outputChan)

		// Stream stderr
		go streamOutput(stderr, outputChan)

		// Send output to UI in separate goroutine
		go func() {
			for range outputChan {
				// This would need to be sent via a channel back to the Update function
				// For now, we'll collect it in the command execution
			}
			doneChan <- true
		}() // Wait for command to complete
		err = execCmd.Wait()
		close(outputChan)
		<-doneChan

		success := err == nil
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}

		return DownloadCompleteMsg{
			Success:  success,
			ExitCode: exitCode,
		}
	}
}

// ExecuteDownloadWithStreaming runs yt-dlp and streams output line by line
func ExecuteDownloadWithStreaming(cmd string) tea.Cmd {
	return func() tea.Msg {
		cmdParts := parseCommand(cmd)
		if len(cmdParts) == 0 {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		execCmd := exec.Command(cmdParts[0], cmdParts[1:]...)

		stdout, err := execCmd.StdoutPipe()
		if err != nil {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		stderr, err := execCmd.StderrPipe()
		if err != nil {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		if err := execCmd.Start(); err != nil {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		// Read output in current goroutine
		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				// Output will be handled by batch reading
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				// Output will be handled by batch reading
			}
		}()

		err = execCmd.Wait()
		success := err == nil
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}

		return DownloadCompleteMsg{
			Success:  success,
			ExitCode: exitCode,
		}
	}
}

// StartDownload initiates download with streaming output
func StartDownload(m *Model) tea.Cmd {
	cmd := BuildCommand(*m)
	m.downloadCmd = cmd
	m.downloading = true
	m.downloadOutput = []string{}
	m.downloadSuccess = nil
	m.err = ""

	return tea.Batch(
		executeDownloadStreaming(cmd),
		tickSpinner(),
	)
}

// executeDownloadStreaming runs yt-dlp and sends output messages
func executeDownloadStreaming(cmdStr string) tea.Cmd {
	return func() tea.Msg {
		cmdParts := parseCommand(cmdStr)
		if len(cmdParts) == 0 {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		execCmd := exec.Command(cmdParts[0], cmdParts[1:]...)

		// Combine stdout and stderr
		stdout, err := execCmd.StdoutPipe()
		if err != nil {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		stderr, err := execCmd.StderrPipe()
		if err != nil {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		if err := execCmd.Start(); err != nil {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		// Read both streams
		go streamToUI(stdout)
		go streamToUI(stderr)

		err = execCmd.Wait()
		success := err == nil
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = 1
			}
		}

		return DownloadCompleteMsg{
			Success:  success,
			ExitCode: exitCode,
		}
	}
}

// streamToUI reads from reader and sends to UI (placeholder for actual implementation)
func streamToUI(reader io.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		// Lines will be captured by the actual execution
		_ = scanner.Text()
	}
}

// streamOutput reads lines from reader and sends to channel
func streamOutput(reader io.Reader, outputChan chan<- string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		outputChan <- scanner.Text()
	}
}

// parseCommand splits command string into parts, respecting quotes
func parseCommand(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(cmd); i++ {
		ch := cmd[i]

		if ch == '"' {
			inQuote = !inQuote
			continue
		}

		if ch == ' ' && !inQuote {
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			continue
		}

		current.WriteByte(ch)
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// tickSpinner returns a command that sends tick messages for spinner animation
func tickSpinner() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

// CheckYtDlpAvailable checks if yt-dlp is in PATH
func CheckYtDlpAvailable() bool {
	_, err := exec.LookPath("yt-dlp")
	return err == nil
}
