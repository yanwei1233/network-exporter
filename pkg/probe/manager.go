package probe

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gammazero/workerpool"
	"github.com/prometheus/client_golang/prometheus"
	pconfig "github.com/prometheus/common/config"
	"k8s-network-probe/pkg/utils"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"time"

	"k8s-network-probe/pkg"
	"sync"
)

type ProbeManager struct {
	serverAddr                  string // server地址
	refreshIntervalSeconds      int    // 多久去刷新一次 targets
	requestServerTimeOutSeconds int    // 请求刷新的超时时间
	nodeName                    string
	localIp                     string
	runType                     string
	probeGlobalTimeoutSeconds   int
	localTs                     []*ProbeTarget // 探测目标的本地的缓存 ，一种是去刷新，一种是 获取 ，不同goroutine之间
	sync.RWMutex
}

func NewProbeManager(serverAddr, nodeName, localIp, runType string, probeGlobalTimeoutSeconds, refreshIntervalSeconds int) *ProbeManager {
	pm := &ProbeManager{
		serverAddr:                serverAddr,
		nodeName:                  nodeName,
		localIp:                   localIp,
		runType:                   runType,
		probeGlobalTimeoutSeconds: probeGlobalTimeoutSeconds,
		refreshIntervalSeconds:    refreshIntervalSeconds,
	}
	return pm
}

func (pb *ProbeManager) Describe(ch chan<- *prometheus.Desc) {
	ch <- M_PROBE_RESULT
	ch <- M_PROBE_DURATION
}

func (pm *ProbeManager) Collect(metricCh chan<- prometheus.Metric) {
	// 拿到本地的targets

	ts := pm.GetTs()
	// 先new wp 对象，for 遍历的你的 数据，wp.submit提交任务
	// 这里千万不要 之间go 出去执行，因为collect 方法执行完，ch会关闭：向关闭的ch中写数据会panic
	wp := workerpool.New(100)
	for _, t := range ts {
		switch t.Func {
		case pkg.PROBE_FUNC_DNS:
			wp.Submit(func() {
				pm.DnsProbe(t, metricCh)
			})
		case pkg.PROBE_FUNC_TCP:
			wp.Submit(func() {
				pm.TcpProbe(t, metricCh)
			})
		case pkg.PROBE_FUNC_HTTP:
			wp.Submit(func() {
				pm.HttpProbe(t, metricCh)
			})
		case pkg.PROBE_FUNC_PING:
			wp.Submit(func() {
				pm.PingProbe(t, metricCh)
			})
		case pkg.PROBE_FUNC_UDP:
			wp.Submit(func() {
				pm.UdpProbe(t, metricCh)
			})
		default:
			klog.Errorf("ProbeManager.Collect.func.type.err:%v", t.Func)
			continue
		}
	}
	wp.StopWait()

}

type ProbeTarget struct {
	DestAddr string `json:"dest_addr" validate:"required"` // 目标地址 可以是域名， 也可以是 ip地址  or tcp 10.1.1.1:2379 http://xxxxx/
	Func     string `json:"func" validate:"required"`      // ping/tcp/http/dns
	ProbeTw  int    `json:"probe_tw" validate:"required"`  // 超时时间：过了这个秒数 就认为探测失败：不可能一直等待
}

func (pm *ProbeManager) SetTs(ts []*ProbeTarget) {
	pm.Lock()
	defer pm.Unlock()
	for i := 0; i < len(ts); i++ {
		t := ts[i]
		if t.ProbeTw == 0 {
			t.ProbeTw = pm.probeGlobalTimeoutSeconds
		}
	}
	pm.localTs = ts
}

func (pm *ProbeManager) GetTs() []*ProbeTarget {
	pm.RLock()
	defer pm.RUnlock()
	return pm.localTs
}

func (pm *ProbeManager) RefreshTargetManager(ctx context.Context) error {
	// 每隔 多长时间去执行一下 RunRefreshTargets ，直到 ctx.Done
	go wait.UntilWithContext(ctx, pm.RunRefreshTargets, time.Duration(pm.refreshIntervalSeconds)*time.Second)
	<-ctx.Done()
	klog.Infof("RunVolumeDiffManager.exit.receive_quit_signal")
	return nil

}

func (pb *ProbeManager) RunRefreshTargets(ctx context.Context) {

	hc := &pconfig.HTTPClientConfig{}

	params := map[string]string{
		"local_ip":  pb.localIp,
		"node_name": pb.nodeName,
		"run_type":  pb.runType,
	}
	url := fmt.Sprintf("%s/api/v1/probe-targets", pb.serverAddr)
	respByte, err := utils.GetWithBearerToken("RunFetchTargets", *hc, pb.requestServerTimeOutSeconds, url, params)
	if err != nil {
		klog.Errorf("[RunFetchTargets.request.error][url:%+v][params:%+v][err:%v]", pb.serverAddr, params, err)
		return
	}
	var ts []*ProbeTarget
	err = json.Unmarshal(respByte, &ts)
	if err != nil {
		klog.Errorf("[RunFetchTargets.json.Unmarshal.error][url:%+v][params:%+v][err:%v]", pb.serverAddr, params, err)
		return
	}
	if len(ts) > 0 {
		pb.SetTs(ts)
	}

}
