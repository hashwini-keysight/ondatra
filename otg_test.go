package ondatra

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/open-traffic-generator/snappi/gosnappi"
	"github.com/openconfig/ondatra/knebind"
)

type WaitForOpts struct {
	Condition string
	Interval  time.Duration
	Timeout   time.Duration
}

func TestGoSnappiK8s_001(t *testing.T) {
	t.Log("TestGoSnappiK8s_001 - START ...")
	_, err := initKneBind("otg-kne-001.yaml", "otg-testbed-001.txt")
	if err != nil {
		t.Fatalf("initKneBind() call failed: %v", err)
	}
	if err := reserve("otg-testbed-001.txt", time.Hour, 0); err != nil {
		t.Fatalf("reserve() call failed: %v", err)
	}

	ATEs(t)
	otg := OTG(t)

	defer otg.NewConfig(t)
	defer otg.StopProtocols(t)

	config := PacketForwardBgpv6Config(t, otg)
	otg.PushConfig(t, config)
	otg.StartProtocols(t)
	WaitFor(t, func() (bool, error) { return AllBgp6SessionUp(otg.API(), config) }, nil)
	otg.StartTraffic(t)
	WaitFor(t, func() (bool, error) { return PortAndFlowMetricsOk(otg.API(), config) }, nil)

	t.Log("TestGoSnappiK8s_001 - END ...")
}
func TestGoSnappiK8s_002(t *testing.T) {
	t.Log("TestGoSnappiK8s_002 - START ...")
	_, err := initKneBind("otg-kne-002.yaml", "otg-testbed-002.txt")
	if err != nil {
		t.Fatalf("initKneBind() call failed: %v", err)
	}
	if err := reserve("otg-testbed-002.txt", time.Hour, 0); err != nil {
		t.Fatalf("reserve() call failed: %v", err)
	}

	ATEs(t)
	otg := OTG(t)

	defer otg.NewConfig(t)
	defer otg.StopProtocols(t)

	config := Bgpv4RoutesConfig(t, otg)
	otg.PushConfig(t, config)
	otg.StartProtocols(t)
	WaitFor(t, func() (bool, error) { return AllBgp4SessionUp(otg.API(), config) }, nil)
	otg.StartTraffic(t)
	WaitFor(t, func() (bool, error) { return PortAndFlowMetricsOk(otg.API(), config) }, nil)

	t.Log("TestGoSnappiK8s_002 - END ...")
}

func PacketForwardBgpv6Config(t *testing.T, otg *knebind.OTG) gosnappi.Config {
	config := otg.NewConfig(t)

	// add ports
	var names []string
	var locations []string
	for name, location := range otg.Ports() {
		names = append(names, name)
		locations = append(locations, location)
	}

	p1 := config.Ports().Add().SetName(names[0]).SetLocation(locations[0])
	p2 := config.Ports().Add().SetName(names[1]).SetLocation(locations[1])
	p3 := config.Ports().Add().SetName(names[2]).SetLocation(locations[2])

	// add devices
	d1 := config.Devices().Add().SetName("d1")
	d2 := config.Devices().Add().SetName("d2")
	d3 := config.Devices().Add().SetName("d3")

	// add flows and common properties
	for i := 1; i <= 4; i++ {
		flow := config.Flows().Add()
		flow.Metrics().SetEnable(true)
		flow.Duration().FixedPackets().SetPackets(1000)
		flow.Rate().SetPps(500)
	}

	// add protocol stacks for device d1
	d1Eth1 := d1.Ethernets().
		Add().
		SetName("d1Eth").
		SetPortName(p1.Name()).
		SetMac("00:00:01:01:01:01").
		SetMtu(1500)

	d1Eth1.
		Ipv6Addresses().
		Add().
		SetName("p1d1ipv6").
		SetAddress("0:1:1:1::1").
		SetGateway("0:1:1:1::2").
		SetPrefix(64)

	d1Bgp := d1.Bgp().
		SetRouterId("1.1.1.1")

	d1BgpIpv6Interface1 := d1Bgp.
		Ipv6Interfaces().Add().
		SetIpv6Name("p1d1ipv6")

	d1BgpIpv6Interface1Peer1 := d1BgpIpv6Interface1.
		Peers().
		Add().
		SetAsNumber(1111).
		SetAsType(gosnappi.BgpV6PeerAsType.EBGP).
		SetPeerAddress("0:1:1:1::2").
		SetName("BGPv6 Peer 1")

	d1BgpIpv6Interface1Peer1V6Route1 := d1BgpIpv6Interface1Peer1.
		V6Routes().
		Add().
		SetNextHopIpv6Address("0:1:1:1::1").
		SetName("p1d1peer1rrv6").
		SetNextHopAddressType(gosnappi.BgpV6RouteRangeNextHopAddressType.IPV6).
		SetNextHopMode(gosnappi.BgpV6RouteRangeNextHopMode.MANUAL)

	d1BgpIpv6Interface1Peer1V6Route1.Addresses().Add().
		SetAddress("0:10:10:10::0").
		SetPrefix(64).
		SetCount(2).
		SetStep(2)

	d1BgpIpv6Interface1Peer1V6Route1.Advanced().
		SetMultiExitDiscriminator(50).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	d1BgpIpv6Interface1Peer1V6Route1.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	d1BgpIpv6Interface1Peer1V6Route1AsPath := d1BgpIpv6Interface1Peer1V6Route1.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	d1BgpIpv6Interface1Peer1V6Route1AsPath.Segments().Add().
		SetAsNumbers([]int64{1112, 1113}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	// add protocol stacks for device d2
	d2Eth1 := d2.Ethernets().
		Add().
		SetName("d2Eth").
		SetPortName(p2.Name()).
		SetMac("00:00:02:02:02:02").
		SetMtu(1500)

	d2Eth1.
		Ipv6Addresses().
		Add().
		SetName("p2d1ipv6").
		SetAddress("0:2:2:2::2").
		SetGateway("0:2:2:2::1").
		SetPrefix(64)

	d2Bgp := d2.Bgp().
		SetRouterId("2.2.2.2")

	d2BgpIpv6Interface1 := d2Bgp.
		Ipv6Interfaces().Add().
		SetIpv6Name("p2d1ipv6")

	d2BgpIpv6Interface1Peer1 := d2BgpIpv6Interface1.
		Peers().
		Add().
		SetAsNumber(2222).
		SetAsType(gosnappi.BgpV6PeerAsType.EBGP).
		SetPeerAddress("0:2:2:2::1").
		SetName("BGPv6 Peer 2")

	d2BgpIpv6Interface1Peer1V6Route1 := d2BgpIpv6Interface1Peer1.
		V6Routes().
		Add().
		SetNextHopIpv6Address("0:2:2:2::2").
		SetName("p2d1peer1rrv6").
		SetNextHopAddressType(gosnappi.BgpV6RouteRangeNextHopAddressType.IPV6).
		SetNextHopMode(gosnappi.BgpV6RouteRangeNextHopMode.MANUAL)

	d2BgpIpv6Interface1Peer1V6Route1.Addresses().Add().
		SetAddress("0:20:20:20::0").
		SetPrefix(64).
		SetCount(2).
		SetStep(2)

	d2BgpIpv6Interface1Peer1V6Route1.Advanced().
		SetMultiExitDiscriminator(40).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	d2BgpIpv6Interface1Peer1V6Route1.Communities().Add().
		SetAsNumber(100).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	d2BgpIpv6Interface1Peer1V6Route1AsPath := d2BgpIpv6Interface1Peer1V6Route1.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	d2BgpIpv6Interface1Peer1V6Route1AsPath.Segments().Add().
		SetAsNumbers([]int64{2223, 2224, 2225}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	// add protocol stacks for device d3

	d3Eth1 := d3.Ethernets().
		Add().
		SetName("d3Eth").
		SetPortName(p3.Name()).
		SetMac("00:00:03:03:03:02").
		SetMtu(1500)

	d3Eth1.
		Ipv6Addresses().
		Add().
		SetName("p3d1ipv6").
		SetAddress("0:3:3:3::2").
		SetGateway("0:3:3:3::1").
		SetPrefix(64)

	d3Bgp := d3.Bgp().
		SetRouterId("3.3.3.2")

	d3BgpIpv6Interface1 := d3Bgp.
		Ipv6Interfaces().Add().
		SetIpv6Name("p3d1ipv6")

	d3BgpIpv6Interface1Peer1 := d3BgpIpv6Interface1.
		Peers().
		Add().
		SetAsNumber(3332).
		SetAsType(gosnappi.BgpV6PeerAsType.EBGP).
		SetPeerAddress("0:3:3:3::1").
		SetName("BGPv6 Peer 3")

	d3BgpIpv6Interface1Peer1V6Route1 := d3BgpIpv6Interface1Peer1.
		V6Routes().
		Add().
		SetNextHopIpv6Address("0:3:3:3::2").
		SetName("p3d1peer1rrv6").
		SetNextHopAddressType(gosnappi.BgpV6RouteRangeNextHopAddressType.IPV6).
		SetNextHopMode(gosnappi.BgpV6RouteRangeNextHopMode.MANUAL)

	d3BgpIpv6Interface1Peer1V6Route1.Addresses().Add().
		SetAddress("0:30:30:30::0").
		SetPrefix(64).
		SetCount(2).
		SetStep(2)

	d3BgpIpv6Interface1Peer1V6Route1.Advanced().
		SetMultiExitDiscriminator(33).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	d3BgpIpv6Interface1Peer1V6Route1.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	d3BgpIpv6Interface1Peer1V6Route1AsPath := d3BgpIpv6Interface1Peer1V6Route1.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	d3BgpIpv6Interface1Peer1V6Route1AsPath.Segments().Add().
		SetAsNumbers([]int64{3333, 3334}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	// add endpoints and packet description flow 1
	f1 := config.Flows().Items()[0]
	f1.SetName(p1.Name() + " -> " + p2.Name()).
		TxRx().Device().
		SetTxNames([]string{d1BgpIpv6Interface1Peer1V6Route1.Name()}).
		SetRxNames([]string{d2BgpIpv6Interface1Peer1V6Route1.Name()})

	f1Eth := f1.Packet().Add().Ethernet()
	f1Eth.Src().SetValue(d1Eth1.Mac())
	f1Eth.Dst().SetValue("00:00:00:00:00:00")

	f1Ip := f1.Packet().Add().Ipv6()
	f1Ip.Src().SetValue("0:10:10:10::1")
	f1Ip.Dst().SetValue("0:20:20:20::1")

	// add endpoints and packet description flow 2
	f2 := config.Flows().Items()[1]
	f2.SetName(p1.Name() + " -> " + p3.Name()).
		TxRx().Device().
		SetTxNames([]string{d1BgpIpv6Interface1Peer1V6Route1.Name()}).
		SetRxNames([]string{d3BgpIpv6Interface1Peer1V6Route1.Name()})

	f2Eth := f2.Packet().Add().Ethernet()
	f2Eth.Src().SetValue(d1Eth1.Mac())
	f2Eth.Dst().SetValue("00:00:00:00:00:00")

	f2Ip := f2.Packet().Add().Ipv6()
	f2Ip.Src().SetValue("0:10:10:10::1")
	f2Ip.Dst().SetValue("0:30:30:30::1")

	// add endpoints and packet description flow 3
	f3 := config.Flows().Items()[2]
	f3.SetName(p2.Name() + " -> " + p1.Name()).
		TxRx().Device().
		SetTxNames([]string{d2BgpIpv6Interface1Peer1V6Route1.Name()}).
		SetRxNames([]string{d1BgpIpv6Interface1Peer1V6Route1.Name()})

	f3Eth := f3.Packet().Add().Ethernet()
	f3Eth.Src().SetValue(d2Eth1.Mac())
	f3Eth.Dst().SetValue("00:00:00:00:00:00")

	f3Ip := f3.Packet().Add().Ipv6()
	f3Ip.Src().SetValue("0:20:20:20::1")
	f3Ip.Dst().SetValue("0:10:10:10::1")

	// add endpoints and packet description flow 4
	f4 := config.Flows().Items()[3]
	f4.SetName(p3.Name() + " -> " + p1.Name()).
		TxRx().Device().
		SetTxNames([]string{d3BgpIpv6Interface1Peer1V6Route1.Name()}).
		SetRxNames([]string{d1BgpIpv6Interface1Peer1V6Route1.Name()})

	f4Eth := f4.Packet().Add().Ethernet()
	f4Eth.Src().SetValue(d3Eth1.Mac())
	f4Eth.Dst().SetValue("00:00:00:00:00:00")

	f4Ip := f4.Packet().Add().Ipv6()
	f4Ip.Src().SetValue("0:30:30:30::1")
	f4Ip.Dst().SetValue("0:10:10:10::1")

	return config

}

func Bgpv4RoutesConfig(t *testing.T, otg *knebind.OTG) gosnappi.Config {
	config := otg.NewConfig(t)

	// add ports
	var names []string
	var locations []string
	for name, location := range otg.Ports() {
		names = append(names, name)
		locations = append(locations, location)
	}

	p1 := config.Ports().Add().SetName(names[0]).SetLocation(locations[0])
	p2 := config.Ports().Add().SetName(names[1]).SetLocation(locations[1])

	// add devices
	d1 := config.Devices().Add().SetName("p1d1")
	d2 := config.Devices().Add().SetName("p2d1")

	// add flows and common properties
	for i := 1; i <= 2; i++ {
		flow := config.Flows().Add()
		flow.Metrics().SetEnable(true)
		flow.Duration().FixedPackets().SetPackets(1000)
		flow.Rate().SetPps(500)
	}

	// add protocol stacks for device d1
	d1Eth1 := d1.Ethernets().
		Add().
		SetName("p1d1eth").
		SetPortName(p1.Name()).
		SetMac("00:00:01:01:01:01").
		SetMtu(1500)

	d1Eth1.
		Ipv4Addresses().
		Add().
		SetName("p1d1ipv4").
		SetAddress("1.1.1.2").
		SetGateway("1.1.1.1").
		SetPrefix(24)

	d1Bgp := d1.Bgp().
		SetRouterId("1.1.1.2")

	d1BgpIpv4Interface1 := d1Bgp.
		Ipv4Interfaces().Add().
		SetIpv4Name("p1d1ipv4")

	d1BgpIpv4Interface1Peer1 := d1BgpIpv4Interface1.
		Peers().
		Add().
		SetAsNumber(2222).
		SetAsType(gosnappi.BgpV4PeerAsType.EBGP).
		SetPeerAddress("1.1.1.1").
		SetName("p1d1bgpv4")

	d1BgpIpv4Interface1Peer1V4Route1 := d1BgpIpv4Interface1Peer1.
		V4Routes().
		Add().
		SetNextHopIpv4Address("1.1.1.2").
		SetName("p1d1peer1rrv4").
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)

	d1BgpIpv4Interface1Peer1V4Route1.Addresses().Add().
		SetAddress("10.10.10.1").
		SetPrefix(32).
		SetCount(4).
		SetStep(1)

	d1BgpIpv4Interface1Peer1V4Route1.Advanced().
		SetMultiExitDiscriminator(50).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	d1BgpIpv4Interface1Peer1V4Route1.Communities().Add().
		SetAsNumber(1).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	d1BgpIpv4Interface1Peer1V4Route1AsPath := d1BgpIpv4Interface1Peer1V4Route1.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	d1BgpIpv4Interface1Peer1V4Route1AsPath.Segments().Add().
		SetAsNumbers([]int64{1112, 1113}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	// add protocol stacks for device d2
	d2Eth1 := d2.Ethernets().
		Add().
		SetName("p2d1eth").
		SetPortName(p2.Name()).
		SetMac("00:00:02:02:02:02").
		SetMtu(1500)

	d2Eth1.
		Ipv4Addresses().
		Add().
		SetName("p2d1ipv4").
		SetAddress("2.2.2.2").
		SetGateway("2.2.2.1").
		SetPrefix(32)

	d2Bgp := d2.Bgp().
		SetRouterId("2.2.2.2")

	d2BgpIpv4Interface1 := d2Bgp.
		Ipv4Interfaces().Add().
		SetIpv4Name("p2d1ipv4")

	d2BgpIpv4Interface1Peer1 := d2BgpIpv4Interface1.
		Peers().
		Add().
		SetAsNumber(3333).
		SetAsType(gosnappi.BgpV4PeerAsType.EBGP).
		SetPeerAddress("2.2.2.1").
		SetName("p2d1bgpv4")

	d2BgpIpv4Interface1Peer1V4Route1 := d2BgpIpv4Interface1Peer1.
		V4Routes().
		Add().
		SetNextHopIpv4Address("2.2.2.2").
		SetName("p2d1peer1rrv4").
		SetNextHopAddressType(gosnappi.BgpV4RouteRangeNextHopAddressType.IPV4).
		SetNextHopMode(gosnappi.BgpV4RouteRangeNextHopMode.MANUAL)

	d2BgpIpv4Interface1Peer1V4Route1.Addresses().Add().
		SetAddress("20.20.20.1").
		SetPrefix(32).
		SetCount(2).
		SetStep(2)

	d2BgpIpv4Interface1Peer1V4Route1.Advanced().
		SetMultiExitDiscriminator(40).
		SetOrigin(gosnappi.BgpRouteAdvancedOrigin.EGP)

	d2BgpIpv4Interface1Peer1V4Route1.Communities().Add().
		SetAsNumber(100).
		SetAsCustom(2).
		SetType(gosnappi.BgpCommunityType.MANUAL_AS_NUMBER)

	d2BgpIpv4Interface1Peer1V4Route1AsPath := d2BgpIpv4Interface1Peer1V4Route1.AsPath().
		SetAsSetMode(gosnappi.BgpAsPathAsSetMode.INCLUDE_AS_SET)

	d2BgpIpv4Interface1Peer1V4Route1AsPath.Segments().Add().
		SetAsNumbers([]int64{2223, 2224, 2225}).
		SetType(gosnappi.BgpAsPathSegmentType.AS_SEQ)

	// add endpoints and packet description flow 1
	f1 := config.Flows().Items()[0]
	f1.SetName(p1.Name() + " -> " + p2.Name()).
		TxRx().Device().
		SetTxNames([]string{d1BgpIpv4Interface1Peer1V4Route1.Name()}).
		SetRxNames([]string{d2BgpIpv4Interface1Peer1V4Route1.Name()})

	f1Eth := f1.Packet().Add().Ethernet()
	f1Eth.Src().SetValue(d1Eth1.Mac())
	f1Eth.Dst().SetValue("00:00:00:00:00:00")

	f1Ip := f1.Packet().Add().Ipv4()
	f1Ip.Src().SetValue("10.10.10.1")
	f1Ip.Dst().SetValue("20.20.20.1")

	// add endpoints and packet description flow 2
	f2 := config.Flows().Items()[1]
	f2.SetName(p2.Name() + " -> " + p1.Name()).
		TxRx().Device().
		SetTxNames([]string{d2BgpIpv4Interface1Peer1V4Route1.Name()}).
		SetRxNames([]string{d1BgpIpv4Interface1Peer1V4Route1.Name()})

	f2Eth := f2.Packet().Add().Ethernet()
	f2Eth.Src().SetValue(d2Eth1.Mac())
	f2Eth.Dst().SetValue("00:00:00:00:00:00")

	f2Ip := f2.Packet().Add().Ipv4()
	f2Ip.Src().SetValue("20.20.20.1")
	f2Ip.Dst().SetValue("10.10.10.1")

	return config
}

func PortAndFlowMetricsOk(api gosnappi.GosnappiApi, config gosnappi.Config) (bool, error) {
	expected := 0
	for _, f := range config.Flows().Items() {
		expected += int(f.Duration().FixedPackets().Packets())
	}

	fNames := []string{}
	for _, f := range config.Flows().Items() {
		fNames = append(fNames, f.Name())
	}

	pNames := []string{}
	for _, p := range config.Ports().Items() {
		pNames = append(pNames, p.Name())
	}

	fReq := api.NewMetricsRequest()
	fReq.Flow().SetFlowNames(fNames)
	fMetrics, err := api.GetMetrics(fReq)
	if err != nil {
		return false, err
	}

	pReq := api.NewMetricsRequest()
	pReq.Port().SetPortNames(pNames)
	pMetrics, err := api.GetMetrics(pReq)
	if err != nil {
		return false, err
	}

	actual := 0
	for _, m := range fMetrics.FlowMetrics().Items() {
		log.Printf("Flow metric: Name: %v, Frames Tx: %v, Frames Rx: %v ...\n", m.Name(), m.FramesTx(), m.FramesRx())
		actual += int(m.FramesRx())
	}

	for _, p := range pMetrics.PortMetrics().Items() {
		log.Printf("Port metric: Name: %v, Frames Tx: %v, Frames Rx: %v ...\n", p.Name(), p.FramesTx(), p.FramesRx())
	}
	log.Printf("################################################\n\n")

	return expected == actual, nil
}

func AllBgp4SessionUp(api gosnappi.GosnappiApi, config gosnappi.Config) (bool, error) {
	dNames := []string{}
	for _, d := range config.Devices().Items() {
		bgp := d.Bgp()
		for _, ipv4 := range bgp.Ipv4Interfaces().Items() {
			for _, peer := range ipv4.Peers().Items() {
				dNames = append(dNames, peer.Name())
			}
		}
	}

	req := api.NewMetricsRequest()
	req.Bgpv4().SetPeerNames(dNames)

	dMetrics, err := api.GetMetrics(req)
	if err != nil {
		return false, err
	}

	routesTx := []int{}
	routesRx := []int{}
	actualSessionUp := 0
	for _, d := range dMetrics.Bgpv4Metrics().Items() {
		log.Printf("BGPv4 metric: Name: %v, Session State: %v, Routes Tx: %v, Routes Rx: %v ...\n", d.Name(), d.SessionState(), d.RoutesAdvertised(), d.RoutesReceived())
		if d.SessionState() == gosnappi.Bgpv4MetricSessionState.UP {
			actualSessionUp += 1
		}
		routesTx = append(routesTx, int(d.RoutesAdvertised()))
		routesRx = append(routesRx, int(d.RoutesReceived()))
	}
	log.Printf("################################################\n\n")

	return len(dNames) == actualSessionUp && TxRxRoutesOk(routesTx, routesRx), nil
}

func AllBgp6SessionUp(api gosnappi.GosnappiApi, config gosnappi.Config) (bool, error) {
	dNames := []string{}
	for _, d := range config.Devices().Items() {
		bgp := d.Bgp()
		for _, ipv6 := range bgp.Ipv6Interfaces().Items() {
			for _, peer := range ipv6.Peers().Items() {
				dNames = append(dNames, peer.Name())
			}
		}
	}

	req := api.NewMetricsRequest()
	req.Bgpv6().SetPeerNames(dNames)

	dMetrics, err := api.GetMetrics(req)
	if err != nil {
		return false, err
	}

	routesTx := []int{}
	routesRx := []int{}
	actualSessionUp := 0
	for _, d := range dMetrics.Bgpv6Metrics().Items() {
		log.Printf("BGPv6 metric: Name: %v, Session State: %v, Routes Tx: %v, Routes Rx: %v ...\n", d.Name(), d.SessionState(), d.RoutesAdvertised(), d.RoutesReceived())
		if d.SessionState() == gosnappi.Bgpv6MetricSessionState.UP {
			actualSessionUp += 1
		}
		routesTx = append(routesTx, int(d.RoutesAdvertised()))
		routesRx = append(routesRx, int(d.RoutesReceived()))
	}
	log.Printf("################################################\n\n")

	return len(dNames) == actualSessionUp && TxRxRoutesOk(routesTx, routesRx), nil
}

func WaitFor(t *testing.T, fn func() (bool, error), opts *WaitForOpts) error {
	if opts == nil {
		opts = &WaitForOpts{
			Condition: "condition to be true",
		}
	}
	defer Timer(time.Now(), fmt.Sprintf("Waiting for %s", opts.Condition))

	if opts.Interval == 0 {
		opts.Interval = 500 * time.Millisecond
	}
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	start := time.Now()
	log.Printf("Waiting for %s ...\n", opts.Condition)

	for {
		done, err := fn()
		if err != nil {
			t.Fatal(fmt.Errorf("error waiting for %s: %v", opts.Condition, err))
		}
		if done {
			log.Printf("Done waiting for %s\n", opts.Condition)
			return nil
		}

		if time.Since(start) > opts.Timeout {
			t.Fatal(fmt.Errorf("timeout occurred while waiting for %s", opts.Condition))
		}
		time.Sleep(opts.Interval)
	}
}

func Timer(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %d ms", name, elapsed.Milliseconds())
}

func TxRxRoutesOk(tx, rx []int) bool {
	if len(tx) != len(rx) {
		return false
	}

	totalTx := 0
	for _, t := range tx {
		// not ok if not routes sent for any of the peer
		if t == 0 {
			return false
		}
		totalTx += t
	}

	for i := range rx {
		// not ok if expected rx doesn't match sum of all tx minus self tx
		if rx[i] != totalTx-tx[i] {
			return false
		}
	}

	return true
}
