package main

import (
	"fmt"
	"os"

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

	// Create and run the Bubble Tea program
	p := tea.NewProgram(InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
