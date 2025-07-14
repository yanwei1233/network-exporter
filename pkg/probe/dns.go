package probe

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"k8s-network-probe/pkg"
	"k8s.io/klog/v2"
	"net"
	"time"
)

func (pm *ProbeManager) DnsProbe(t *ProbeTarget, metricCh chan<- prometheus.Metric) {
	klog.Infof("DnsProbe.start[target:%+v]", t)
	domain := t.DestAddr
	start := time.Now()
	// 这个秒数一到 ctx就会执行ctx.Done
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(t.ProbeTw))
	defer cancel()
	// dns探测的时候 加超时，看看支不支持 带ctx
	result, err := net.DefaultResolver.LookupHost(ctx, domain)

	var (
		value     = pkg.ProbeFailed
		errReason = pkg.ErrReasonEmpty
	)

	if err != nil {
		klog.Errorf("[net.LookupHost.err][domain:%v][err:%v]", domain, err)
		errReason = err.Error()
	} else {
		klog.Infof("[net.LookupHost.success][domain:%v][result:%v]", domain, result)
	}
	if len(result) > 0 {
		value = pkg.ProbeSuccess
	}

	seconds := time.Since(start).Seconds()
	if value == -1 {
		seconds = -1
	}
	mc := prometheus.MustNewConstMetric(M_PROBE_RESULT,
		prometheus.GaugeValue,
		value,
		pm.nodeName,
		pm.localIp,
		pm.runType,
		domain,
		pkg.PROBE_FUNC_DNS,
		errReason,
	)

	mc2 := prometheus.MustNewConstMetric(M_PROBE_DURATION,
		prometheus.GaugeValue,
		seconds,
		pm.nodeName,
		pm.localIp,
		pm.runType,
		pkg.PROBE_FUNC_DNS,
		domain,
	)

	metricCh <- mc
	metricCh <- mc2
	klog.Infof("DnsProbe.end[target:%+v]", t)

}
