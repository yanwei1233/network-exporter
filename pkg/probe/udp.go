package probe

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s-network-probe/pkg"
	"k8s.io/klog/v2"
	"net"
	"time"
)

func (pm *ProbeManager) UdpProbe(t *ProbeTarget, metricCh chan<- prometheus.Metric) {
	klog.Infof("UdpProbe.start[target:%+v]", t)
	start := time.Now()
	target := t.DestAddr
	var (
		result    float64              // 探测成功与否 =-1 代表失败 =1 代表成功
		seconds   float64              // 探测耗时多少秒 0.1 代表100毫秒
		errReason = pkg.ErrReasonEmpty // 错误原因：connection refused
	)
	conn, err := net.DialTimeout("udp", target, time.Duration(t.ProbeTw)*time.Second)
	result = pkg.ProbeSuccess
	if err != nil {
		klog.Errorf("[net.DialTimeout.err][target:%v][err:%v]", target, err)
		result = pkg.ProbeFailed
		errReason = err.Error()

	} else {
		conn.Close()
		seconds = time.Since(start).Seconds()
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
		pkg.PROBE_FUNC_UDP,
		errReason,
	)
	mc2 := prometheus.MustNewConstMetric(M_PROBE_DURATION,
		prometheus.GaugeValue,
		seconds,
		pm.nodeName,
		pm.localIp,
		pm.runType,
		pkg.PROBE_FUNC_UDP,
		target,
	)
	metricCh <- mc
	metricCh <- mc2
	klog.Infof("UdpProbe.end[target:%+v]", t)
}
