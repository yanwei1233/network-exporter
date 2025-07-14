package target_store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"k8s-network-probe/pkg/probe"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"os"
	"sync"
	"time"
)

var (
	MPROBE_TARGETS = prometheus.NewDesc(
		"network_probe_server_targets",
		"network_probe_server_targets",
		[]string{
			"target",
			"func",
			"tw",
		},
		nil)
)

// json file
type FileStore struct {
	FileName              string               //文件的位置
	reLoadIntervalSeconds int                  // 重载的间隔秒数
	Targets               []*probe.ProbeTarget //targets缓存
	sync.RWMutex
}

func (fs *FileStore) Describe(ch chan<- *prometheus.Desc) {
	ch <- MPROBE_TARGETS
}

func (fs *FileStore) Collect(metricCh chan<- prometheus.Metric) {
	// 从存储中加载回来
	ts, err := fs.Load()
	if err != nil {
		return
	}
	if len(ts) == 0 {
		return
	}
	for _, t := range ts {
		t := t
		mc2 := prometheus.MustNewConstMetric(MPROBE_TARGETS,
			prometheus.GaugeValue,
			1,
			t.DestAddr,
			t.Func,
			fmt.Sprintf("%d", t.ProbeTw),
		)
		metricCh <- mc2
	}
}

// 从存储中重载回来
// 真实的实现就是 读取json文件，加载到内存中
func (fs *FileStore) Load() (ts []*probe.ProbeTarget, err error) {
	fs.Lock()
	defer fs.Unlock()
	var content []byte
	content, err = os.ReadFile(fs.FileName)
	if err != nil {
		klog.Errorf("[FileStore.Load.err:%v][fileName:%v]", err, fs.FileName)
		return
	}

	err = json.Unmarshal(content, &ts)
	if err != nil {
		klog.Errorf("[FileStore.json.Unmarshal.err:%v]", err)
		return
	}
	fs.Targets = ts
	return
}

// 更新
func (fs *FileStore) UpdateTargets(ts []*probe.ProbeTarget) (addNum, totalNum int, err error) {
	fs.Lock()
	defer fs.Unlock()
	var content []byte
	// 首先要读取原来的文件内容
	content, err = os.ReadFile(fs.FileName)
	if err != nil {
		klog.Errorf("[FileStore.UpdateTargets.ReadFile.err:%v]", err)
		return
	}

	var oldTs []*probe.ProbeTarget
	err = json.Unmarshal(content, &oldTs)
	if err != nil {
		klog.Errorf("[FileStore.UpdateTargets.json.Unmarshal.err:%v]", err)
		return
	}

	// 去重
	cM := map[string]*probe.ProbeTarget{}
	for _, t := range oldTs {
		t := t
		uniqueName := fmt.Sprintf("%s_%s", t.DestAddr, t.Func)
		cM[uniqueName] = t
	}

	for _, t := range ts {
		t := t
		uniqueName := fmt.Sprintf("%s_%s", t.DestAddr, t.Func)
		if _, exist := cM[uniqueName]; !exist {
			addNum++
			cM[uniqueName] = t
		}

	}

	// 到这里 cM 是本地和更新的 并集
	fTs := make([]*probe.ProbeTarget, 0)
	for _, v := range cM {
		v := v
		fTs = append(fTs, v)
	}
	var data []byte
	data, err = json.Marshal(fTs)
	if err != nil {
		klog.Errorf("[FileStore.UpdateTargets.json.Marshal.err:%v]", err)
		return
	}
	totalNum = len(fTs)

	err = os.WriteFile(fs.FileName, data, 0644)
	if err != nil {
		klog.Errorf("[FileStore.UpdateTargets.WriteFile.err:%v]", err)
		return
	}
	return
}

func (fs *FileStore) GetTargets() []*probe.ProbeTarget {
	fs.RLock()
	defer fs.RUnlock()
	return fs.Targets
}

func (fs *FileStore) ReLoadTargetManager(ctx context.Context) error {
	go wait.UntilWithContext(ctx, fs.ReLoadTarget, time.Duration(fs.reLoadIntervalSeconds)*time.Second)
	<-ctx.Done()
	klog.Infof("ReLoadTargetManager.exit.receive_quit_signal")
	return nil

}

func (fs *FileStore) ReLoadTarget(ctx context.Context) {
	fs.Load()
}
