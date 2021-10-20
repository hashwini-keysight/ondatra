package ondatra

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	log "github.com/golang/glog"
	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/openconfig/ondatra/internal/binding"
	"github.com/openconfig/ondatra/internal/reservation"
	"github.com/openconfig/ondatra/knebind"
	opb "github.com/openconfig/ondatra/proto"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	CONTROLLER_FAKE_SERVER = "localhost:40051"
)

var (
	mutex   sync.Mutex
	kenBind *knebind.Bind
)

func initKneBind(kneconfig string, testbedconfig string) (*reservation.Reservation, error) {
	cfg, err := knebind.ParseConfigFile(kneconfig)
	if err != nil {
		return nil, errors.Errorf("Error in reading kne config, file: %v, err: %v", kneconfig, err)
	}
	tb := &opb.Testbed{}
	s, err := ioutil.ReadFile(testbedconfig)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read testbed proto %s", testbedconfig)
	}
	err = prototext.Unmarshal(s, tb)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse testbed proto %s", testbedconfig)
	}
	kne, err := knebind.New(cfg)
	if err != nil {
		return nil, errors.Errorf("New failed: %v", err)
	}
	kenRes, err := kne.Reserve(context.Background(), tb, time.Minute, time.Minute)
	if err != nil {
		return nil, errors.Errorf("Reserve() got error: %v", err)
	}
	if kenRes.ID == "" {
		return nil, errors.Errorf("Reserve() got reservation missing ID: %v", kenRes)
	}

	mutex.Lock()
	defer mutex.Unlock()
	if kenBind == nil {
		kenBind = kne
		binding.Init(kenBind)
	}
	return kenRes, nil
}

func initOTGFakes(t *testing.T) (*reservation.Reservation, error) {
	t.Helper()
	initFakeBinding(t)

	fakeBind.OTGDialer = func(ctx context.Context) (binding.OTGClientApi, error) {
		log.Infof("Dialing GRPC server %s", CONTROLLER_FAKE_SERVER)
		api := gosnappi.NewApi()
		api.NewHttpTransport().SetLocation(CONTROLLER_FAKE_SERVER).SetVerify(false)
		ports := make(map[string]string)
		ports["p1"] = "location1"
		ports["p2"] = "location2"
		ports["p3"] = "location3"

		client := knebind.NewOTGClient(api, CONTROLLER_FAKE_SERVER, "", ports)
		return client, nil

	}

	fmt.Print("Starting mock gRPC server for gosnappi ...\n")
	if err := gosnappi.StartMockGrpcServer(CONTROLLER_FAKE_SERVER); err != nil {
		return nil, errors.Wrapf(err, "Could not start gosnappi mock server %s", CONTROLLER_FAKE_SERVER)
	}
	reserveFakeTestbed(t)
	return fakeBind.Reservation, nil
}
