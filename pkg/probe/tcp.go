package probe

import (
	"github.com/prometheus/client_golang/prometheus"
	"k8s-network-probe/pkg"
	"k8s.io/klog/v2"
	"net"
	"time"
)

// 传参：探测哪个目标+ 写结果值的ch
func (pm *ProbeManager) TcpProbe(t *ProbeTarget, metricCh chan<- prometheus.Metric) {
	klog.Infof("TcpProbe.start[target:%+v]", t)
	start := time.Now()
	target := t.DestAddr
	var (
		result    float64              // 探测成功与否 =-1 代表失败 =1 代表成功
		seconds   float64              // 探测耗时多少秒 0.1 代表100毫秒
		errReason = pkg.ErrReasonEmpty // 错误原因：connection refused
	)
	conn, err := net.DialTimeout("tcp", target, time.Second*time.Duration(t.ProbeTw))
	result = pkg.ProbeSuccess
	if err != nil {
		klog.Errorf("[net.DialTimeout.err][target:%v][err:%v]", target, err)
		result = pkg.ProbeFailed
		errReason = err.Error()
		//return
	} else {
		conn.Close()
		seconds = time.Since(start).Seconds()
	}
	// 如果探测失败了，那么 耗时我也给他设置个-1
	if result == -1 {
		seconds = -1
	}
	//4种参数 desc对象，counter/gauge,float64值，标签值 ，顺序别错了

	mc := prometheus.MustNewConstMetric(M_PROBE_RESULT,
		prometheus.GaugeValue,
		result,
		pm.nodeName,
		pm.localIp,
		pm.runType,
		target,
		pkg.PROBE_FUNC_TCP,
		errReason,
	)

	mc2 := prometheus.MustNewConstMetric(M_PROBE_DURATION,
		prometheus.GaugeValue,
		seconds,
		pm.nodeName,
		pm.localIp,
		pm.runType,
		pkg.PROBE_FUNC_TCP,
		target,
	)

	metricCh <- mc
	metricCh <- mc2
	klog.Infof("TcpProbe.end[target:%+v]", t)
}
