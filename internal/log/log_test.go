package log

import "testing"

func TestRedact(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "authorization header with Token scheme",
			input: `Authorization: Token abcd1234EFGH5678`,
			want:  `Authorization: Token REDACTED`,
		},
		{
			name:  "authorization header with Bearer scheme",
			input: `Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.payload.sig`,
			want:  `Authorization: Bearer REDACTED`,
		},
		{
			name:  "bare scheme without Authorization prefix",
			input: `using Token 9f8e7d6c5b4a for upstream call`,
			want:  `using Token REDACTED for upstream call`,
		},
		{
			name:  "lowercase scheme",
			input: `authorization: token abc123`,
			want:  `authorization: token REDACTED`,
		},
		{
			name:  "no credential present",
			input: `GET /api/v2/findings/ 200 OK`,
			want:  `GET /api/v2/findings/ 200 OK`,
		},
		{
			name:  "multiple occurrences",
			input: `req Authorization: Token aaa111; retry Authorization: Token bbb222`,
			want:  `req Authorization: Token REDACTED; retry Authorization: Token REDACTED`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Redact(tt.input)
			if got != tt.want {
				t.Errorf("Redact(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestRedact_NeverLeaksSecret(t *testing.T) {
	const secret = "sk_live_super_secret_value_1234567890"
	input := "Authorization: Token " + secret
	got := Redact(input)
	if got == input {
		t.Fatal("Redact() did not modify input containing a token")
	}
	for i := 0; i+8 <= len(secret); i += 8 {
		chunk := secret[i : i+8]
		if contains(got, chunk) {
			t.Fatalf("Redact() output %q still contains secret chunk %q", got, chunk)
		}
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
