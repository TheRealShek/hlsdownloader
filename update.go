package main

import (
	"bufio"
	"os/exec"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

// Global channel for streaming download output
var downloadOutputChan chan tea.Msg

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case TickMsg:
		if m.downloading && m.downloadSuccess == nil {
			m.spinnerFrame++
			return m, tickSpinner()
		}
		return m, nil

	case DownloadOutputMsg:
		m.AddOutputLine(msg.Line)
		// Continue listening for more output
		if m.downloading && m.downloadSuccess == nil {
			return m, waitForOutput()
		}
		return m, nil

	case DownloadCompleteMsg:
		success := msg.Success
		m.downloadSuccess = &success
		return m, nil

	case DownloadCompleteWithOutputMsg:
		// Add all output lines
		for _, line := range msg.Output {
			if strings.TrimSpace(line) != "" {
				m.AddOutputLine(line)
			}
		}
		success := msg.Success
		m.downloadSuccess = &success
		return m, nil

	default:
		return m, nil
	}
}

// handleKeyPress processes keyboard input
func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global quit keys
	if msg.String() == "ctrl+c" || msg.String() == "q" {
		if !m.downloading {
			return m, tea.Quit
		}
	}

	// If downloading and finished, any key returns to form
	if m.downloading && m.downloadSuccess != nil {
		m.downloading = false
		m.downloadOutput = []string{}
		m.downloadSuccess = nil
		m.downloadCmd = ""
		m.err = ""
		return m, nil
	}

	// Don't handle input during active download
	if m.downloading {
		return m, nil
	}

	// Handle navigation
	switch msg.String() {
	case "tab", "down":
		return m.navigateNext(), nil

	case "shift+tab", "up":
		return m.navigatePrevious(), nil

	case "enter":
		return m.handleEnter()

	case "left":
		m.MoveCursorLeft()
		return m, nil

	case "right":
		m.MoveCursorRight()
		return m, nil

	case "home":
		m.MoveCursorHome()
		return m, nil

	case "end":
		m.MoveCursorEnd()
		return m, nil

	case "backspace":
		m.Backspace()
		return m, nil

	case "delete":
		m.DeleteChar()
		return m, nil

	case "ctrl+v":
		m.PasteFromClipboard()
		return m, nil

	case " ", "space":
		return m.handleSpace()

	default:
		// Handle character input for text fields
		return m.handleCharInput(msg)
	}
}

// navigateNext moves focus to next field
func (m Model) navigateNext() Model {
	switch m.focusedField {
	case FieldURL:
		m.focusedField = FieldConcurrent
	case FieldConcurrent:
		m.focusedField = FieldOutputFolder
	case FieldOutputFolder:
		m.focusedField = FieldSubtitles
	case FieldSubtitles:
		m.focusedField = FieldPlaylist
	case FieldPlaylist:
		m.focusedField = FieldExtraFlags
	case FieldExtraFlags:
		m.focusedField = FieldDownloadButton
	case FieldDownloadButton:
		m.focusedField = FieldURL
	}
	return m
}

// navigatePrevious moves focus to previous field
func (m Model) navigatePrevious() Model {
	switch m.focusedField {
	case FieldURL:
		m.focusedField = FieldDownloadButton
	case FieldConcurrent:
		m.focusedField = FieldURL
	case FieldOutputFolder:
		m.focusedField = FieldConcurrent
	case FieldSubtitles:
		m.focusedField = FieldOutputFolder
	case FieldPlaylist:
		m.focusedField = FieldSubtitles
	case FieldExtraFlags:
		m.focusedField = FieldPlaylist
	case FieldDownloadButton:
		m.focusedField = FieldExtraFlags
	}
	return m
}

// handleEnter processes Enter key
func (m Model) handleEnter() (Model, tea.Cmd) {
	switch m.focusedField {
	case FieldSubtitles:
		m.subtitles = !m.subtitles
		return m, nil

	case FieldPlaylist:
		m.playlist = !m.playlist
		return m, nil

	case FieldDownloadButton:
		return m.startDownload()

	default:
		// For text fields, Enter also submits
		return m.startDownload()
	}
}

// handleSpace processes Space key
func (m Model) handleSpace() (Model, tea.Cmd) {
	switch m.focusedField {
	case FieldSubtitles:
		m.subtitles = !m.subtitles
		return m, nil

	case FieldPlaylist:
		m.playlist = !m.playlist
		return m, nil

	default:
		// For text fields, space is a regular character
		m.InsertChar(' ')
		return m, nil
	}
}

// handleCharInput processes character input for text fields
func (m Model) handleCharInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	// Only handle single character input
	key := msg.String()
	if len(key) != 1 {
		return m, nil
	}

	ch := rune(key[0])

	// For concurrent field, only allow digits
	if m.focusedField == FieldConcurrent {
		if !unicode.IsDigit(ch) {
			return m, nil
		}
	}

	// Insert character if on a text field
	if m.isTextField(m.focusedField) {
		m.InsertChar(ch)
	}

	return m, nil
}

// startDownload validates inputs and starts download
func (m Model) startDownload() (Model, tea.Cmd) {
	// Clear previous error
	m.err = ""

	// Validate inputs
	if err := ValidateInputs(m); err != nil {
		m.err = err.Error()
		return m, nil
	}

	// Start download with live output streaming
	return m, startDownloadWithOutput(&m)
}

// startDownloadWithOutput initiates download and captures output
func startDownloadWithOutput(m *Model) tea.Cmd {
	cmd := BuildCommand(*m)
	m.downloadCmd = cmd
	m.downloading = true
	m.downloadOutput = []string{}
	m.downloadSuccess = nil
	m.spinnerFrame = 0

	return tea.Batch(
		streamDownloadOutput(cmd),
		tickSpinner(),
	)
}

// streamDownloadOutput runs yt-dlp and streams output line by line
func streamDownloadOutput(cmdStr string) tea.Cmd {
	return func() tea.Msg {
		cmdParts := parseCommand(cmdStr)
		if len(cmdParts) == 0 {
			return DownloadCompleteMsg{Success: false, ExitCode: 1}
		}

		// Create channel for streaming
		downloadOutputChan = make(chan tea.Msg, 100)

		go func() {
			defer close(downloadOutputChan)

			execCmd := exec.Command(cmdParts[0], cmdParts[1:]...)

			// Use CombinedOutput for simpler streaming
			outputPipe, err := execCmd.StdoutPipe()
			if err != nil {
				downloadOutputChan <- DownloadCompleteMsg{Success: false, ExitCode: 1}
				return
			}

			stderrPipe, err := execCmd.StderrPipe()
			if err != nil {
				downloadOutputChan <- DownloadCompleteMsg{Success: false, ExitCode: 1}
				return
			}

			if err := execCmd.Start(); err != nil {
				downloadOutputChan <- DownloadCompleteMsg{Success: false, ExitCode: 1}
				return
			}

			// Read stdout with byte-by-byte for carriage returns
			go func() {
				buf := make([]byte, 1)
				line := strings.Builder{}
				for {
					n, err := outputPipe.Read(buf)
					if n > 0 {
						ch := buf[0]
						if ch == '\n' {
							// New line - send accumulated line
							if line.Len() > 0 {
								downloadOutputChan <- DownloadOutputMsg{Line: line.String()}
								line.Reset()
							}
						} else if ch == '\r' {
							// Carriage return - send current line and reset
							if line.Len() > 0 {
								downloadOutputChan <- DownloadOutputMsg{Line: line.String()}
								line.Reset()
							}
						} else {
							line.WriteByte(ch)
						}
					}
					if err != nil {
						if line.Len() > 0 {
							downloadOutputChan <- DownloadOutputMsg{Line: line.String()}
						}
						break
					}
				}
			}()

			// Read stderr
			go func() {
				scanner := bufio.NewScanner(stderrPipe)
				for scanner.Scan() {
					downloadOutputChan <- DownloadOutputMsg{Line: scanner.Text()}
				}
			}()

			// Wait for command
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

			downloadOutputChan <- DownloadCompleteMsg{Success: success, ExitCode: exitCode}
		}()

		// Wait for and return first message
		return waitForOutput()()
	}
}

// waitForOutput waits for the next output message from the download channel
func waitForOutput() tea.Cmd {
	return func() tea.Msg {
		if downloadOutputChan != nil {
			return <-downloadOutputChan
		}
		return nil
	}
}
