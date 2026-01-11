package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Field represents input fields in the form
type Field int

const (
	FieldURL Field = iota
	FieldConcurrent
	FieldOutputFolder
	FieldSubtitles
	FieldPlaylist
	FieldExtraFlags
	FieldDownloadButton
)

// Model represents the application state
type Model struct {
	// Form fields
	url          string
	concurrent   string
	outputFolder string
	subtitles    bool
	playlist     bool
	extraFlags   string

	// UI state
	focusedField Field
	cursorPos    map[Field]int
	err          string

	// Download state
	downloading     bool
	downloadCmd     string
	downloadOutput  []string
	downloadSuccess *bool
	spinnerFrame    int
}

// Spinner frames for download animation
var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// TickMsg is sent periodically during download for spinner animation
type TickMsg struct{}

// DownloadCompleteMsg is sent when download finishes
type DownloadCompleteMsg struct {
	Success  bool
	ExitCode int
}

// DownloadCompleteWithOutputMsg is sent when download finishes with all output
type DownloadCompleteWithOutputMsg struct {
	Success  bool
	ExitCode int
	Output   []string
}

// DownloadOutputMsg contains streaming output from yt-dlp
type DownloadOutputMsg struct {
	Line string
}

// InitialModel creates the initial application state
func InitialModel() Model {
	cursorPos := make(map[Field]int)
	cursorPos[FieldURL] = 0
	cursorPos[FieldConcurrent] = 0
	cursorPos[FieldOutputFolder] = 0
	cursorPos[FieldExtraFlags] = 0

	// Get default download folder
	defaultFolder := getDefaultDownloadFolder()

	return Model{
		url:             "",
		concurrent:      "4",
		outputFolder:    defaultFolder,
		subtitles:       false,
		playlist:        false,
		extraFlags:      "",
		focusedField:    FieldURL,
		cursorPos:       cursorPos,
		err:             "",
		downloading:     false,
		downloadCmd:     "",
		downloadOutput:  []string{},
		downloadSuccess: nil,
		spinnerFrame:    0,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return nil
}

// GetFieldValue returns the current value of a text field
func (m Model) GetFieldValue(field Field) string {
	switch field {
	case FieldURL:
		return m.url
	case FieldConcurrent:
		return m.concurrent
	case FieldOutputFolder:
		return m.outputFolder
	case FieldExtraFlags:
		return m.extraFlags
	default:
		return ""
	}
}

// SetFieldValue sets the value of a text field
func (m *Model) SetFieldValue(field Field, value string) {
	switch field {
	case FieldURL:
		m.url = value
	case FieldConcurrent:
		m.concurrent = value
	case FieldOutputFolder:
		m.outputFolder = value
	case FieldExtraFlags:
		m.extraFlags = value
	}
}

// GetCursorPos returns cursor position for the field
func (m Model) GetCursorPos(field Field) int {
	if pos, ok := m.cursorPos[field]; ok {
		return pos
	}
	return 0
}

// SetCursorPos sets cursor position for the field
func (m *Model) SetCursorPos(field Field, pos int) {
	m.cursorPos[field] = pos
}

// MoveCursorLeft moves cursor left in current field
func (m *Model) MoveCursorLeft() {
	field := m.focusedField
	if !m.isTextField(field) {
		return
	}

	pos := m.GetCursorPos(field)
	if pos > 0 {
		m.SetCursorPos(field, pos-1)
	}
}

// MoveCursorRight moves cursor right in current field
func (m *Model) MoveCursorRight() {
	field := m.focusedField
	if !m.isTextField(field) {
		return
	}

	value := m.GetFieldValue(field)
	pos := m.GetCursorPos(field)
	if pos < len(value) {
		m.SetCursorPos(field, pos+1)
	}
}

// MoveCursorHome moves cursor to start of field
func (m *Model) MoveCursorHome() {
	field := m.focusedField
	if m.isTextField(field) {
		m.SetCursorPos(field, 0)
	}
}

// MoveCursorEnd moves cursor to end of field
func (m *Model) MoveCursorEnd() {
	field := m.focusedField
	if m.isTextField(field) {
		value := m.GetFieldValue(field)
		m.SetCursorPos(field, len(value))
	}
}

// InsertChar inserts a character at cursor position
func (m *Model) InsertChar(ch rune) {
	field := m.focusedField
	if !m.isTextField(field) {
		return
	}

	value := m.GetFieldValue(field)
	pos := m.GetCursorPos(field)

	newValue := value[:pos] + string(ch) + value[pos:]
	m.SetFieldValue(field, newValue)
	m.SetCursorPos(field, pos+1)
}

// DeleteChar deletes character at cursor position (Delete key)
func (m *Model) DeleteChar() {
	field := m.focusedField
	if !m.isTextField(field) {
		return
	}

	value := m.GetFieldValue(field)
	pos := m.GetCursorPos(field)

	if pos < len(value) {
		newValue := value[:pos] + value[pos+1:]
		m.SetFieldValue(field, newValue)
	}
}

// Backspace deletes character before cursor position
func (m *Model) Backspace() {
	field := m.focusedField
	if !m.isTextField(field) {
		return
	}

	value := m.GetFieldValue(field)
	pos := m.GetCursorPos(field)

	if pos > 0 {
		newValue := value[:pos-1] + value[pos:]
		m.SetFieldValue(field, newValue)
		m.SetCursorPos(field, pos-1)
	}
}

// PasteFromClipboard pastes clipboard content at cursor position
func (m *Model) PasteFromClipboard() {
	field := m.focusedField
	if !m.isTextField(field) {
		return
	}

	// Get clipboard content using xclip, xsel, or wl-paste (Linux)
	clipboardText := getClipboardContent()
	if clipboardText == "" {
		return
	}

	value := m.GetFieldValue(field)
	pos := m.GetCursorPos(field)

	// Insert clipboard content at cursor position
	newValue := value[:pos] + clipboardText + value[pos:]
	m.SetFieldValue(field, newValue)
	m.SetCursorPos(field, pos+len(clipboardText))
}

// isTextField checks if field is a text input field
func (m Model) isTextField(field Field) bool {
	return field == FieldURL || field == FieldConcurrent ||
		field == FieldOutputFolder || field == FieldExtraFlags
}

// AddOutputLine adds a line to download output
func (m *Model) AddOutputLine(line string) {
	m.downloadOutput = append(m.downloadOutput, strings.TrimSpace(line))
}

// GetLastOutputLines returns last n lines of output
func (m Model) GetLastOutputLines(n int) []string {
	if len(m.downloadOutput) <= n {
		return m.downloadOutput
	}
	return m.downloadOutput[len(m.downloadOutput)-n:]
}

// getClipboardContent retrieves clipboard content (Linux)
func getClipboardContent() string {
	// Try xclip first
	cmd := exec.Command("xclip", "-o", "-selection", "clipboard")
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	// Try xsel
	cmd = exec.Command("xsel", "--clipboard", "--output")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	// Try wl-paste (Wayland)
	cmd = exec.Command("wl-paste")
	output, err = cmd.Output()
	if err == nil && len(output) > 0 {
		return strings.TrimSpace(string(output))
	}

	return ""
}

// getDefaultDownloadFolder returns the default download folder path
func getDefaultDownloadFolder() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "."
	}

	downloadFolder := filepath.Join(homeDir, "yt-dlp Downloads")

	// Check if folder exists
	if _, err := os.Stat(downloadFolder); err == nil {
		return downloadFolder
	}

	// Fallback to current directory if folder doesn't exist
	return "."
}
