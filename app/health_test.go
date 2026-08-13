package app

import (
	"context"
	"testing"

	"github.com/webitel/engine/call_manager"
	"github.com/webitel/engine/model"
)

type fakeCallClient struct {
	call_manager.CallClient

	ready bool
}

func (f *fakeCallClient) Ready() bool { return f.ready }

type fakeCallManager struct {
	call_manager.CallManager

	cli call_manager.CallClient
	err model.AppError
}

func (f *fakeCallManager) CallClient() (call_manager.CallClient, model.AppError) {
	return f.cli, f.err
}

func TestFreeswitchCheck(t *testing.T) {
	tests := []struct {
		name    string
		cm      call_manager.CallManager
		wantErr bool
	}{
		{
			name: "client ready",
			cm:   &fakeCallManager{cli: &fakeCallClient{ready: true}},
		},
		{
			name:    "client not ready",
			cm:      &fakeCallManager{cli: &fakeCallClient{ready: false}},
			wantErr: true,
		},
		{
			name:    "no client available",
			cm:      &fakeCallManager{err: model.NewInternalError("test.no_client", "none")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := freeswitchCheck(tt.cm)(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// A health check must never hand back a nil-valued error interface: the Consul
// TTL updater treats a non-nil error as safe to call .Error() on.
func TestFreeswitchCheckNoTypedNilError(t *testing.T) {
	var nilAppErr model.AppError // nil interface value of a concrete-free type

	cm := &fakeCallManager{cli: &fakeCallClient{ready: true}, err: nilAppErr}

	if err := freeswitchCheck(cm)(context.Background()); err != nil {
		t.Errorf("a nil AppError must not surface as a non-nil error, got %v", err)
	}
}
