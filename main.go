package main

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Check if yt-dlp is available
	if !CheckYtDlpAvailable() {
		fmt.Println("Error: yt-dlp is not installed or not in PATH")
		fmt.Println("Please install yt-dlp first:")
		fmt.Println("  pip install yt-dlp")
		fmt.Println("  or visit: https://github.com/yt-dlp/yt-dlp")
		os.Exit(1)
	}

	// Ensure default download folder exists
	ensureDefaultDownloadFolder()

	// Create and run the Bubble Tea program
	p := tea.NewProgram(InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// ensureDefaultDownloadFolder creates "yt-dlp Downloads" folder if it doesn't exist
func ensureDefaultDownloadFolder() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if can't get home dir
		homeDir = "."
	}

	downloadFolder := filepath.Join(homeDir, "yt-dlp Downloads")

	// Check if folder exists
	if _, err := os.Stat(downloadFolder); os.IsNotExist(err) {
		// Create folder with appropriate permissions
		err := os.MkdirAll(downloadFolder, 0755)
		if err != nil {
			// If creation fails, silently continue (will use current dir)
			return
		}
	}
}
