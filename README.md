# Scanner Simulator

A desktop app that simulates a barcode scanner over TCP. It listens on a configurable port, waits for a middleware client to connect, and sends barcodes wrapped in STX/ETX framing — the protocol used by most industrial barcode scanners.

## Features

- **Manual Send** — type a barcode and send it instantly
- **Batch Mode** — send a list of barcodes with a configurable delay between scans, with optional looping
- **Range Mode** — generate and send a sequential range of SSCC-18 barcodes (with GS1 check digit)
- Live activity log with timestamps
- Auto-reconnect when the client disconnects

## Download

Pre-built binaries for Windows, macOS, and Linux are available on the [Releases](../../releases) page.

## Usage

1. Enter the TCP port your middleware listens on and click **Start Listening**
2. Start your middleware (or use `nc 127.0.0.1 9000` to test)
3. Once connected, use Manual Send, Batch, or Range mode to send barcodes

Each barcode is sent as `STX + data + ETX` (`0x02 ... 0x03`).

## Building from Source

**Prerequisites:** Go 1.22+, Node.js 18+, [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
wails build
```

For development with hot reload:

```bash
wails dev
```

See [DEVELOPMENT.md](DEVELOPMENT.md) for details.
