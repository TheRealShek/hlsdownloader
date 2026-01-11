# yt-dlp TUI Downloader

A terminal-based GUI (TUI) tool for downloading videos with yt-dlp, built with Go and Bubble Tea.

## Features

- **Interactive TUI**: Clean, user-friendly terminal interface with box-drawing characters
- **Real-time Progress**: Live streaming output shows download progress as it happens
- **Smart Defaults**: Pre-configured for best quality video + audio in MP4 format
- **Customizable Options**:
  - Concurrent fragment downloads (default: 4)
  - Custom output folder
  - Subtitle downloads (auto + manual)
  - Playlist support
  - Advanced flags support
- **Input Validation**: Checks URL presence and folder writability before download
- **Clipboard Support**: Ctrl+V to paste URLs and paths
- **Keyboard Navigation**: Full keyboard control with intuitive shortcuts

## Prerequisites

- **Go 1.21+** (for building)
- **yt-dlp** (must be installed and in PATH)

Install yt-dlp:
```bash
# Using pip
pip install yt-dlp

# Or using your package manager
# macOS: brew install yt-dlp
# Linux: apt install yt-dlp / pacman -S yt-dlp
```

## Installation

```bash
# Clone the repository
cd /path/to/HlsDownloader

# Build and install
go install .

# Or just build
go build
```

## Usage

Simply run the executable:

```bash
hlsdownloader
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Tab / ↓ | Navigate to next field |
| Shift+Tab / ↑ | Navigate to previous field |
| Enter | Toggle checkbox or start download |
| Space | Toggle checkbox or insert space in text field |
| Left / Right | Move cursor within text field |
| Home / End | Jump to start/end of field |
| Backspace / Delete | Remove character |
| Ctrl+V | Paste from clipboard |
| q / Ctrl+C | Quit application |

## Configuration Options

### URL (Required)
The video URL to download. Supports YouTube and hundreds of other sites via yt-dlp.

### Concurrent Fragments
Number of parallel connections for downloading (default: 4). Higher values can speed up downloads but may be throttled by some servers.

### Output Folder
Directory where videos will be saved (default: current directory). The tool validates that the folder exists and is writable.

### Subtitles
When enabled, downloads available subtitles (both auto-generated and manual) using `--write-subs --write-auto-subs`.

### Playlist Mode
- **Unchecked** (default): Downloads only the single video, even if URL is a playlist
- **Checked**: Downloads entire playlist if URL contains one

### Extra Flags
Advanced yt-dlp flags for power users. Examples:
- `--format-sort res:1080` - Prefer 1080p
- `--cookies cookies.txt` - Use cookies file
- `--proxy socks5://127.0.0.1:1080` - Use proxy

## Default Command

The tool builds commands based on this template:

```bash
yt-dlp -f "bv*+ba/b" --merge-output-format mp4 --newline [OPTIONS] [URL]
```

- `-f "bv*+ba/b"`: Best video + best audio, fallback to best combined
- `--merge-output-format mp4`: Output as MP4
- `--newline`: Force line-by-line output for real-time progress display

## Download Progress

During download, you'll see:
- Animated spinner (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏)
- Full command being executed
- Live streaming output from yt-dlp showing:
  - Download percentage
  - File size
  - Download speed
  - ETA
- Success (✓) or failure (✗) status on completion

## Project Structure

```
HlsDownloader/
├── main.go         # Entry point, yt-dlp availability check
├── model.go        # Data structures and state management
├── view.go         # UI rendering logic
├── update.go       # Event handling and streaming output
├── validation.go   # Input validation
├── executor.go     # Command building and execution
├── go.mod          # Go module definition
└── README.md       # This file
```

## Technical Details

- **Framework**: [Bubble Tea](https://github.com/charmbracelet/bubbletea) for TUI
- **Language**: Go
- **Architecture**: Modular design with separation of concerns
- **Streaming**: Real-time output using channel-based message passing
- **Clipboard**: Supports xclip, xsel, and wl-paste (Linux)

## Error Handling

The application handles:
- Missing yt-dlp binary (exit on startup)
- Empty URL validation
- Non-existent output folder
- Non-writable output folder
- Command execution failures with exit codes

## License

This project is provided as-is for personal and educational use.

## Contributing

This is a personal project. Feel free to fork and modify for your own needs.
