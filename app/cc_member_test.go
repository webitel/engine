package app

import (
	"context"
	"slices"
	"testing"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type fakeCommTypeStore struct {
	store.CommunicationTypeStore

	def          *model.CommunicationType
	defErr       model.AppError
	types        map[uint32]*model.CommunicationType
	defaultCalls int
	lastChannel  string
	fetchedIDs   []uint32
}

func (f *fakeCommTypeStore) GetDefault(_ context.Context, _ int64, channel string) (*model.CommunicationType, model.AppError) {
	f.defaultCalls++
	f.lastChannel = channel

	return f.def, f.defErr
}

func (f *fakeCommTypeStore) GetAllPage(_ context.Context, _ int64, search *model.SearchCommunicationType) ([]*model.CommunicationType, model.AppError) {
	f.fetchedIDs = append(f.fetchedIDs, search.Ids...)

	out := make([]*model.CommunicationType, 0, len(search.Ids))
	for _, id := range search.Ids {
		if t, ok := f.types[id]; ok {
			out = append(out, t)
		}
	}

	return out, nil
}

type fakeStore struct {
	store.Store

	comm store.CommunicationTypeStore
}

func (f *fakeStore) CommunicationType() store.CommunicationTypeStore { return f.comm }

func comms(types ...int) []*model.MemberCommunication {
	out := make([]*model.MemberCommunication, 0, len(types))
	for _, id := range types {
		if id < 0 {
			out = append(out, nil)

			continue
		}

		out = append(out, &model.MemberCommunication{Destination: "380000", Type: model.Lookup{Id: id}})
	}

	return out
}

func TestApp_resolveCommunicationTypeChannels(t *testing.T) {
	mobile := &model.CommunicationType{Id: 42, Name: "Mobile", Channel: model.CommunicationChannelPhone}
	mail := &model.CommunicationType{Id: 5, Name: "Mail", Channel: model.CommunicationChannelEmail}

	tests := []struct {
		name             string
		fillDefaults     bool
		def              *model.CommunicationType
		defErr           model.AppError
		communications   []*model.MemberCommunication
		wantErrID        string
		wantDefaultCalls int
		wantFetchedIDs   []uint32
		wantChannels     map[int]string
		wantTypeIDs      []int
	}{
		{
			name:             "type-less comm filled with phone default, default not refetched",
			fillDefaults:     true,
			def:              mobile,
			communications:   comms(0),
			wantDefaultCalls: 1,
			wantFetchedIDs:   nil,
			wantChannels:     map[int]string{42: model.CommunicationChannelPhone},
			wantTypeIDs:      []int{42},
		},
		{
			name:             "typed and type-less mixed, only typed ids fetched",
			fillDefaults:     true,
			def:              mobile,
			communications:   comms(5, 0, -1, 0, 5),
			wantDefaultCalls: 1,
			wantFetchedIDs:   []uint32{5},
			wantChannels:     map[int]string{42: model.CommunicationChannelPhone, 5: model.CommunicationChannelEmail},
			wantTypeIDs:      []int{5, 42, -1, 42, 5},
		},
		{
			name:             "explicit default id not refetched after fill",
			fillDefaults:     true,
			def:              mobile,
			communications:   comms(0, 42),
			wantDefaultCalls: 1,
			wantFetchedIDs:   nil,
			wantChannels:     map[int]string{42: model.CommunicationChannelPhone},
			wantTypeIDs:      []int{42, 42},
		},
		{
			name:             "fill disabled leaves type-less comm untouched",
			fillDefaults:     false,
			def:              mobile,
			communications:   comms(0, 5),
			wantDefaultCalls: 0,
			wantFetchedIDs:   []uint32{5},
			wantChannels:     map[int]string{5: model.CommunicationChannelEmail},
			wantTypeIDs:      []int{0, 5},
		},
		{
			name:             "no default configured -> not_found",
			fillDefaults:     true,
			defErr:           model.NewNotFoundError("store.sql_communication_type.get_default.app_error", "no rows"),
			communications:   comms(0),
			wantErrID:        "app.member.communications.default_type.not_found",
			wantDefaultCalls: 1,
		},
		{
			name:             "store failure is returned as is",
			fillDefaults:     true,
			defErr:           model.NewInternalError("store.sql_communication_type.get_default.app_error", "conn refused"),
			communications:   comms(0),
			wantErrID:        "store.sql_communication_type.get_default.app_error",
			wantDefaultCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCommTypeStore{def: tt.def, defErr: tt.defErr, types: map[uint32]*model.CommunicationType{5: mail, 42: mobile}}
			app := &App{Store: &fakeStore{comm: fake}}

			got, err := app.resolveCommunicationTypeChannels(context.Background(), 1, tt.communications, tt.fillDefaults)
			if tt.wantErrID != "" {
				if err == nil || err.GetId() != tt.wantErrID {
					t.Fatalf("err = %v, want id %v", err, tt.wantErrID)
				}
			} else if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}

			if fake.defaultCalls != tt.wantDefaultCalls {
				t.Errorf("GetDefault calls = %d, want %d", fake.defaultCalls, tt.wantDefaultCalls)
			}

			if tt.wantDefaultCalls > 0 && fake.lastChannel != model.CommunicationChannelPhone {
				t.Errorf("default channel = %q, want %q", fake.lastChannel, model.CommunicationChannelPhone)
			}

			if !slices.Equal(fake.fetchedIDs, tt.wantFetchedIDs) {
				t.Errorf("fetched ids = %v, want %v", fake.fetchedIDs, tt.wantFetchedIDs)
			}

			if tt.wantErrID != "" {
				return
			}

			if len(got) != len(tt.wantChannels) {
				t.Errorf("channels = %v, want %v", got, tt.wantChannels)
			}

			for id, ch := range tt.wantChannels {
				if got[id] != ch {
					t.Errorf("channel[%d] = %q, want %q", id, got[id], ch)
				}
			}

			for i, c := range tt.communications {
				want := tt.wantTypeIDs[i]
				if c == nil {
					if want != -1 {
						t.Errorf("comm %d is nil, want type %d", i, want)
					}

					continue
				}

				if c.Type.Id != want {
					t.Errorf("comm %d type = %d, want %d", i, c.Type.Id, want)
				}
			}
		})
	}
}
