package addrha

import (
	"sync"
	"net/url"
	"errors"
	"time"
	"log"
	"fmt"
)

const (
	DEFAULT_SPEEDTEST_INTERVAL = 10 * time.Second
	//默认系统最大的测速时间
	DEFAULT_SPEED_MAX = time.Hour

	//http协议支持完整url请求，返回2XX系统状态码即可
	SCHEME_HTTP = "http"
	SCHEME_TCP  = "tcp"
	SCHEME_UDP  = "udp"

	//icmp是原始ping命令结果，可以统计丢包率，要久点 https://github.com/sparrc/go-ping
	SCHEME_ICMP = "icmp"
)

type Addr struct {
	uinfo *url.URL
}

//测试数据结果
type Speed struct {
	//测试返回的计时
	Dur time.Duration
	//测试时发生的错误
	Err error
}

type AddrHa struct {
	addrList map[string]*Addr

	//key is Addr.uinfo.String()
	speed map[string]*Speed

	testInterval time.Duration

	sync.Mutex

	//是否初始化
	inited bool

	//是否退出
	isExit bool

	isDebug bool
}

func NewAddr(uinfo *url.URL) *Addr {
	return &Addr{uinfo: uinfo}
}

func NewAddrHa(list ...string) (*AddrHa, error) {
	addrList := make(map[string]*Addr)

	for _, val := range list {
		uinfo, err := url.Parse(val)
		if err != nil {
			return nil, errors.New("Error in " + val + ": " + err.Error())
		}

		if _, ok := addrList[uinfo.String()]; ok {
			return nil, errors.New(fmt.Sprintf("Addr %s exists already, please check.", uinfo.String()))
		}

		addr := NewAddr(uinfo)
		addrList[addr.Key()] = addr
	}

	if len(addrList) == 0 {
		return nil, errors.New("实际可用addr列表为空，请确认")
	}

	ha := &AddrHa{
		addrList:     addrList,
		speed:        make(map[string]*Speed),
		testInterval: DEFAULT_SPEEDTEST_INTERVAL,
		isExit:       false,
		inited:       false,
		isDebug:      false,
	}

	//back ground update
	go ha.bgRuntime()

	return ha, nil
}

//获取测试结果接口
func (o *AddrHa) SpeedResult() (map[string]*Speed) {
	o.isInited()

	return o.speed
}

func (o *AddrHa) isInited() () {
	var isInited bool
	o.Lock()
	isInited = o.inited
	o.Unlock()

	if !isInited {
		if o.isDebug {
			log.Println("SpeedResult initing 触发更新测速...")
		}
		o.testNowSync()
	}
}

//获取测试结果接口
func (o *AddrHa) FatestAddr() (*url.URL, time.Duration) {
	fatest := DEFAULT_SPEED_MAX
	var faddr *url.URL

	o.isInited()

	o.Lock()
	o.Unlock()
	for addr, speed := range o.speed {
		if speed.DurationForCompare() < fatest {
			fatest = speed.DurationForCompare()
			faddr = o.addrList[addr].UrlInfo()
		}
	}

	return faddr, fatest
}

//获取测试结果
func (o *AddrHa) testNowAsync() () {
	if o.isDebug {
		log.Println("AddrHa testNowAsync...")
	}

	//wait group
	wg := &sync.WaitGroup{}
	for _, addr := range o.addrList {
		wg.Add(1)
		go func(addr *Addr) {
			o.testForSingal(addr)
			wg.Done()
		}(addr)
	}
	wg.Wait()

	o.Lock()
	o.inited = true
	o.Unlock()
}

func (o *AddrHa) testNowSync() () {
	if o.isDebug {
		log.Println("AddrHa testNowSync...")
	}
	for _, addr := range o.addrList {
		o.testForSingal(addr)
	}

	o.Lock()
	o.inited = true
	o.Unlock()
}

func (o *AddrHa) testForSingal(addr *Addr) () {
	dur, err := addr.Ping()
	speed := &Speed{Dur: dur, Err: err}
	if o.isDebug {
		if err != nil {
			log.Println(fmt.Sprintf("SpeedResult error of %s: %v", addr.Key(), err.Error()))
		} else {
			log.Println(fmt.Sprintf("SpeedResult of %s: %v", addr.Key(), dur.String()))
		}
	}

	o.Lock()
	o.speed[addr.Key()] = speed
	o.Unlock()
}

func (o *AddrHa) EnableDebug() () {
	o.Lock()
	o.Unlock()

	o.isDebug = true
}

//间隔获取测速结果
func (o *AddrHa) bgRuntime() () {
	tick := time.NewTicker(o.testInterval)

	for {
		//exit signal
		if o.isExit {
			tick.Stop()

			break
		}

		select {
		case <-tick.C:
			if o.isDebug {
				log.Println("tick.C 触发更新测速...")
			}

			//触发测试
			o.testNowAsync()
		}
	}
}

//停止测速
func (o *AddrHa) Stop() () {
	o.Lock()
	defer o.Unlock()

	o.isExit = true
}

//增加一个地址 只能单个
func (o *AddrHa) Add(addr string) (error) {
	uinfo, err := url.Parse(addr)
	if err != nil {
		return errors.New("Error in " + addr + ": " + err.Error())
	}
	addrObj := NewAddr(uinfo)

	o.Lock()
	defer o.Unlock()

	if _, ok := o.addrList[addrObj.Key()]; ok {
		return errors.New(fmt.Sprintf("Addr %s exists already, please check.", uinfo.String()))
	}

	o.addrList[addrObj.Key()] = addrObj

	//update once
	go o.testNowAsync()
	return nil
}

//删除一个地址 只能单个
func (o *AddrHa) Remove(addr string) (error) {
	uinfo, err := url.Parse(addr)
	if err != nil {
		return errors.New("Error in " + addr + ": " + err.Error())
	}
	addrObj := NewAddr(uinfo)

	o.Lock()
	defer o.Unlock()

	if _, ok := o.addrList[addrObj.Key()]; !ok {
		return errors.New(fmt.Sprintf("Addr %s not exists, please check.", uinfo.String()))
	}

	delete(o.addrList, addrObj.Key())

	//update once
	go o.testNowAsync()
	return nil
}

func (o *Addr) Key() (string) {
	return o.uinfo.String()
}

func (o *Addr) UrlInfo() (*url.URL) {
	return o.uinfo
}

// time ping for single addr
func (o *Addr) Ping() (time.Duration, error) {
	var dur time.Duration
	var err error

	switch o.uinfo.Scheme {
	case SCHEME_HTTP:
		dur, err = SpeedTestHttp(o.uinfo.String())
	case SCHEME_TCP:
		dur, err = SpeedTestTcp(o.uinfo.Host)
	case SCHEME_UDP:
		dur, err = SpeedTestUdp(o.uinfo.Host)
	case SCHEME_ICMP:
		dur, err = SpeedTestIcmpPing(o.uinfo.Hostname())
	default:
		return dur, errors.New("Not support scheme: " + o.uinfo.Scheme)
	}

	return dur, err
}

func (o *Speed) String() (string) {
	if o.Err != nil {
		return o.Err.Error()
	} else {
		return o.Dur.String()
	}
}

//获取用来排序的时间，如果出错，直接返回默认最大值
func (o *Speed) DurationForCompare() (time.Duration) {
	if o.Err != nil {
		return DEFAULT_SPEED_MAX
	} else {
		return o.Dur
	}
}