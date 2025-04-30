package main

import "testing"

func TestColorParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected RGBColor
		wantErr  bool
	}{
		{"#ff0080", RGBColor{R: uint8(255), G: uint8(0), B: uint8(128)}, false},
		{"#00ff00", RGBColor{R: uint8(0), G: uint8(255), B: uint8(0)}, false},
		{"invalid", RGBColor{}, true},
		{"#12345", RGBColor{}, true},
		{"#gg0000", RGBColor{}, true},
		{"#000000", RGBColor{}, false},
		{"", RGBColor{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseColor(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseColor(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && result.Compare(tt.expected) != 0 {
				t.Errorf("parseColor(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}
