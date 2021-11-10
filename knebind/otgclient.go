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
package knebind

import (
	"log"
	"testing"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/openconfig/ondatra/internal/binding"
)

type OTG struct {
	cliApi *binding.OTGClientApi
}

func NewOTG(cliApi *binding.OTGClientApi) *OTG {
	return &OTG{cliApi: cliApi}
}

func (otg *OTG) API() gosnappi.GosnappiApi {
	return otg.cliApi.API()
}

func (otg *OTG) Controller() string {
	return otg.cliApi.Controller()
}

func (otg *OTG) Gnmi() string {
	return otg.cliApi.Gnmi()
}

func (otg *OTG) Ports() map[string]string {
	return otg.cliApi.Ports()
}

func (otg *OTG) NewConfig(t *testing.T) gosnappi.Config {
	return otg.cliApi.API().NewConfig()
}

func (otg *OTG) PushConfig(t *testing.T, config gosnappi.Config) {
	log.Println("Pushing config ...")
	if _, err := otg.cliApi.API().SetConfig(config); err != nil {
		t.Fatal(err)
	}
}

func (otg *OTG) StartProtocols(t *testing.T) {
	log.Println("Start protocols ...")
	state := otg.API().NewProtocolState().SetState(gosnappi.ProtocolStateState.START)
	if _, err := otg.API().SetProtocolState(state); err != nil {
		t.Fatal(err)
	}
}

func (otg *OTG) StopProtocols(t *testing.T) {
	log.Println("Stop protocols ...")
	state := otg.API().NewProtocolState().SetState(gosnappi.ProtocolStateState.STOP)
	if _, err := otg.API().SetProtocolState(state); err != nil {
		t.Fatal(err)
	}
}

func (otg *OTG) StartTraffic(t *testing.T) {
	log.Println("Starting transmit ...")
	ts := otg.API().NewTransmitState().SetState(gosnappi.TransmitStateState.START)
	if _, err := otg.API().SetTransmitState(ts); err != nil {
		t.Fatal(err)
	}
}

func (otg *OTG) StopTraffic(t *testing.T) {
	log.Println("Stopping transmit ...")
	ts := otg.API().NewTransmitState().SetState(gosnappi.TransmitStateState.STOP)
	if _, err := otg.API().SetTransmitState(ts); err != nil {
		t.Fatal(err)
	}
}
