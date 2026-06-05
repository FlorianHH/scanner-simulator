# Scanner Simulator — Development Guide

## Prerequisites

- Go 1.22+
- Node.js 18+ and npm (for frontend build)
- Wails v2 CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

Verify all dependencies: `wails doctor`

## Running in Development

```bash
wails dev
```

Opens the app window with hot-reload. Use netcat to simulate middleware:

```bash
nc 127.0.0.1 9000
```

1. Click **Start Listening** in the app (port 9000)
2. Connect netcat — status dot turns green
3. Type a barcode in the Manual Send field, press Send
4. Verify `^BHELLO123^C` appears in netcat (STX + data + ETX)
5. Test Batch Mode: paste barcodes one per line, set delay, click Start
6. Disconnect netcat (Ctrl+C) — app auto-reconnects to next connection

## Running Tests

```bash
go test ./... -v
```

## Building for Production

```bash
wails build
```

Output: `build/bin/scanner-simulator` (macOS/Linux) or `build/bin/scanner-simulator.exe` (Windows).
