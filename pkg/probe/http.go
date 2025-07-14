package probe

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"k8s-network-probe/pkg"
	"k8s.io/klog/v2"
	"net/http"
	"time"
)

func (pm *ProbeManager) HttpProbe(t *ProbeTarget, metricCh chan<- prometheus.Metric) {
	klog.Infof("HttpProbe.start[target:%+v]", t)
	start := time.Now()
	target := t.DestAddr
	var (
		result    = pkg.ProbeFailed
		seconds   float64
		errReason = pkg.ErrReasonEmpty
	)

	client := &http.Client{
		Timeout: time.Duration(seconds) * time.Second,
	}
	resp, err := client.Get(t.DestAddr)
	if err != nil {
		klog.Errorf("[HttpProbe.client.Get.err][url:%v][err:%v]", t.DestAddr, err)
		errReason = err.Error()
	} else {
		// 2xx认为成功
		if resp.StatusCode > 300 {
			errReason = fmt.Sprintf("StatusCode.%d.err", resp.StatusCode)
		} else {
			result = pkg.ProbeSuccess
		}

		seconds = time.Since(start).Seconds()
		// 成功打开http之后一定要closet body
		defer resp.Body.Close()
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
		pkg.PROBE_FUNC_HTTP,
		errReason,
	)

	mc2 := prometheus.MustNewConstMetric(M_PROBE_DURATION,
		prometheus.GaugeValue,
		seconds,
		pm.nodeName,
		pm.localIp,
		pm.runType,
		pkg.PROBE_FUNC_HTTP,
		target,
	)

	metricCh <- mc
	metricCh <- mc2
	klog.Infof("HttpProbe.end[target:%+v]", t)
}
