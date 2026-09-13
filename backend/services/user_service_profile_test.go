package services

import "testing"

func TestNormalizeAndValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "normal email", input: "a@example.com", want: "a@example.com"},
		{name: "trim and lower", input: "  A@Example.COM  ", want: "a@example.com"},
		{name: "empty", input: "   ", wantErr: true},
		{name: "invalid", input: "abc", wantErr: true},
	}
	for _, tt := range tests {
		got, err := normalizeAndValidateEmail(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Fatalf("%s: expected error", tt.name)
			}
			continue
		}
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("%s: want %q got %q", tt.name, tt.want, got)
		}
	}
}
