package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx         context.Context
	mu          sync.Mutex
	listener    net.Listener
	conn        net.Conn
	batchCancel context.CancelFunc
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) emit(event string, data interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event, data)
	}
}

func (a *App) logEntry(level, message string) {
	a.emit("log:entry", map[string]string{
		"time":    time.Now().Format("15:04:05"),
		"level":   level,
		"message": message,
	})
}

// StartListening binds a TCP listener on the given port and starts the accept loop.
func (a *App) StartListening(port int) error {
	a.mu.Lock()
	if a.listener != nil {
		a.mu.Unlock()
		return fmt.Errorf("already listening")
	}
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		a.mu.Unlock()
		return err
	}
	a.listener = ln
	a.mu.Unlock()

	go a.acceptLoop(ln, port)
	return nil
}

func (a *App) acceptLoop(ln net.Listener, port int) {
	a.emit("status:listening", map[string]int{"port": port})
	a.logEntry("INFO", fmt.Sprintf("Listening on :%d", port))

	for {
		conn, err := ln.Accept()
		if err != nil {
			break // listener was closed by StopListening
		}

		a.mu.Lock()
		a.conn = conn
		a.mu.Unlock()

		remote := conn.RemoteAddr().String()
		a.emit("status:connected", map[string]string{"remoteAddr": remote})
		a.logEntry("INFO", fmt.Sprintf("Connected — %s", remote))

		// Drain reads to detect client disconnect (we never expect data from the client).
		buf := make([]byte, 64)
		for {
			_, err := conn.Read(buf)
			if err != nil {
				break
			}
		}

		a.mu.Lock()
		a.cancelBatch()
		a.conn = nil
		a.mu.Unlock()

		a.emit("status:disconnected", nil)
		a.logEntry("INFO", "Client disconnected, waiting for reconnect")
	}

	a.mu.Lock()
	a.listener = nil
	a.mu.Unlock()

	a.emit("status:idle", nil)
	a.logEntry("INFO", "Stopped listening")
}

// StopListening closes the listener and any active client connection.
func (a *App) StopListening() {
	a.mu.Lock()
	a.cancelBatch()
	ln := a.listener
	conn := a.conn
	a.listener = nil
	a.conn = nil
	a.mu.Unlock()

	if conn != nil {
		conn.Close()
	}
	if ln != nil {
		ln.Close()
	}
}

// Send wraps data in STX/ETX and writes it to the connected client.
func (a *App) Send(data string) error {
	a.mu.Lock()
	conn := a.conn
	a.mu.Unlock()

	if conn == nil {
		return fmt.Errorf("no client connected")
	}

	if _, err := conn.Write(frame(data)); err != nil {
		return err
	}

	a.logEntry("SENT", data)
	return nil
}

// StartBatch sends items sequentially with delayMs between each, starting immediately.
func (a *App) StartBatch(items []string, delayMs int) error {
	a.mu.Lock()
	if a.conn == nil {
		a.mu.Unlock()
		return fmt.Errorf("no client connected")
	}
	a.cancelBatch()
	ctx, cancel := context.WithCancel(context.Background())
	a.batchCancel = cancel
	a.mu.Unlock()

	go a.runBatch(ctx, items, delayMs)
	return nil
}

func (a *App) runBatch(ctx context.Context, items []string, delayMs int) {
	delay := time.Duration(delayMs) * time.Millisecond

	for i, item := range items {
		if i > 0 {
			select {
			case <-ctx.Done():
				a.emit("batch:done", nil)
				return
			case <-time.After(delay):
			}
		} else {
			select {
			case <-ctx.Done():
				a.emit("batch:done", nil)
				return
			default:
			}
		}

		if err := a.Send(item); err != nil {
			a.logEntry("ERR", fmt.Sprintf("Batch send failed: %v", err))
			a.emit("batch:done", nil)
			return
		}
		a.emit("batch:progress", map[string]int{"index": i + 1, "total": len(items)})
	}

	a.mu.Lock()
	a.batchCancel = nil
	a.mu.Unlock()

	a.emit("batch:done", nil)
}

// StopBatch cancels a running batch. Safe to call when no batch is running.
func (a *App) StopBatch() {
	a.mu.Lock()
	a.cancelBatch()
	a.mu.Unlock()
	a.emit("batch:done", nil)
}

// cancelBatch must be called with a.mu held.
func (a *App) cancelBatch() {
	if a.batchCancel != nil {
		a.batchCancel()
		a.batchCancel = nil
	}
}
