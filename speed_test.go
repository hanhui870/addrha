package addrha

import (
	"testing"
	"net/url"
	"github.com/zhujingfa/go-ping"
	"time"
)

func TestFindFatestServerTest(t *testing.T) {
	hosts := []string{"tcp://hn99.node.ifanghui.cn:8500", "tcp://hn88.node.ifanghui.cn:8500", "tcp://hn77.node.ifanghui.cn:8500"}
	if len(hosts) != 3 {
		t.Fatalf("Found err len of hosts: %v", len(hosts))
	}

	addr := "tcp://hn99.node.ifanghui.cn:8500/user?query=hello"
	uinfo, err := url.Parse(addr)
	if err != nil {
		t.Fatalf("Found err when Parse: %v", err.Error())
	}

	t.Logf("Parse result: scheme %v, host %v, path %v, query %v RequestURI %v", uinfo.Scheme, uinfo.Host, uinfo.Path, uinfo.RawQuery, uinfo.RequestURI())
	if uinfo.Scheme != "tcp" {
		t.Fatalf("Found err uinfo.Scheme!=tcp: %v", uinfo.Scheme)
	}
	if uinfo.Host != "hn99.node.ifanghui.cn:8500" {
		t.Fatalf("Found err uinfo.Host!=hn99.node.ifanghui.cn:8500: %v", uinfo.Host)
	}
	if uinfo.Hostname() != "hn99.node.ifanghui.cn" {
		t.Fatalf("Found err uinfo.Hostname()!=hn99.node.ifanghui.cn: %v", uinfo.Host)
	}

	ha, err := NewAddrHa(hosts...)
	if err != nil {
		t.Fatalf("Found err when NewAddrHa(hosts...): %v", err.Error())
	}
	ha.EnableDebug()
	result := ha.SpeedResult()
	for i, v := range result {
		t.Logf("ha.SpeedResult: %v, %v", i, v.String())
	}

	uinfo, dur := ha.FatestAddr()
	if dur == DEFAULT_SPEED_MAX {
		t.Fatalf("ha.FatestAddr error: dur==DEFAULT_SPEED_MAX.")
	}
	t.Logf("ha.FatestAddr: %v, %v", uinfo.String(), dur)
}


func TestSpeedHttpCall(t *testing.T) {
	url := "http://hn88.node.ifanghui.cn:8500/v1/internal/ui/nodes?dc=dc1&token="
	dur, err := SpeedTestHttp(url)
	if err != nil {
		t.Fatalf("Found err when SpeedTestTcp(%v): %v", url, err.Error())
	}
	t.Logf("Duration for connect %v, %v", url, dur.String())

}


func TestSpeedTcpUdpCall(t *testing.T) {
	addr := "hn99.node.ifanghui.cn:8500"
	dur, err := SpeedTestTcp(addr)
	if err != nil {
		t.Fatalf("Found err when SpeedTestTcp(%v): %v", addr, err.Error())
	}
	t.Logf("Duration for connect %v, %v", addr, dur.String())

	addr = "hn99.node.ifanghui.cn:53"
	dur, err = SpeedTestUdp(addr)
	if err != nil {
		t.Fatalf("Found err when SpeedTestUdp(%v): %v", addr, err.Error())
	}
	t.Logf("Duration for SpeedTestUdp %v, %v", addr, dur.String())

	addr = "114.114.114.114:53"
	dur, err = SpeedTestUdp(addr)
	if err != nil {
		t.Fatalf("Found err when SpeedTestUdp(%v): %v", addr, err.Error())
	}
	t.Logf("Duration for SpeedTestUdp %v, %v", addr, dur.String())
}

func TestSpeedPingCall(t *testing.T) {
	addr := "hn99.node.ifanghui.cn"
	dur, err := SpeedTestIcmpPing(addr)
	if err != nil {
		t.Fatalf("Found err when SpeedTestIcmpPing(%v): %v", addr, err.Error())
	}
	t.Logf("AVG Duration for ping %v, %v", addr, dur.String())

	addr = "ifanghui.com"
	dur, err = SpeedTestIcmpPing(addr)
	if err != nil {
		t.Fatalf("Found err when SpeedTestIcmpPing(%v): %v", addr, err.Error())
	}
	t.Logf("AVG Duration for ping %v, %v", addr, dur.String())

	addr = "baidu.com"
	dur, err = SpeedTestIcmpPing(addr)
	if err != nil {
		t.Fatalf("Found err when SpeedTestIcmpPing(%v): %v", addr, err.Error())
	}
	t.Logf("AVG Duration for ping %v, %v", addr, dur.String())

	addr = "google.com"
	dur, err = SpeedTestIcmpPing(addr)
	if err == nil {
		t.Fatalf("Found err when SpeedTestIcmpPing(%v): 在无VPN环境下ping肯定完全丢包", addr)
	}
}

func TestPingRawCall(t *testing.T) {
	addr := "localhost"
	pinger, err := ping.NewPinger(addr)
	pinger.Count = 10
	pinger.Interval = 10 * time.Millisecond
	pinger.Timeout = time.Second * 2
	if err != nil {
		t.Fatalf("Found err when ping.NewPinger(addr): %v", err.Error())
	}

	pinger.OnRecv = func(pkt *ping.Packet) {
		t.Logf("%d bytes from %s: icmp_seq=%d time=%v\n",
			pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	}
	pinger.OnFinish = func(stats *ping.Statistics) {
		t.Logf("\n--- %s ping statistics ---\n", stats.Addr)
		t.Logf("%d packets transmitted, %d packets received, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketLoss)
		t.Logf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}

	t.Logf("PING %s (%s):\n", pinger.Addr(), pinger.IPAddr())
	pinger.Run()
}
