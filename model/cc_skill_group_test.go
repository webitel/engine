package model

import "testing"

func TestSkillGroupIsValid(t *testing.T) {
	tests := []struct {
		name    string
		group   SkillGroup
		wantErr bool
	}{
		{name: "valid", group: SkillGroup{Name: "Night shift"}, wantErr: false},
		{name: "empty name", group: SkillGroup{Name: ""}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.group.IsValid()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestSkillGroupSkillIsValid(t *testing.T) {
	skill := &Lookup{Id: 5}
	tests := []struct {
		name    string
		item    SkillGroupSkill
		wantErr bool
	}{
		{name: "valid", item: SkillGroupSkill{Skill: skill, Capacity: NewInt(10)}, wantErr: false},
		{name: "min capacity", item: SkillGroupSkill{Skill: skill, Capacity: NewInt(0)}, wantErr: false},
		{name: "max capacity", item: SkillGroupSkill{Skill: skill, Capacity: NewInt(100)}, wantErr: false},
		{name: "no skill", item: SkillGroupSkill{Skill: &Lookup{}, Capacity: NewInt(10)}, wantErr: true},
		{name: "nil capacity", item: SkillGroupSkill{Skill: skill, Capacity: nil}, wantErr: true},
		{name: "capacity below range", item: SkillGroupSkill{Skill: skill, Capacity: NewInt(-1)}, wantErr: true},
		{name: "capacity above range", item: SkillGroupSkill{Skill: skill, Capacity: NewInt(101)}, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.item.IsValid()
			if tc.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
