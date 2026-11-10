package app

import (
	"context"
	"testing"

	"github.com/webitel/wlog"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type fakeDomainDefaultsStore struct {
	store.Store

	skills   *fakeOnlineSkillsStore
	settings *fakeSystemSettingsStore
}

func (f *fakeDomainDefaultsStore) OnlineSkills() store.OnlineSkillsStore {
	return f.skills
}

func (f *fakeDomainDefaultsStore) SystemSettings() store.SystemSettingsStore {
	return f.settings
}

type fakeOnlineSkillsStore struct {
	store.OnlineSkillsStore

	err      model.AppError
	domainID int64
}

func (f *fakeOnlineSkillsStore) CreateSystem(_ context.Context, domainID int64) model.AppError {
	f.domainID = domainID

	return f.err
}

type fakeSystemSettingsStore struct {
	store.SystemSettingsStore

	err      model.AppError
	domainID int64
}

func (f *fakeSystemSettingsStore) CreateDefaults(_ context.Context, domainID int64) model.AppError {
	f.domainID = domainID

	return f.err
}

func TestHandleDomainEventCreated(t *testing.T) {
	failure := model.NewInternalError("test", "failure")

	cases := []struct {
		name        string
		skillsErr   model.AppError
		settingsErr model.AppError
		wantErr     bool
	}{
		{name: "creates all defaults"},
		{name: "settings created when skills fail", skillsErr: failure, wantErr: true},
		{name: "skills created when settings fail", settingsErr: failure, wantErr: true},
		{name: "both fail", skillsErr: failure, settingsErr: failure, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			skills := &fakeOnlineSkillsStore{err: tc.skillsErr}
			settings := &fakeSystemSettingsStore{err: tc.settingsErr}
			a := &App{
				Store: &fakeDomainDefaultsStore{skills: skills, settings: settings},
				Log:   wlog.NewLogger(&wlog.LoggerConfiguration{}),
			}

			err := a.handleDomainEventCreated(context.Background(), &model.DomainEvent{ID: 7})

			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}

			if skills.domainID != 7 || settings.domainID != 7 {
				t.Fatalf("domain ids: skills=%d settings=%d, want 7", skills.domainID, settings.domainID)
			}
		})
	}
}
