// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package fakebind implements a fake testbed binding, backed by a fake SSH server.
package fakebind

import (
	"time"

	"golang.org/x/net/context"

	log "github.com/golang/glog"
	"github.com/openconfig/ondatra/internal/binding"
	"github.com/openconfig/ondatra/internal/reservation"
	"google.golang.org/grpc"

	gpb "github.com/openconfig/gnmi/proto/gnmi"
	opb "github.com/openconfig/ondatra/proto"
	p4pb "github.com/p4lang/p4runtime/go/p4/v1"
)

var _ binding.Binding = &Binding{}

// Binding is a fake testbed binding comprised of stub implementations.
type Binding struct {
	Reservation      *reservation.Reservation
	ConfigPusher     func(context.Context, *reservation.DUT, string, *binding.ConfigOptions) error
	CLIDialer        func(context.Context, *reservation.DUT, ...grpc.DialOption) (binding.StreamClient, error)
	ConsoleDialer    func(context.Context, *reservation.DUT, ...grpc.DialOption) (binding.StreamClient, error)
	TopologyPusher   func(*reservation.ATE, *opb.Topology) error
	TrafficStarter   func(*reservation.ATE, []*opb.Flow) error
	GNMIDialer       func(context.Context, *reservation.DUT, ...grpc.DialOption) (gpb.GNMIClient, error)
	GNOIDialer       func(context.Context, *reservation.DUT, ...grpc.DialOption) (binding.GNOIClients, error)
	P4RTDialer       func(context.Context, *reservation.DUT, ...grpc.DialOption) (p4pb.P4RuntimeClient, error)
	RoutingRestarter func(*reservation.DUT) error
	PortStateSetter  func(*reservation.ATE, string, bool) error

	OTGDialer     func(context.Context) (binding.OTGClientApi, error)
	OTGGNMIDialer func(context.Context, ...grpc.DialOption) (gpb.GNMIClient, error)
}

// Reset zeros out all the stub implementations.
func (b *Binding) Reset() {
	b.Reservation = nil
	b.ConfigPusher = nil
	b.TopologyPusher = nil
	b.TrafficStarter = nil
	b.CLIDialer = nil
	b.ConsoleDialer = nil
	b.GNMIDialer = nil
	b.GNOIDialer = nil
	b.P4RTDialer = nil
	b.RoutingRestarter = nil
	b.PortStateSetter = nil

	b.OTGDialer = nil
	b.OTGGNMIDialer = nil
}

// Reserve reserves a new fake testbed, reading the definition from the given path.
// If the path is a plain filename, interprets it relative to the target directory.
func (b *Binding) Reserve(_ context.Context, _ *opb.Testbed, _, _ time.Duration) (*reservation.Reservation, error) {
	return b.Reservation, nil
}

// Release is a noop.
func (b *Binding) Release(context.Context) (rerr error) {
	return nil
}

// DialATEGNMI is a noop.
func (b *Binding) DialATEGNMI(ctx context.Context, ate *reservation.ATE, opts ...grpc.DialOption) (gpb.GNMIClient, error) {
	return nil, nil
}

// PushConfig delegates to b.ConfigPusher.
func (b *Binding) PushConfig(ctx context.Context, dut *reservation.DUT, config string, opts *binding.ConfigOptions) error {
	return b.ConfigPusher(ctx, dut, config, opts)
}

// PushTopology delegates to b.TopologyPusher.
func (b *Binding) PushTopology(ate *reservation.ATE, top *opb.Topology) error {
	return b.TopologyPusher(ate, top)
}

// UpdateTopology delegates to b.TopologyPusher.
func (b *Binding) UpdateTopology(ate *reservation.ATE, top *opb.Topology) error {
	return b.TopologyPusher(ate, top)
}

// UpdateBGPPeerStates is a noop.
// TODO: Remove this method once new Ixia config binding is used.
func (b *Binding) UpdateBGPPeerStates(ate *reservation.ATE, interfaces []*opb.InterfaceConfig) error {
	return nil
}

// StartProtocols is a noop.
func (b *Binding) StartProtocols(ate *reservation.ATE) error {
	return nil
}

// StopProtocols is a noop.
func (b *Binding) StopProtocols(ate *reservation.ATE) error {
	return nil
}

// StartTraffic delegates to b.TrafficStarter.
func (b *Binding) StartTraffic(ate *reservation.ATE, fs []*opb.Flow) error {
	return b.TrafficStarter(ate, fs)
}

// UpdateTraffic delegates to b.TrafficStarter.
func (b *Binding) UpdateTraffic(ate *reservation.ATE, fs []*opb.Flow) error {
	return b.TrafficStarter(ate, fs)
}

// StopTraffic is a noop.
func (b *Binding) StopTraffic(ate *reservation.ATE) error {
	return nil
}

// DialGNMI creates a client connection to the fake GNMI server.
func (b *Binding) DialGNMI(ctx context.Context, dut *reservation.DUT, opts ...grpc.DialOption) (gpb.GNMIClient, error) {
	return b.GNMIDialer(ctx, dut, opts...)
}

// DialGNOI creates a client connection to the fake GNOI server.
func (b *Binding) DialGNOI(ctx context.Context, dut *reservation.DUT, opts ...grpc.DialOption) (binding.GNOIClients, error) {
	return b.GNOIDialer(ctx, dut, opts...)
}

// DialP4RT creates a client connection to the fake P4RT server.
func (b *Binding) DialP4RT(ctx context.Context, dut *reservation.DUT, opts ...grpc.DialOption) (p4pb.P4RuntimeClient, error) {
	return b.P4RTDialer(ctx, dut, opts...)
}

// DialCLI creates a client connection to the fake CLI server.
func (b *Binding) DialCLI(ctx context.Context, dut *reservation.DUT, opts ...grpc.DialOption) (binding.StreamClient, error) {
	return b.CLIDialer(ctx, dut, opts...)
}

// DialConsole creates a client connection to the fake Console server.
func (b *Binding) DialConsole(ctx context.Context, dut *reservation.DUT, opts ...grpc.DialOption) (binding.StreamClient, error) {
	return b.ConsoleDialer(ctx, dut, opts...)
}

// DialOTG creates a client connection to the fake OTG server.
func (b *Binding) DialOTG(ctx context.Context) (binding.OTGClientApi, error) {
	return b.OTGDialer(ctx)
}

// DialOTGGNMI creates a client connection to the fake GNMI server.
func (b *Binding) DialOTGGNMI(ctx context.Context, opts ...grpc.DialOption) (gpb.GNMIClient, error) {
	return b.OTGGNMIDialer(ctx, opts...)
}

// HandleInfraFail logs the error and returns it unchanged.
func (b *Binding) HandleInfraFail(err error) error {
	log.Errorf("Infrastructure failure: %v", err)
	return err
}

// SetATEPortState delegates to b.PortStateSetter.
func (b *Binding) SetATEPortState(ate *reservation.ATE, port string, enabled bool) error {
	return b.PortStateSetter(ate, port, enabled)
}

// SetTestMetadata is a noop.
func (b *Binding) SetTestMetadata(md *binding.TestMetadata) error {
	return nil
}

// RestartRouting delegates to b.RoutingRestarter.
func (b *Binding) RestartRouting(dut *reservation.DUT) error {
	return b.RoutingRestarter(dut)
}
