package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestSSCCCheckDigit(t *testing.T) {
	tests := []struct {
		input string
		want  byte
	}{
		{"00000000000000000", '0'}, // all zeros: sum=0, check=0
		{"00000000000000001", '7'}, // rightmost=1×3=3, check=7
		{"00000000000000002", '4'}, // rightmost=2×3=6, check=4
		{"34001234500000001", '1'}, // sum=49, check=1
	}
	for _, tt := range tests {
		got := ssccCheckDigit(tt.input)
		if got != tt.want {
			t.Errorf("ssccCheckDigit(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestGenerateRange(t *testing.T) {
	t.Run("with check digit", func(t *testing.T) {
		items := generateRange(1, 3, true)
		want := []string{
			"000000000000000017",
			"000000000000000024",
			"000000000000000031",
		}
		if len(items) != len(want) {
			t.Fatalf("got %d items, want %d", len(items), len(want))
		}
		for i, w := range want {
			if items[i] != w {
				t.Errorf("items[%d] = %q, want %q", i, items[i], w)
			}
		}
	})

	t.Run("without check digit", func(t *testing.T) {
		items := generateRange(100, 3, false)
		want := []string{"100", "101", "102"}
		if len(items) != len(want) {
			t.Fatalf("got %d items, want %d", len(items), len(want))
		}
		for i, w := range want {
			if items[i] != w {
				t.Errorf("items[%d] = %q, want %q", i, items[i], w)
			}
		}
	})

	t.Run("single item", func(t *testing.T) {
		items := generateRange(42, 1, false)
		if len(items) != 1 || items[0] != "42" {
			t.Errorf("got %v, want [42]", items)
		}
	})
}

func TestStartRangeSendsCorrectFrames(t *testing.T) {
	app := NewApp()
	const port = 19881
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

	if err := app.StartRange("1", 3, 0, true); err != nil {
		t.Fatalf("StartRange: %v", err)
	}

	want := []string{"000000000000000017", "000000000000000024", "000000000000000031"}
	for _, w := range want {
		frame := append([]byte{0x02}, append([]byte(w), 0x03)...)
		got := make([]byte, len(frame))
		conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		if _, err := io.ReadFull(conn, got); err != nil {
			t.Fatalf("ReadFull for %q: %v", w, err)
		}
		if !bytes.Equal(got, frame) {
			t.Errorf("frame: got %v, want %v", got, frame)
		}
	}

	// Drain: no extra frames
	conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	extra := make([]byte, 1)
	if _, err := conn.Read(extra); err == nil {
		t.Error("expected no extra frames after range completes")
	}
}

func TestStartRangeValidation(t *testing.T) {
	app := NewApp()
	if err := app.StartRange("abc", 5, 0, false); err == nil {
		t.Error("expected error for non-numeric start")
	}
	if err := app.StartRange("1", 0, 0, false); err == nil {
		t.Error("expected error for count=0")
	}
	if err := app.StartRange("1", -1, 0, false); err == nil {
		t.Error("expected error for negative count")
	}
	// overflow: start=99999999999999998 + count=5 exceeds 17-digit max
	if err := app.StartRange("99999999999999998", 5, 0, true); err == nil {
		t.Error("expected error for range overflow")
	}
	// no client connected — StartBatch will return error
	if err := app.StartRange("1", 5, 0, false); err == nil {
		t.Error("expected error when no client connected")
	}
}
