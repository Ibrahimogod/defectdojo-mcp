package safety

import "testing"

func TestAllowMethod(t *testing.T) {
	tests := []struct {
		method            string
		enableDestructive bool
		want              bool
	}{
		{"GET", false, true},
		{"POST", false, true},
		{"PATCH", false, true},
		{"PUT", false, true},
		{"DELETE", false, false},
		{"DELETE", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.method+"/destructive="+boolStr(tt.enableDestructive), func(t *testing.T) {
			err := AllowMethod(tt.method, tt.enableDestructive)
			got := err == nil
			if got != tt.want {
				t.Errorf("AllowMethod(%q, %v) allowed = %v, want %v (err: %v)", tt.method, tt.enableDestructive, got, tt.want, err)
			}
		})
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
