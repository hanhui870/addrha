package addrha

import (
	"time"
	"net"

	"github.com/zhujingfa/go-ping"
	"errors"
	"fmt"
	"bufio"
	"net/url"
	"net/http"
)

const (
	DIAL_TIMEOUT = 2 * time.Second
)

// addr格式: host:port
func SpeedTestTcp(addr string) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout(SCHEME_TCP, addr, DIAL_TIMEOUT)
	if err != nil {
		return 0, err
	}

	defer conn.Close()

	dur := time.Now().Sub(start)
	return dur, nil
}

func SpeedTestUdp(addr string) (time.Duration, error) {
	start := time.Now()
	conn, err := net.DialTimeout(SCHEME_UDP, addr, DIAL_TIMEOUT)
	if err != nil {
		return 0, err
	}

	defer conn.Close()

	dur := time.Now().Sub(start)
	return dur, nil
}

//HTTP测速，会请求链接，检测http code
func SpeedTestHttp(fullurl string) (time.Duration, error) {
	uinfo, err := url.Parse(fullurl)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	conn, err := net.DialTimeout(SCHEME_TCP, uinfo.Host, DIAL_TIMEOUT)
	if err != nil {
		return 0, err
	}

	defer conn.Close()
	fmt.Fprintf(conn, "GET %s HTTP/1.1\r\nHost: %s\r\n\r\n", uinfo.RequestURI(), uinfo.Host)
	rsp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		return 0, err
	}

	if rsp.StatusCode == 200 {
		dur := time.Now().Sub(start)
		return dur, nil
	} else {
		return 0, errors.New(fmt.Sprintf("Found error http_code: %d", rsp.StatusCode))
	}
}

//使用传统icmp ping方法。
//host：传IP或者域名即可，无需port
func SpeedTestIcmpPing(host string) (time.Duration, error) {
	pinger, err := ping.NewPinger(host)
	//发包次数
	pinger.Count = 10
	//发包间隔：10毫秒
	pinger.Interval = 10 * time.Millisecond
	pinger.Timeout = time.Second * 1
	if err != nil {
		return 0, nil
	}

	//需要实体，不能指针，不然可能访问野指针
	var result ping.Statistics
	pinger.OnFinish = func(stats *ping.Statistics) {
		result = *stats
	}

	//run ping
	pinger.Run()

	//丢包率达到了10%直接返回错误
	if result.PacketLoss > 10 {
		//github.com/zhujingfa/go-ping库已经修复了，最多发count个包。
		//目前ping包处理丢包模型不正确，比如外网的，来回实际小于发包频率的会出错，比如google.com，10ms发包，返回时间200ms内可以发送大量包，然后计算就会出现大量丢包，其实不一定是丢包。
		return 0, errors.New(fmt.Sprintf("失败，丢包率太高，达到了：%f", result.PacketLoss))
	} else {
		return result.AvgRtt, nil
	}
}
