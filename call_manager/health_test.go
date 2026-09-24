package call_manager

import (
	"context"
	"errors"
	"testing"

	"github.com/webitel/engine/pkg/discovery"
)

type fakeConn struct {
	name  string
	ready bool
}

func (f *fakeConn) Name() string { return f.name }
func (f *fakeConn) Ready() bool  { return f.ready }
func (f *fakeConn) Close() error { return nil }

func newManager(conns ...discovery.Connection) *callManager {
	pool := discovery.NewPoolConnections()
	for _, c := range conns {
		pool.Append(c)
	}

	return &callManager{poolConnections: pool}
}

func TestReady(t *testing.T) {
	tests := []struct {
		name  string
		conns []discovery.Connection
		want  error
	}{
		{
			name:  "no connection registered",
			conns: nil,
			want:  ErrNoConnection,
		},
		{
			name:  "every connection down",
			conns: []discovery.Connection{&fakeConn{name: "a"}, &fakeConn{name: "b"}},
			want:  ErrNotReady,
		},
		{
			name:  "one of two up",
			conns: []discovery.Connection{&fakeConn{name: "a"}, &fakeConn{name: "b", ready: true}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := newManager(tt.conns...).Ready(context.Background()); !errors.Is(err, tt.want) {
				t.Errorf("Ready() = %v, want %v", err, tt.want)
			}
		})
	}
}

// The check runs on a timer, so it must not consume a round-robin slot.
func TestReadyDoesNotAdvanceRoundRobin(t *testing.T) {
	pick := func(probe bool) []string {
		cm := newManager(
			&fakeConn{name: "a", ready: true},
			&fakeConn{name: "b", ready: true},
			&fakeConn{name: "c", ready: true},
		)

		got := make([]string, 0, 6)

		for range 6 {
			if probe {
				if err := cm.Ready(context.Background()); err != nil {
					t.Fatalf("Ready() = %v, want nil", err)
				}
			}

			cli, err := cm.poolConnections.Get(discovery.StrategyRoundRobin)
			if err != nil {
				t.Fatalf("Get() = %v", err)
			}

			got = append(got, cli.Name())
		}

		return got
	}

	undisturbed, probed := pick(false), pick(true)
	for i := range undisturbed {
		if undisturbed[i] != probed[i] {
			t.Fatalf("probing shifted the round-robin marker: %v, want %v", probed, undisturbed)
		}
	}
}

// A typed nil reads as non-nil, and the Consul TTL updater calls .Error() on it.
func TestReadyReturnsNoTypedNil(t *testing.T) {
	cm := newManager(&fakeConn{name: "a", ready: true})

	if err := cm.Ready(context.Background()); err != nil {
		t.Errorf("Ready() = %#v, want an untyped nil", err)
	}
}
