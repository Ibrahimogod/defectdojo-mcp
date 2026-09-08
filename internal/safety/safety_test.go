package safety

import "testing"

func TestAllowMethod(t *testing.T) {
	tests := []struct {
		method string
		want   bool
	}{
		{"GET", true},
		{"POST", false},
		{"PATCH", false},
		{"PUT", false},
		{"DELETE", false},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			err := AllowMethod(tt.method)
			got := err == nil
			if got != tt.want {
				t.Errorf("AllowMethod(%q) allowed = %v, want %v (err: %v)", tt.method, got, tt.want, err)
			}
		})
	}
}
