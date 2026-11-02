package model_test

import (
	"errors"
	"testing"

	"github.com/webitel/engine/model"
)

func TestResourceDisplay_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		display    string
		resourceID int64
		wantErr    error
	}{
		{
			name:       "only_numbers",
			display:    "380991112233",
			resourceID: 1,
		},
		{
			name:       "with_prefix_whitespace",
			display:    "   380991112233",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "with_suffix_whitespace",
			display:    "380991112233    ",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		}, {
			name:       "with_both_whitespaces",
			display:    " 380991112233    ",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "accepted_symbols",
			display:    "380-99_111.22!33",
			resourceID: 1,
		},
		{
			name:       "only_chars",
			display:    "CALLCENTER123",
			resourceID: 1,
		},
		{
			name:       "with_plus",
			display:    "+380991112233",
			resourceID: 1,
		},
		{
			name:       "with_plus_invalid_prefix",
			display:    "++380991112233",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "with_plus_in_the_middle",
			display:    "3+80991112233",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "with_plus_in_the_end",
			display:    "380991112233+",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "with_many_pluses",
			display:    "+380+991112233+",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "mixed_data",
			display:    "abc-123_test(01)",
			resourceID: 1,
		},
		{
			name:       "all_allowed_chars",
			display:    "abcABC123-_.!~*'()",
			resourceID: 1,
		},
		{
			name:       "display_with_whitespaces",
			display:    "380 99 111 22 33",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "display_with_tabulation",
			display:    "380\t991112233",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "display_with_new_row_char",
			display:    "38099\n1112233",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "with_@_char",
			display:    "38099@1112233",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "with_non_english",
			display:    "тест123",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "with_emoji",
			display:    "38099😀111",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "sql_injection",
			display:    "' OR 1=1 --",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
		{
			name:       "xss_payload",
			display:    "<script>alert(1)</script>",
			resourceID: 1,
			wantErr:    model.ValidatePhoneNumberError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := model.NewResourceDisplay(tt.display, tt.resourceID)

			err := d.IsValid()

			switch {
			case tt.wantErr == nil && err != nil:
				t.Fatalf("expected nil error, got %v", err)

			case tt.wantErr != nil && !errors.Is(err, tt.wantErr):
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestResourceDisplay_Prepare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		display    string
		resourceID int64
		want       string
	}{
		{
			name:       "only_numbers",
			display:    "380991112233",
			resourceID: 1,
			want:       "380991112233",
		},
		{
			name:       "with_prefix_whitespace",
			display:    "   380991112233",
			resourceID: 1,
			want:       "380991112233",
		},
		{
			name:       "with_suffix_whitespace",
			display:    "380991112233    ",
			resourceID: 1,
			want:       "380991112233",
		},
		{
			name:       "with_both_whitespaces",
			display:    " 380991112233    ",
			resourceID: 1,
			want:       "380991112233",
		},
		{
			name:       "accepted_symbols",
			display:    "380-99_111.22!33",
			resourceID: 1,
			want:       "380-99_111.22!33",
		},
		{
			name:       "only_chars",
			display:    "CALLCENTER123",
			resourceID: 1,
			want:       "CALLCENTER123",
		},
		{
			name:       "with_plus",
			display:    "+380991112233",
			resourceID: 1,
			want:       "+380991112233",
		},
		{
			name:       "with_plus_invalid_prefix",
			display:    "++380991112233",
			resourceID: 1,
			want:       "++380991112233",
		},
		{
			name:       "mixed_data",
			display:    "abc-123_test(01)",
			resourceID: 1,
			want:       "abc-123_test(01)",
		},
		{
			name:       "all_allowed_chars",
			display:    "abcABC123-_.!~*'()",
			resourceID: 1,
			want:       "abcABC123-_.!~*'()",
		},
		{
			name:       "display_with_whitespaces",
			display:    "380 99 111 22 33",
			resourceID: 1,
			want:       "380991112233",
		},
		{
			name:       "display_with_tabulation",
			display:    "380\t991112233",
			resourceID: 1,
			want:       "380991112233",
		},
		{
			name:       "display_with_new_row_char",
			display:    "38099\n1112233",
			resourceID: 1,
			want:       "380991112233",
		},
		{
			name:       "with_@_char",
			display:    "38099@1112233",
			resourceID: 1,
			want:       "380991112233",
		},
		{
			name:       "with_non_english",
			display:    "тест123",
			resourceID: 1,
			want:       "123",
		},
		{
			name:       "with_emoji",
			display:    "38099😀111",
			resourceID: 1,
			want:       "38099111",
		},
		{
			name:       "sql_injection",
			display:    "' OR 1=1 --",
			resourceID: 1,
			want:       "'OR11--",
		},
		{
			name:       "xss_payload",
			display:    "<script>alert(1)</script>",
			resourceID: 1,
			want:       "scriptalert(1)script",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := model.NewResourceDisplay(tt.display, tt.resourceID)

			d.Prepare()

			if d.Display != tt.want {
				t.Errorf("Prepare() target field Display = %q, want %q", d.Display, tt.want)
			}
		})
	}
}

func TestResourceDisplay_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		display    string
		resourceID int64
		want       model.AppError
	}{
		{
			name:       "valid_only_numbers",
			display:    "380991112233",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_with_prefix_whitespace",
			display:    "   380991112233",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_with_suffix_whitespace",
			display:    "380991112233    ",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_with_both_whitespaces",
			display:    " 380991112233    ",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_accepted_symbols",
			display:    "380-99_111.22!33",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_only_chars",
			display:    "CALLCENTER123",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_with_plus",
			display:    "+380991112233",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_with_plus_invalid_prefix",
			display:    "++380991112233",
			resourceID: 1,
			want:       model.ValidatePhoneNumberError,
		},
		{
			name:       "valid_mixed_data",
			display:    "abc-123_test(01)",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_all_allowed_chars",
			display:    "abcABC123-_.!~*'()",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_display_with_whitespaces",
			display:    "380 99 111 22 33",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_display_with_tabulation",
			display:    "380\t991112233",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_display_with_new_row_char",
			display:    "38099\n1112233",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_with_@_char",
			display:    "38099@1112233",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_with_non_english",
			display:    "тест123",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "valid_with_emoji",
			display:    "38099😀111",
			resourceID: 1,
			want:       nil,
		},
		{
			name:       "invalid_sql_injection_empty_after_clean",
			display:    "'  --",
			resourceID: 1,
			want:       model.ValidatePhoneNumberError,
		},
		{
			name:       "invalid_xss_payload_empty_after_clean",
			display:    "<></>",
			resourceID: 1,
			want:       model.ValidatePhoneNumberError,
		},
		{
			name:       "invalid_empty_string",
			display:    "",
			resourceID: 1,
			want:       model.ValidatePhoneNumberError,
		},
		{
			name:       "invalid_only_spaces",
			display:    "     ",
			resourceID: 1,
			want:       model.ValidatePhoneNumberError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := model.NewResourceDisplay(tt.display, tt.resourceID)
			got := d.Parse()

			if !errors.Is(got, tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}
