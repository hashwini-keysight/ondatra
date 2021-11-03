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
)

type OTGClientApiImpl struct {
	api   gosnappi.GosnappiApi
	grpc  string
	gnmi  string
	ports map[string]string
}

func NewOTGClient(api gosnappi.GosnappiApi, grpc string, gnmi string, ports map[string]string) *OTGClientApiImpl {
	return &OTGClientApiImpl{
		api:   api,
		grpc:  grpc,
		gnmi:  gnmi,
		ports: ports,
	}
}

func (cli *OTGClientApiImpl) API() gosnappi.GosnappiApi {
	return cli.api
}

func (cli *OTGClientApiImpl) Controller() string {
	return cli.grpc
}

func (cli *OTGClientApiImpl) Gnmi() string {
	return cli.gnmi
}

func (cli *OTGClientApiImpl) Ports() map[string]string {
	return cli.ports
}

func (cli *OTGClientApiImpl) NewConfig(t *testing.T) gosnappi.Config {
	return cli.api.NewConfig()
}

func (cli *OTGClientApiImpl) PushConfig(t *testing.T, config gosnappi.Config) {
	log.Println("Pushing config ...")
	if _, err := cli.api.SetConfig(config); err != nil {
		t.Fatal(err)
	}
}

func (cli *OTGClientApiImpl) StartProtocols(t *testing.T) {
	log.Println("Start protocols ...")
	state := cli.api.NewProtocolState().SetState(gosnappi.ProtocolStateState.START)
	if _, err := cli.api.SetProtocolState(state); err != nil {
		t.Fatal(err)
	}
}

func (cli *OTGClientApiImpl) StopProtocols(t *testing.T) {
	log.Println("Stop protocols ...")
	state := cli.api.NewProtocolState().SetState(gosnappi.ProtocolStateState.STOP)
	if _, err := cli.api.SetProtocolState(state); err != nil {
		t.Fatal(err)
	}
}

func (cli *OTGClientApiImpl) StartTraffic(t *testing.T) {
	log.Println("Starting transmit ...")
	ts := cli.api.NewTransmitState().SetState(gosnappi.TransmitStateState.START)
	if _, err := cli.api.SetTransmitState(ts); err != nil {
		t.Fatal(err)
	}
}

func (cli *OTGClientApiImpl) StopTraffic(t *testing.T) {
	log.Println("Stopping transmit ...")
	ts := cli.api.NewTransmitState().SetState(gosnappi.TransmitStateState.STOP)
	if _, err := cli.api.SetTransmitState(ts); err != nil {
		t.Fatal(err)
	}
}
