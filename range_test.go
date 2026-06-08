package main

import "testing"

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
