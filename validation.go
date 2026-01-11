package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ValidateInputs validates all form inputs before download
func ValidateInputs(m Model) error {
	// Validate URL is not empty
	if strings.TrimSpace(m.url) == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Validate output folder
	folder := strings.TrimSpace(m.outputFolder)
	if folder == "" {
		return fmt.Errorf("Output folder cannot be empty")
	}

	// If not current directory, validate it exists and is writable
	if folder != "." {
		// Check if path exists
		info, err := os.Stat(folder)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("Output folder does not exist: %s", folder)
			}
			return fmt.Errorf("Cannot access output folder: %s", err.Error())
		}

		// Check if it's a directory
		if !info.IsDir() {
			return fmt.Errorf("Output path is not a directory: %s", folder)
		}

		// Check if writable by trying to create a test file
		testFile := filepath.Join(folder, ".hlsdownloader_test")
		f, err := os.Create(testFile)
		if err != nil {
			return fmt.Errorf("Output folder is not writable: %s", folder)
		}
		f.Close()
		os.Remove(testFile)
	}

	return nil
}
