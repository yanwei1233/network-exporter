package probe

import (
	"context"
	"fmt"
	"github.com/go-ping/ping"
	"github.com/prometheus/client_golang/prometheus"
	"k8s-network-probe/pkg"
	"k8s.io/klog/v2"
	"time"
)

// https://github.com/go-ping/ping/blob/master/cmd/ping/ping.go
func (pm *ProbeManager) PingProbe(t *ProbeTarget, metricCh chan<- prometheus.Metric) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(t.ProbeTw))
	defer cancel()

	klog.Infof("PingProbe.start[target:%+v]", t)
	target := t.DestAddr
	var (
		result    = pkg.ProbeFailed
		seconds   float64
		errReason = pkg.ErrReasonEmpty
	)

	pinger, err := ping.NewPinger(t.DestAddr)
	if err != nil {
		return
	}
	exitCh := make(chan struct{}, 1)
	go func() {
		for {
			select {
			// 当这个ping targets的时间到的时候 ，在pinger 关闭的时候就可以调用
			case <-ctx.Done():
				pinger.Stop()
				result = -1
				seconds = -1
				errReason = fmt.Sprintf("ping.timeout:%v", t.ProbeTw)

				//mc := prometheus.MustNewConstMetric(M_PROBE_RESULT,
				//	prometheus.GaugeValue,
				//	result,
				//	pm.nodeName,
				//	pm.localIp,
				//	pm.runType,
				//	target,
				//	pkg.PROBE_FUNC_PING,
				//	errReason,
				//)
				//
				//mc2 := prometheus.MustNewConstMetric(M_PROBE_DURATION,
				//	prometheus.GaugeValue,
				//	seconds,
				//	pm.nodeName,
				//	pm.localIp,
				//	pm.runType,
				//	pkg.PROBE_FUNC_PING,
				//	target,
				//)
				//
				//metricCh <- mc
				//metricCh <- mc2
				//return
			case <-exitCh:
				return
			}
		}

	}()

	// 只发1个ICMP hello报文
	packetsRecv := 0
	pinger.Count = 1
	//pinger.Timeout = time.Duration(t.ProbeTw) * time.Second
	//pinger.Interval = time.Microsecond * 100
	pinger.SetPrivileged(true)
	// 如何定义ping 成功与否，影响的报文数量是否为1
	// OnFinish回调函数 在pinger 关闭的时候就可以调用
	pinger.OnFinish = func(stats *ping.Statistics) {
		seconds = stats.AvgRtt.Seconds()
		packetsRecv = stats.PacketsRecv
		fmt.Printf("\n--- %s ping statistics ---\n", stats.Addr)
		fmt.Printf("%d packets transmitted, %d packets received, %d duplicates, %v%% packet loss\n",
			stats.PacketsSent, stats.PacketsRecv, stats.PacketsRecvDuplicates, stats.PacketLoss)
		fmt.Printf("round-trip min/avg/max/stddev = %v/%v/%v/%v\n",
			stats.MinRtt, stats.AvgRtt, stats.MaxRtt, stats.StdDevRtt)
	}

	err = pinger.Run()
	if err != nil {
		errReason = fmt.Sprintf("ping.Run.err:%v", err)
		return
	}

	// 有回包才认为成功
	if packetsRecv == 1 {
		result = pkg.ProbeSuccess
	} else {
		errReason = fmt.Sprintf("ping.packetsRecv.zero:%v", packetsRecv)
	}

	if result == -1 {
		seconds = -1
	}

	mc := prometheus.MustNewConstMetric(M_PROBE_RESULT,
		prometheus.GaugeValue,
		result,
		pm.nodeName,
		pm.localIp,
		pm.runType,
		target,
		pkg.PROBE_FUNC_PING,
		errReason,
	)

	mc2 := prometheus.MustNewConstMetric(M_PROBE_DURATION,
		prometheus.GaugeValue,
		seconds,
		pm.nodeName,
		pm.localIp,
		pm.runType,
		pkg.PROBE_FUNC_PING,
		target,
	)
	metricCh <- mc
	metricCh <- mc2
	exitCh <- struct{}{}
	klog.Infof("PingProbe.end[target:%+v]", t)
}
