// Copyright 2021 Google LLC
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

// Package knebind provides an Ondatra binding for KNE devices.

package ondatra

import (
	"log"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	gnmiclient "github.com/openconfig/gnmi/client"
	"github.com/openconfig/ondatra/internal/binding"
)

type OTG struct {
	cliApi *binding.OTGClientApi
}

func NewOTG(cliApi *binding.OTGClientApi) *OTG {
	return &OTG{cliApi: cliApi}
}

func (otg *OTG) NewConfig(t *testing.T) gosnappi.Config {
	return otg.cliApi.API().NewConfig()
}

func (otg *OTG) PushConfig(t *testing.T, config gosnappi.Config) {

	config_ports := config.Ports().Items()
	topology_ports := otg.cliApi.Ports()
	if len(topology_ports) < len(config_ports) {
		t.Fatalf("Insufficient resources, total ports in config are %v, total ports in testbed are %v",
			len(config_ports), len(topology_ports))
	}

	log.Println("Setting port name and location ...")
	matched_ports := 0
	for _, port := range config_ports {
		for topo_name, topo_location := range topology_ports {
			if port.Name() == topo_name {
				log.Printf("Setting port Name(Topology): %s, Location: %s, Name(Config) %s",
					topo_name, topo_location, port.Name())
				port.SetLocation(topo_location)
				matched_ports++
			}
		}
	}

	if matched_ports != len(config_ports) {
		log.Printf("Error finding matching ports...")
		for _, port := range config_ports {
			log.Printf("Config port name: %s", port.Name())
		}
		for topo_name, topo_location := range topology_ports {
			log.Printf("Topology port name: %s, location: %s", topo_name, topo_location)
		}
		t.Fatalf("Couldn't set locations for all ports, found match for %v out of total ports  %v",
			matched_ports, len(topology_ports))
	}

	log.Println("Pushing config ...")
	if _, err := otg.cliApi.API().SetConfig(config); err != nil {
		t.Fatal(err)
	}
}

func (otg *OTG) StartProtocols(t *testing.T) {
	log.Println("Start protocols ...")
	state := otg.cliApi.API().NewProtocolState().SetState(gosnappi.ProtocolStateState.START)
	if _, err := otg.cliApi.API().SetProtocolState(state); err != nil {
		t.Fatal(err)
	}
}

func (otg *OTG) StopProtocols(t *testing.T) {
	log.Println("Stop protocols ...")
	state := otg.cliApi.API().NewProtocolState().SetState(gosnappi.ProtocolStateState.STOP)
	if _, err := otg.cliApi.API().SetProtocolState(state); err != nil {
		t.Fatal(err)
	}
}

func (otg *OTG) StartTraffic(t *testing.T) {
	log.Println("Starting transmit ...")
	ts := otg.cliApi.API().NewTransmitState().SetState(gosnappi.TransmitStateState.START)
	if _, err := otg.cliApi.API().SetTransmitState(ts); err != nil {
		t.Fatal(err)
	}
}

func (otg *OTG) StopTraffic(t *testing.T) {
	log.Println("Stopping transmit ...")
	ts := otg.cliApi.API().NewTransmitState().SetState(gosnappi.TransmitStateState.STOP)
	if _, err := otg.cliApi.API().SetTransmitState(ts); err != nil {
		t.Fatal(err)
	}
}

func (otg *OTG) NewGnmiQuery(t *testing.T) *gnmiclient.Query {
	addr := otg.cliApi.Gnmi()
	log.Printf("New GNMI Query @%s", addr)
	query := &gnmiclient.Query{
		Addrs:   []string{addr},
		Timeout: 10 * time.Second,
		TLS:     nil,
		Type:    gnmiclient.Once,
	}
	return query
}
