# Find a fastest addr lib for golang

这个module主要解决，从[10.0.0.1:8500, 10.0.0.2:8500, 10.0.0.3:8500] ip-port组合中自动选择最快并且可连接的地址。

可用于Failover HA、Registry HA等项目中。

```go
func TestFindFatestServerTest(t *testing.T) {
	hosts := []string{"tcp://10.0.0.1:8500", "tcp://10.0.0.2:8500", "tcp://10.0.0.3:8500"}
	if len(hosts) != 3 {
		t.Fatalf("Found err len of hosts: %v", len(hosts))
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
```