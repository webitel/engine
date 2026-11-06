package fsvar_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/webitel/engine/pkg/fsvar"
)

func TestValidateNumber(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr error
		input   string
	}{{
		name:    "number without country code should return success",
		wantErr: nil,
		input:   "0960299201",
	}, {
		name:    "number without plus before country code should return success",
		wantErr: nil,
		input:   "380960299201",
	}, {
		name:    "internal extension number should return success",
		wantErr: nil,
		input:   "2408",
	}, {
		name:    "loopback echo number should return success",
		wantErr: nil,
		input:   "00",
	}, {
		name:    "dialplan number should return success",
		wantErr: nil,
		input:   "illia",
	}, {
		name:    "empty number should return error",
		wantErr: fsvar.ErrBadValue,
		input:   "",
	}, {
		name:    "invalid number format should return error",
		wantErr: fsvar.ErrBadValue,
		input:   "3993%?/8348",
	}, {
		name:    "freeswitch injection should return error",
		wantErr: fsvar.ErrBadValue,
		input:   "{usr_probe='${hostname}'}loopback/9999 &park",
	}, {
		name:    "all allowed special characters should return success",
		wantErr: nil,
		input:   "+380-96_02.99@test#*123",
	}, {
		name:    "slash character should return error (no raw routes allowed)",
		wantErr: fsvar.ErrBadValue,
		input:   "gateway/0960299201",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := fsvar.ValidateNumber(tt.input); !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateNumber(%q) want %v got %v", tt.name, tt.wantErr, err)
			}
		})
	}
}

func TestValidateUserVariableName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "Valid standard name", input: "john_doe", wantErr: nil},
		{name: "Valid with symbols", input: "User.123-test", wantErr: nil},
		{name: "Valid min length", input: "a", wantErr: nil},
		{name: "Valid max length", input: strings.Repeat("a", 64), wantErr: nil},
		{name: "Valid SIP exception", input: "sip_h_x-custom", wantErr: nil},
		{name: "Empty string", input: "", wantErr: fsvar.ErrBadName},
		{name: "Invalid space", input: "user name", wantErr: fsvar.ErrBadName},
		{name: "Invalid symbol @", input: "user@name", wantErr: fsvar.ErrBadName},
		{name: "Cyrillic characters", input: "користувач", wantErr: fsvar.ErrBadName},
		{name: "Exact deny match lowercase", input: "export_vars", wantErr: fsvar.ErrDenied},
		{name: "Exact deny match uppercase", input: "PARK_AFTER_BRIDGE", wantErr: fsvar.ErrDenied},
		{name: "Exact deny match mixed", input: "PaRk_AfTeR_BrIDGE", wantErr: fsvar.ErrDenied},
		{name: "Prefix deny match api_", input: "api_call_bridge", wantErr: fsvar.ErrDenied},
		{name: "Prefix deny match uppercase API_", input: "API_call_bridge", wantErr: fsvar.ErrDenied},
		{name: "Prefix deny match origination_", input: "origination_var", wantErr: fsvar.ErrDenied},
		{name: "SIP blocked regular prefix", input: "sip_user", wantErr: fsvar.ErrDenied},
		{name: "SIP blocked uppercase prefix", input: "SIP_CALL", wantErr: fsvar.ErrDenied},
		{name: "SIP allowed exception", input: "sip_h_x-audio", wantErr: nil},
		{name: "Non-SIP starting with sips", input: "sips_secure", wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := fsvar.ValidateUserVariableName(tt.input); !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateUserName(%q) error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{name: "Valid ASCII text", input: "Hello, World!", wantErr: nil},
		{name: "Valid UTF-8 Cyrillic", input: "Привіт світ 123!", wantErr: nil},
		{name: "Empty string", input: "", wantErr: nil},
		{name: "Valid spaces included", input: "word1 word2 word3", wantErr: nil},
		{name: "Invalid UTF-8 sequence", input: "\xff\xfe", wantErr: fsvar.ErrBadValue},
		{name: "Contains newline (\\n)", input: "line1\nline2", wantErr: fsvar.ErrBadValue},
		{name: "Contains tab (\\t)", input: "col1\tcol2", wantErr: fsvar.ErrBadValue},
		{name: "Contains carriage return (\\r)", input: "text\rline", wantErr: fsvar.ErrBadValue},
		{name: "Contains null byte", input: "text\x00byte", wantErr: fsvar.ErrBadValue},
		{name: "Contains DEL character", input: "text\x7fchar", wantErr: fsvar.ErrBadValue},
		{name: "Contains unit separator (0x1f)", input: "text\x1fchar", wantErr: fsvar.ErrBadValue},
		{name: "Contains template injection syntax (${)", input: "Hello ${USER}", wantErr: fsvar.ErrBadValue},
		{name: "Contains backslash (\\)", input: "C:\\path\\file", wantErr: fsvar.ErrBadValue},
		{name: "Edge case: standalone dollar and brace", input: "$ {not_a_template}", wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if err := fsvar.ValidateValue(tt.input); !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateValue(%q) error = %v, wanted %v", tt.name, err, tt.wantErr)
			}
		})
	}
}

func TestQuote(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantErr   error
		wantQuote string
	}{
		{
			name:      "Empty string",
			input:     "",
			wantQuote: "''",
			wantErr:   nil,
		},
		{
			name:      "Standard plain text",
			input:     "hello_world",
			wantQuote: "'hello_world'",
			wantErr:   nil,
		},
		{
			name:      "String with single quote (English contraction)",
			input:     "it's working",
			wantQuote: `'it\'s working'`,
			wantErr:   nil,
		},
		{
			name:      "Multiple consecutive quotes",
			input:     "a'''b",
			wantQuote: `'a\'\'\'b'`,
			wantErr:   nil,
		},
		{
			name:      "Quotes at boundaries",
			input:     "'value'",
			wantQuote: "'\\'value\\''",
			wantErr:   nil,
		},
		{
			name:      "Validation fails: contains backslash",
			input:     "path\\to\\file",
			wantQuote: "",
			wantErr:   fsvar.ErrBadValue,
		},
		{
			name:      "Validation fails: contains template injection",
			input:     "user_${ID}",
			wantQuote: "",
			wantErr:   fsvar.ErrBadValue,
		},
		{
			name:      "Validation fails: control character (newline)",
			input:     "line1\nline2",
			wantQuote: "",
			wantErr:   fsvar.ErrBadValue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := fsvar.Quote(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Quote(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.wantQuote {
				t.Errorf("Quote(%q) = %q, want %q", tt.input, got, tt.wantQuote)
			}
		})
	}
}
