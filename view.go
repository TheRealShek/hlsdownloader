package main

import (
	"fmt"
	"strings"
)

// View renders the UI
func (m Model) View() string {
	if m.downloading {
		return m.renderDownloadView()
	}
	return m.renderFormView()
}

// renderFormView renders the input form
func (m Model) renderFormView() string {
	var b strings.Builder

	// Header
	b.WriteString("\n")
	b.WriteString("  ╔════════════════════════════════════════════════════════════════════════════╗\n")
	b.WriteString("  ║                          yt-dlp Download Manager                           ║\n")
	b.WriteString("  ╚════════════════════════════════════════════════════════════════════════════╝\n")
	b.WriteString("\n")

	// URL field
	b.WriteString(m.renderTextField(FieldURL, "URL", m.url, true))
	b.WriteString("\n")

	// Concurrent fragments field
	b.WriteString(m.renderTextField(FieldConcurrent, "Concurrent Fragments (-N)", m.concurrent, false))
	b.WriteString("\n")

	// Output folder field
	b.WriteString(m.renderTextField(FieldOutputFolder, "Output Folder", m.outputFolder, false))
	b.WriteString("\n")

	// Subtitles checkbox
	b.WriteString(m.renderCheckbox(FieldSubtitles, "Subtitles", m.subtitles))
	b.WriteString("\n")

	// Playlist checkbox
	b.WriteString(m.renderCheckbox(FieldPlaylist, "Playlist Mode", m.playlist))
	b.WriteString("\n")

	// Extra flags field
	b.WriteString(m.renderTextField(FieldExtraFlags, "Extra Flags", m.extraFlags, false))
	b.WriteString("\n")

	// Command preview
	b.WriteString("  ──────────────────────────────────────────────────────────────────────────────\n")
	b.WriteString("  Command Preview:\n")
	cmd := BuildCommand(m)
	b.WriteString(m.wrapText(cmd, 76, "  "))
	b.WriteString("\n")
	b.WriteString("  ──────────────────────────────────────────────────────────────────────────────\n")
	b.WriteString("\n")

	// Download button
	b.WriteString(m.renderButton(FieldDownloadButton, "[ Download ]"))
	b.WriteString("\n\n")

	// Error message
	if m.err != "" {
		b.WriteString(m.renderError(m.err))
		b.WriteString("\n")
	}

	// Help text
	b.WriteString("  Tab/↑↓: Navigate  |  Space: Toggle  |  Enter: Submit  |  q/Ctrl+C: Quit\n")

	return b.String()
}

// renderDownloadView renders the download progress view
func (m Model) renderDownloadView() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("  ╔════════════════════════════════════════════════════════════════════════════╗\n")
	b.WriteString("  ║                              Download in Progress                          ║\n")
	b.WriteString("  ╚════════════════════════════════════════════════════════════════════════════╝\n")
	b.WriteString("\n")

	// Command that was executed
	b.WriteString("  Executing:\n")
	b.WriteString(m.wrapText(m.downloadCmd, 76, "  "))
	b.WriteString("\n")
	b.WriteString("  ──────────────────────────────────────────────────────────────────────────────\n")
	b.WriteString("\n")

	// Status with spinner
	if m.downloadSuccess == nil {
		spinner := spinnerFrames[m.spinnerFrame%len(spinnerFrames)]
		b.WriteString(fmt.Sprintf("  %s Download in progress...\n", spinner))
	} else if *m.downloadSuccess {
		b.WriteString("  ✓ SUCCESS\n")
	} else {
		b.WriteString("  ✗ FAILED\n")
	}
	b.WriteString("\n")

	// Output (last 15 lines)
	b.WriteString("  Output:\n")
	b.WriteString("  ──────────────────────────────────────────────────────────────────────────────\n")
	outputLines := m.GetLastOutputLines(15)
	if len(outputLines) == 0 {
		b.WriteString("  (waiting for output...)\n")
	} else {
		for _, line := range outputLines {
			wrapped := m.wrapText(line, 76, "  ")
			b.WriteString(wrapped)
			b.WriteString("\n")
		}
	}
	b.WriteString("  ──────────────────────────────────────────────────────────────────────────────\n")
	b.WriteString("\n")

	if m.downloadSuccess != nil {
		b.WriteString("  Press any key to return to form...\n")
	}

	return b.String()
}

// renderTextField renders a text input field
func (m Model) renderTextField(field Field, label, value string, required bool) string {
	focused := m.focusedField == field
	cursor := m.GetCursorPos(field)

	var b strings.Builder

	// Label
	prefix := "  "
	if required {
		prefix = "  * "
	}

	focusIndicator := " "
	if focused {
		focusIndicator = ">"
	}

	b.WriteString(fmt.Sprintf(" %s%s %s\n", focusIndicator, prefix, label))

	// Input box
	boxWidth := 74
	b.WriteString("  ┌")
	b.WriteString(strings.Repeat("─", boxWidth))
	b.WriteString("┐\n")

	// Value with cursor
	displayValue := value
	if focused && cursor <= len(value) {
		// Insert cursor character
		if cursor == len(value) {
			displayValue = value + "_"
		} else {
			displayValue = value[:cursor] + "_" + value[cursor:]
		}
	}

	// Pad or truncate to fit box
	if len(displayValue) > boxWidth-2 {
		displayValue = displayValue[:boxWidth-2]
	}
	padding := boxWidth - 2 - len(displayValue)

	b.WriteString("  │ ")
	b.WriteString(displayValue)
	b.WriteString(strings.Repeat(" ", padding))
	b.WriteString(" │\n")

	b.WriteString("  └")
	b.WriteString(strings.Repeat("─", boxWidth))
	b.WriteString("┘")

	return b.String()
}

// renderCheckbox renders a checkbox field
func (m Model) renderCheckbox(field Field, label string, checked bool) string {
	focused := m.focusedField == field

	focusIndicator := " "
	if focused {
		focusIndicator = ">"
	}

	checkMark := " "
	if checked {
		checkMark = "X"
	}

	return fmt.Sprintf(" %s   [%s] %s", focusIndicator, checkMark, label)
}

// renderButton renders a button
func (m Model) renderButton(field Field, label string) string {
	focused := m.focusedField == field

	if focused {
		return fmt.Sprintf("  > %s <", label)
	}
	return fmt.Sprintf("    %s", label)
}

// renderError renders an error message
func (m Model) renderError(err string) string {
	var b strings.Builder
	b.WriteString("  ┌─ Error ─────────────────────────────────────────────────────────────────────┐\n")
	wrapped := m.wrapText(err, 76, "  │ ")
	b.WriteString(wrapped)
	b.WriteString("\n")
	b.WriteString("  └──────────────────────────────────────────────────────────────────────────────┘")
	return b.String()
}

// wrapText wraps text to fit within specified width
func (m Model) wrapText(text string, width int, prefix string) string {
	if len(text) == 0 {
		return prefix
	}

	var lines []string
	words := strings.Fields(text)

	if len(words) == 0 {
		return prefix
	}

	currentLine := prefix
	prefixLen := len(prefix)

	for i, word := range words {
		// Check if adding this word would exceed width
		testLine := currentLine
		if i > 0 || len(currentLine) > prefixLen {
			testLine += " "
		}
		testLine += word

		if len(testLine) > width+prefixLen {
			// Line would be too long, start new line
			if len(currentLine) > prefixLen {
				lines = append(lines, currentLine)
				currentLine = prefix + word
			} else {
				// Word itself is too long, just add it
				currentLine = testLine
			}
		} else {
			currentLine = testLine
		}
	}

	if len(currentLine) > prefixLen {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}
