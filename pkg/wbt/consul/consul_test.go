package consul

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

type fakeClientConn struct {
	resolver.ClientConn
	states chan resolver.State
}

func (f *fakeClientConn) UpdateState(s resolver.State) error {
	f.states <- s
	return nil
}

type fakeSubConn struct {
	balancer.SubConn
	name string
}

func TestPopulateEndpointsKeepsServerNameEmpty(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cc := &fakeClientConn{states: make(chan resolver.State, 1)}
	input := make(chan []serviceMeta)
	go populateEndpoints(ctx, cc, input)

	input <- []serviceMeta{
		{addr: "10.0.0.2:8080", id: "im-gateway-service@node2"},
		{addr: "10.0.0.1:8080", id: "im-gateway-service@node1"},
	}

	var state resolver.State
	select {
	case state = <-cc.states:
	case <-time.After(time.Second):
		t.Fatal("UpdateState was not called")
	}

	want := map[string]string{
		"10.0.0.1:8080": "im-gateway-service@node1",
		"10.0.0.2:8080": "im-gateway-service@node2",
	}
	if len(state.Addresses) != len(want) {
		t.Fatalf("got %d addresses, want %d", len(state.Addresses), len(want))
	}
	for _, addr := range state.Addresses {
		if addr.ServerName != "" {
			t.Errorf("address %s: ServerName = %q, want empty (it overrides the TLS authority)", addr.Addr, addr.ServerName)
		}
		if got := nodeID(addr); got != want[addr.Addr] {
			t.Errorf("address %s: nodeID = %q, want %q", addr.Addr, got, want[addr.Addr])
		}
	}
}

func TestNodeIDAbsent(t *testing.T) {
	if got := nodeID(resolver.Address{Addr: "10.0.0.1:8080"}); got != "" {
		t.Errorf("nodeID = %q, want empty", got)
	}
}

func TestAddressEqualityTracksNodeID(t *testing.T) {
	a := withNodeID(resolver.Address{Addr: "10.0.0.1:8080"}, "svc@node1")
	b := withNodeID(resolver.Address{Addr: "10.0.0.1:8080"}, "svc@node2")
	if a.Equal(b) {
		t.Error("addresses with different node IDs must not be equal")
	}
	if !a.Equal(withNodeID(resolver.Address{Addr: "10.0.0.1:8080"}, "svc@node1")) {
		t.Error("addresses with the same node ID must be equal")
	}
}

func buildPicker(t *testing.T) (balancer.Picker, map[string]balancer.SubConn) {
	t.Helper()
	scs := map[string]balancer.SubConn{
		"svc@node1": &fakeSubConn{name: "node1"},
		"svc@node2": &fakeSubConn{name: "node2"},
	}
	ready := map[balancer.SubConn]base.SubConnInfo{
		scs["svc@node1"]: {Address: withNodeID(resolver.Address{Addr: "10.0.0.1:8080"}, "svc@node1")},
		scs["svc@node2"]: {Address: withNodeID(resolver.Address{Addr: "10.0.0.2:8080"}, "svc@node2")},
	}
	return (&rrPickerBuilder{}).Build(base.PickerBuildInfo{ReadySCs: ready}), scs
}

func TestPickStaticHost(t *testing.T) {
	picker, scs := buildPicker(t)

	for id, want := range scs {
		ctx := context.WithValue(context.Background(), StaticHostKey, StaticHost{Name: id})
		for i := 0; i < 3; i++ {
			res, err := picker.Pick(balancer.PickInfo{Ctx: ctx})
			if err != nil {
				t.Fatalf("Pick(%s): %v", id, err)
			}
			if res.SubConn != want {
				t.Fatalf("Pick(%s) = %v, want %v", id, res.SubConn, want)
			}
		}
	}
}

func TestPickStaticHostUnknown(t *testing.T) {
	picker, _ := buildPicker(t)

	ctx := context.WithValue(context.Background(), StaticHostKey, StaticHost{Name: "svc@missing"})
	if _, err := picker.Pick(balancer.PickInfo{Ctx: ctx}); err == nil {
		t.Fatal("Pick with unknown static host: expected error")
	}
}

func TestPickRoundRobin(t *testing.T) {
	picker, scs := buildPicker(t)

	seen := map[balancer.SubConn]int{}
	for i := 0; i < 4; i++ {
		res, err := picker.Pick(balancer.PickInfo{Ctx: context.Background()})
		if err != nil {
			t.Fatalf("Pick: %v", err)
		}
		seen[res.SubConn]++
	}
	for id, sc := range scs {
		if seen[sc] != 2 {
			t.Errorf("%s picked %d times, want 2", id, seen[sc])
		}
	}
}
