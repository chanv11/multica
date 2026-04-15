package daemon

import (
	"encoding/json"
	"testing"
)

func normalizeJSON(t *testing.T, s string) string {
	t.Helper()
	var v any
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		t.Fatalf("invalid JSON in test: %v\ninput: %s", err, s)
	}
	out, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("re-marshal JSON: %v", err)
	}
	return string(out)
}

func TestResolveMCPVars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  string
		env     map[string]string
		want    string
		wantErr bool
	}{
		{
			name:   "no placeholders",
			config: `{"command":"echo","args":["hello"]}`,
			env:    nil,
			want:   `{"command":"echo","args":["hello"]}`,
		},
		{
			name:   "single var in string",
			config: `{"command":"${MY_CMD}","args":["hello"]}`,
			env:    map[string]string{"MY_CMD": "/usr/bin/echo"},
			want:   `{"command":"/usr/bin/echo","args":["hello"]}`,
		},
		{
			name:   "var in array element",
			config: `{"command":"echo","args":["${API_URL}","--verbose"]}`,
			env:    map[string]string{"API_URL": "https://example.com/api"},
			want:   `{"command":"echo","args":["https://example.com/api","--verbose"]}`,
		},
		{
			name:   "multiple vars in same string",
			config: `{"url":"${HOST}:${PORT}"}`,
			env:    map[string]string{"HOST": "localhost", "PORT": "8080"},
			want:   `{"url":"localhost:8080"}`,
		},
		{
			name:   "var in nested object",
			config: `{"outer":{"inner":"${SECRET}"}}`,
			env:    map[string]string{"SECRET": "abc123"},
			want:   `{"outer":{"inner":"abc123"}}`,
		},
		{
			name:   "non-placeholder string preserved",
			config: `{"command":"echo ${LITERAL} text","key":"no-vars-here"}`,
			env:    map[string]string{"LITERAL": "hello"},
			want:   `{"command":"echo hello text","key":"no-vars-here"}`,
		},
		{
			name:    "missing var returns error",
			config:  `{"command":"${MISSING_VAR}"}`,
			env:     map[string]string{"OTHER": "value"},
			wantErr: true,
		},
		{
			name:    "missing var among multiple returns error",
			config:  `{"a":"${FOUND}","b":"${MISSING}"}`,
			env:     map[string]string{"FOUND": "yes"},
			wantErr: true,
		},
		{
			name:   "empty env with no placeholders",
			config: `{"key":"value"}`,
			env:    nil,
			want:   `{"key":"value"}`,
		},
		{
			name:   "boolean and number values unchanged",
			config: `{"bool":true,"num":42,"str":"${VAR}"}`,
			env:    map[string]string{"VAR": "hello"},
			want:   `{"bool":true,"num":42,"str":"hello"}`,
		},
		{
			name:   "null value unchanged",
			config: `{"key":null,"str":"${VAR}"}`,
			env:    map[string]string{"VAR": "resolved"},
			want:   `{"key":null,"str":"resolved"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveMCPVars([]byte(tt.config), tt.env)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveMCPVars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			// Normalize whitespace for comparison.
			want := normalizeJSON(t, tt.want)
			result := normalizeJSON(t, string(got))
			if result != want {
				t.Errorf("resolveMCPVars()\ngot:  %s\nwant: %s", result, want)
			}
		})
	}
}

func TestResolveMCPVarsMissingVarErrorMessage(t *testing.T) {
	t.Parallel()

	config := `{"command":"${MISSING_VAR}","args":["${ANOTHER_MISSING}"]}`
	_, err := resolveMCPVars([]byte(config), map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing vars")
	}

	errMsg := err.Error()
	if !containsSubstring(errMsg, "MISSING_VAR") {
		t.Errorf("error should mention MISSING_VAR, got: %s", errMsg)
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
