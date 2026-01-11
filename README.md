# yt-dlp TUI Downloader

A terminal-based GUI tool for downloading videos with yt-dlp, built with Go and Bubble Tea.

## Features

- **Interactive TUI**: Clean terminal interface with box-drawing characters
- **Real-time Progress**: Live streaming output shows download progress as it happens
- **Smart Defaults**: Pre-configured for best quality video + audio in MP4 format
- **Auto-rename**: Downloaded files are automatically renamed to unique 20-character alphanumeric IDs
- **Default Downloads Folder**: Creates and uses `~/yt-dlp Downloads` by default
- **Customizable Options**: Concurrent downloads, output folder, subtitles, playlists, advanced flags
- **Input Validation**: Checks URL presence and folder writability before download
- **Clipboard Support**: Ctrl+V to paste URLs and paths (Linux: xclip/xsel/wl-paste)
- **Keyboard Navigation**: Full keyboard control with intuitive shortcuts

## Prerequisites

- **Go 1.21+** (for building)
- **yt-dlp** (must be installed and in PATH)

```bash
# Install yt-dlp
pip install yt-dlp
# Or: brew install yt-dlp (macOS) / apt install yt-dlp (Linux)
```

## Installation

```bash
cd /path/to/HlsDownloader
go install .
# Or: go build
```

## Usage

```bash
hlsdownloader
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| Tab / ↓ | Navigate to next field |
| Shift+Tab / ↑ | Navigate to previous field |
| Enter | Toggle checkbox or start download |
| Space | Toggle checkbox or insert space |
| Left / Right | Move cursor within text field |
| Home / End | Jump to start/end of field |
| Backspace / Delete | Remove character |
| Ctrl+V | Paste from clipboard |
| q / Ctrl+C | Quit application |

## Configuration

### URL (Required)
Video URL to download (supports YouTube and 1000+ sites via yt-dlp)

### Concurrent Fragments
Number of parallel connections (default: 4). Higher values may speed up downloads.

### Output Folder
Download destination (default: `~/yt-dlp Downloads`). Must exist and be writable.

### Subtitles
Downloads available subtitles (auto-generated + manual) via `--write-subs --write-auto-subs`

### Playlist Mode
- Unchecked: Single video only
- Checked: Downloads entire playlist

### Extra Flags
Advanced yt-dlp options (e.g., `--format-sort res:1080`, `--cookies cookies.txt`)

## Default Command

```bash
yt-dlp -f "bv*+ba/b" --merge-output-format mp4 --newline [OPTIONS] [URL]
```

- Best video + best audio, merged to MP4
- Newline-separated output for real-time progress display
- Files auto-renamed to 20-char unique IDs after download

## Project Structure

```
├── main.go         # Entry point, setup default folder
├── model.go        # Data structures and state
├── view.go         # UI rendering
├── update.go       # Event handling and file rename
├── validation.go   # Input validation
├── executor.go     # Command building
└── README.md
```

## Technical Details

- **Framework**: Bubble Tea (TUI framework)
- **Streaming**: Channel-based real-time output with carriage return handling
- **File Naming**: Cryptographic random 20-char alphanumeric IDs
- **Default Folder**: `~/yt-dlp Downloads` (auto-created on startup)

## License

Provided as-is for personal and educational use.
