package app

import (
	"context"
	"testing"

	"github.com/webitel/engine/model"
	"github.com/webitel/engine/store"
)

type fakeCommTypeStore struct {
	store.CommunicationTypeStore
	def         *model.CommunicationType
	err         model.AppError
	calls       int
	lastChannel string
}

func (f *fakeCommTypeStore) GetDefault(_ context.Context, _ int64, channel string) (*model.CommunicationType, model.AppError) {
	f.calls++
	f.lastChannel = channel
	return f.def, f.err
}

type fakeStore struct {
	store.Store
	comm store.CommunicationTypeStore
}

func (f *fakeStore) CommunicationType() store.CommunicationTypeStore { return f.comm }

func TestApp_fillDefaultCommunications(t *testing.T) {
	tests := []struct {
		name        string
		queueType   int8
		def         *model.CommunicationType
		members     []*model.Member
		wantErr     bool
		wantErrID   string
		wantCalls   int
		wantChannel string
		checkFilled func(t *testing.T, members []*model.Member)
	}{
		{
			name:        "type-less comm filled with call default",
			queueType:   model.QueueTypeProgressiveCall,
			def:         &model.CommunicationType{Id: 42, Name: "Mobile"},
			members:     []*model.Member{{Communications: []*model.MemberCommunication{{Destination: "380000", Type: model.Lookup{Id: 0}}}}},
			wantCalls:   1,
			wantChannel: model.CommunicationChannelCall,
			checkFilled: func(t *testing.T, members []*model.Member) {
				got := members[0].Communications[0].Type
				if got.Id != 42 || got.Name != "Mobile" {
					t.Errorf("type not filled: got %+v, want {42 Mobile}", got)
				}
			},
		},
		{
			name:        "task queue resolves task channel",
			queueType:   model.QueueTypeOfflineTask,
			def:         &model.CommunicationType{Id: 7, Name: "Job"},
			members:     []*model.Member{{Communications: []*model.MemberCommunication{{Type: model.Lookup{Id: 0}}}}},
			wantCalls:   1,
			wantChannel: model.CommunicationChannelTask,
		},
		{
			name:      "already-typed comm left unchanged, store not queried",
			queueType: model.QueueTypeProgressiveCall,
			def:       &model.CommunicationType{Id: 42, Name: "Mobile"},
			members:   []*model.Member{{Communications: []*model.MemberCommunication{{Type: model.Lookup{Id: 5, Name: "Work"}}}}},
			wantCalls: 0,
			checkFilled: func(t *testing.T, members []*model.Member) {
				if got := members[0].Communications[0].Type; got.Id != 5 {
					t.Errorf("typed comm mutated: got %+v", got)
				}
			},
		},
		{
			name:      "no default configured -> not_found error",
			queueType: model.QueueTypeProgressiveCall,
			def:       nil,
			members:   []*model.Member{{Communications: []*model.MemberCommunication{{Type: model.Lookup{Id: 0}}}}},
			wantErr:   true,
			wantErrID: "app.member.communications.default_type.not_found",
			wantCalls: 1,
		},
		{
			name:      "default resolved once across many members/comms",
			queueType: model.QueueTypeProgressiveCall,
			def:       &model.CommunicationType{Id: 42, Name: "Mobile"},
			members: []*model.Member{
				{Communications: []*model.MemberCommunication{{Type: model.Lookup{Id: 0}}, nil, {Type: model.Lookup{Id: 0}}}},
				{Communications: []*model.MemberCommunication{{Type: model.Lookup{Id: 3}}, {Type: model.Lookup{Id: 0}}}},
			},
			wantCalls: 1,
			checkFilled: func(t *testing.T, members []*model.Member) {
				for _, m := range members {
					for _, c := range m.Communications {
						if c != nil && c.Type.Id == 0 {
							t.Error("a type-less comm was left unfilled")
						}
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := &fakeCommTypeStore{def: tt.def}
			app := &App{Store: &fakeStore{comm: fake}}
			q := &model.Queue{Type: tt.queueType}

			err := app.fillDefaultCommunications(context.Background(), 1, q, tt.members...)

			if (err != nil) != tt.wantErr {
				t.Fatalf("fillDefaultCommunications() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.wantErrID != "" && err.GetId() != tt.wantErrID {
				t.Errorf("err.Id = %v, want %v", err.GetId(), tt.wantErrID)
			}
			if fake.calls != tt.wantCalls {
				t.Errorf("GetDefault calls = %d, want %d", fake.calls, tt.wantCalls)
			}
			if tt.wantChannel != "" && fake.lastChannel != tt.wantChannel {
				t.Errorf("resolved channel = %q, want %q", fake.lastChannel, tt.wantChannel)
			}
			if tt.checkFilled != nil {
				tt.checkFilled(t, tt.members)
			}
		})
	}
}
