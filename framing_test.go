package main

import (
	"bytes"
	"testing"
)

func TestFrame(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []byte
	}{
		{
			name:  "empty string",
			input: "",
			want:  []byte{0x02, 0x03},
		},
		{
			name:  "ascii barcode",
			input: "ABC123",
			want:  []byte{0x02, 'A', 'B', 'C', '1', '2', '3', 0x03},
		},
		{
			name:  "utf8 content",
			input: "héllo",
			want:  append([]byte{0x02}, append([]byte("héllo"), 0x03)...),
		},
		{
			name:  "newline in data",
			input: "foo\nbar",
			want:  []byte{0x02, 'f', 'o', 'o', '\n', 'b', 'a', 'r', 0x03},
		},
		{
			name:  "single char",
			input: "X",
			want:  []byte{0x02, 'X', 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frame(tt.input)
			if !bytes.Equal(got, tt.want) {
				t.Errorf("frame(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
