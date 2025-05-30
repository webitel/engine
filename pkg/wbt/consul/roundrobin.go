/*
 *
 * Copyright 2017 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package roundrobin defines a roundrobin balancer. Roundrobin balancer is
// installed as one of the default balancers in gRPC, users don't need to
// explicitly install this balancer.
package consul

import (
	"errors"
	"math/rand"
	"sync/atomic"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
)

// Name is the name of round_robin balancer.
const Name = "wbt_round_robin"

var logger = grpclog.Component("wbt_round_robin")

// newBuilder creates a new roundrobin balancer builder.
func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, &rrPickerBuilder{}, base.Config{HealthCheck: true})
}

func init() {
	balancer.Register(newBuilder())
}

type rrPickerBuilder struct{}

func (*rrPickerBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	logger.Infof("roundrobinPicker: Build called with info: %v", info)
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}
	scs := make([]balancer.SubConn, 0, len(info.ReadySCs))
	subIdIndex := make(map[string]balancer.SubConn)
	for sc, inf := range info.ReadySCs {
		scs = append(scs, sc)
		if inf.Address.ServerName != "" {
			subIdIndex[inf.Address.ServerName] = sc
		}
	}
	return &rrPicker{
		subConns:   scs,
		subIdIndex: subIdIndex,
		// Start at a random index, as the same RR balancer rebuilds a new
		// picker when SubConn states change, and we don't want to apply excess
		// load to the first server in the list.
		next: uint32(rand.Intn(len(scs))),
	}
}

type rrPicker struct {
	// subConns is the snapshot of the roundrobin balancer when this picker was
	// created. The slice is immutable. Each Get() will do a round robin
	// selection from it and return the selected SubConn.
	subConns   []balancer.SubConn
	subIdIndex map[string]balancer.SubConn

	next uint32
}

type StaticHost struct {
	Name string
}

var StaticHostKey struct{}

func (p *rrPicker) Pick(r balancer.PickInfo) (balancer.PickResult, error) {
	v := r.Ctx.Value(StaticHostKey)
	if v != nil {
		if sc, ok := p.subIdIndex[v.(StaticHost).Name]; ok {
			return balancer.PickResult{SubConn: sc}, nil
		}
		return balancer.PickResult{}, errors.New("no such host")
	}

	subConnsLen := uint32(len(p.subConns))
	nextIndex := atomic.AddUint32(&p.next, 1)

	sc := p.subConns[nextIndex%subConnsLen]
	return balancer.PickResult{SubConn: sc}, nil
}
