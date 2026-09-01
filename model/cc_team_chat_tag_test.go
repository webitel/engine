package model_test

import (
	"strings"
	"testing"

	"github.com/webitel/engine/model"
)

func TestTeamChatTag_IsValid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		tag     model.TeamChatTag
		wantErr bool
		errID   string
	}{
		{
			name: "valid tag",
			tag: model.TeamChatTag{
				Tag: "support",
			},
			wantErr: false,
		},
		{
			name: "empty tag",
			tag: model.TeamChatTag{
				Tag: "",
			},
			wantErr: true,
			errID:   "model.cc_team_chat_tag.is_valid.tag.app_error",
		},
		{
			name: "tag with surrounding whitespace is trimmed",
			tag: model.TeamChatTag{
				Tag: "  support  ",
			},
			wantErr: false,
		},
		{
			name: "whitespace-only tag is treated as empty",
			tag: model.TeamChatTag{
				Tag: "   ",
			},
			wantErr: true,
			errID:   "model.cc_team_chat_tag.is_valid.tag.app_error",
		},
		{
			name: "tag exceeds max length",
			tag: model.TeamChatTag{
				Tag: "this_is_a_very_long_tag_that_exceeds_the_maximum_allowed_length_of_64_characters",
			},
			wantErr: true,
			errID:   "model.cc_team_chat_tag.is_valid.tag.max_length",
		},
		{
			name: "tag at exactly max length (64 runes)",
			tag: model.TeamChatTag{
				Tag: strings.Repeat("x", 64),
			},
			wantErr: false,
		},
		{
			name: "tag one rune over max length (65 runes)",
			tag: model.TeamChatTag{
				Tag: strings.Repeat("x", 65),
			},
			wantErr: true,
			errID:   "model.cc_team_chat_tag.is_valid.tag.max_length",
		},
		{
			name: "non-ASCII tag at max length is counted in runes, not bytes",
			tag: model.TeamChatTag{
				// Cyrillic characters are 2 bytes each in UTF-8 (128 bytes total),
				// but only 64 runes, so this must be accepted.
				Tag: strings.Repeat("щ", 64),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.tag.IsValid()
			if (err != nil) != tt.wantErr {
				t.Errorf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && err != nil && err.GetId() != tt.errID {
				t.Errorf("IsValid() error ID = %s, want %s", err.GetId(), tt.errID)
			}
		})
	}
}

func TestTeamChatTag_IsValid_TrimsTag(t *testing.T) {
	t.Parallel()

	tag := model.TeamChatTag{Tag: "  support  "}
	if err := tag.IsValid(); err != nil {
		t.Fatalf("IsValid() unexpected error = %v", err)
	}
	if tag.Tag != "support" {
		t.Errorf("IsValid() did not trim tag, got %q, want %q", tag.Tag, "support")
	}
}
