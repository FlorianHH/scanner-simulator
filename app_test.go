package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestSendFraming(t *testing.T) {
	app := NewApp()
	// app.ctx is nil — emit() is a no-op, which is correct for tests.

	const port = 19876
	if err := app.StartListening(port); err != nil {
		t.Fatalf("StartListening: %v", err)
	}
	defer app.StopListening()

	// Dial as the middleware client.
	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	// Give the accept goroutine time to run and set app.conn.
	time.Sleep(50 * time.Millisecond)

	const payload = "HELLO123"
	if err := app.Send(payload); err != nil {
		t.Fatalf("Send: %v", err)
	}

	want := append([]byte{0x02}, append([]byte(payload), 0x03)...)
	got := make([]byte, len(want))
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("ReadFull: %v", err)
	}

	if !bytes.Equal(got, want) {
		t.Errorf("received %v, want %v", got, want)
	}
}

func TestBatchSendsAllItems(t *testing.T) {
	app := NewApp()

	const port = 19877
	if err := app.StartListening(port); err != nil {
		t.Fatalf("StartListening: %v", err)
	}
	defer app.StopListening()

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	time.Sleep(50 * time.Millisecond)

	items := []string{"A", "BB", "CCC"}
	if err := app.StartBatch(items, 20, 1); err != nil {
		t.Fatalf("StartBatch: %v", err)
	}

	for _, item := range items {
		want := append([]byte{0x02}, append([]byte(item), 0x03)...)
		got := make([]byte, len(want))
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		if _, err := io.ReadFull(conn, got); err != nil {
			t.Fatalf("ReadFull for %q: %v", item, err)
		}
		if !bytes.Equal(got, want) {
			t.Errorf("item %q: got %v, want %v", item, got, want)
		}
	}
}

func TestSendFailsWhenNotConnected(t *testing.T) {
	app := NewApp()
	err := app.Send("test")
	if err == nil {
		t.Error("expected error when sending with no connection, got nil")
	}
}

func TestStopListeningCancelsBatch(t *testing.T) {
	app := NewApp()

	const port = 19878
	if err := app.StartListening(port); err != nil {
		t.Fatalf("StartListening: %v", err)
	}
	defer app.StopListening()

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()

	// Give the accept goroutine time to run and set app.conn.
	time.Sleep(50 * time.Millisecond)

	// Start a long-running batch (10s delay per item — will be cancelled).
	items := []string{"X", "Y", "Z"}
	if err := app.StartBatch(items, 10000, 1); err != nil {
		t.Fatalf("StartBatch: %v", err)
	}

	// Stop should cancel the batch without deadlocking.
	done := make(chan struct{})
	go func() {
		app.StopListening()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("StopListening timed out — possible deadlock")
	}
}

func TestBatchLoopsMultipleTimes(t *testing.T) {
	app := NewApp()
	const port = 19879
	if err := app.StartListening(port); err != nil {
		t.Fatalf("StartListening: %v", err)
	}
	defer app.StopListening()

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()
	time.Sleep(50 * time.Millisecond)

	items := []string{"A", "B"}
	const loops = 3
	if err := app.StartBatch(items, 10, loops); err != nil {
		t.Fatalf("StartBatch: %v", err)
	}

	for cycle := 1; cycle <= loops; cycle++ {
		for _, item := range items {
			want := append([]byte{0x02}, append([]byte(item), 0x03)...)
			got := make([]byte, len(want))
			conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			if _, err := io.ReadFull(conn, got); err != nil {
				t.Fatalf("cycle %d item %q: ReadFull: %v", cycle, item, err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("cycle %d item %q: got %v, want %v", cycle, item, got, want)
			}
		}
	}

	// Confirm no extra frames arrive (batch terminated after loops cycles).
	extra := make([]byte, 1)
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	n, err := conn.Read(extra)
	if n > 0 || err == nil {
		t.Errorf("unexpected data after batch completion: n=%d err=%v", n, err)
	}
}

func TestBatchInfiniteLoopStopsOnCancel(t *testing.T) {
	app := NewApp()
	const port = 19880
	if err := app.StartListening(port); err != nil {
		t.Fatalf("StartListening: %v", err)
	}
	defer app.StopListening()

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer conn.Close()
	time.Sleep(50 * time.Millisecond)

	if err := app.StartBatch([]string{"X"}, 10, 0); err != nil {
		t.Fatalf("StartBatch: %v", err)
	}

	want := []byte{0x02, 'X', 0x03}
	got := make([]byte, len(want))
	conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	if _, err := io.ReadFull(conn, got); err != nil {
		t.Fatalf("ReadFull: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}

	done := make(chan struct{})
	go func() {
		app.StopBatch()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("StopBatch timed out — possible deadlock")
	}
}
