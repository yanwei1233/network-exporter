package target_store

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"k8s-network-probe/pkg/probe"
	"k8s.io/klog/v2"
)

// 存储的接口
type TargetStore interface {
	// 新增/更新探测目标 ：对应的是用户 post /probe-targets 来新增/更新探测目标的 view函数底层的方法
	// 更新底层存储，而不是mem-cache ，避免宕机
	UpdateTargets(ts []*probe.ProbeTarget) (addNum, totalNum int, err error)
	// 给用户 或者 agent 来 获取探测目标的  get /probe-targets
	GetTargets() (ts []*probe.ProbeTarget)
	// 重载的管理器
	ReLoadTargetManager(ctx context.Context) error
	// 重载的方法： mem-cache 和 底层存储 定时从底层存储中把数据加载到mem-cache中
	Load() (ts []*probe.ProbeTarget, err error)
	// prometheus sdk所必须要实现的接口
	Describe(ch chan<- *prometheus.Desc)
	Collect(metricCh chan<- prometheus.Metric)
}

type StoreOptions struct {
	StoreFilePath  string
	KubeconfigPath string
	ConfigMapName  string
	ConfigMapNs    string
	Type           string
}

const (
	STORE_TYPE_FILE = "file"
	STORE_TYPE_CM   = "configmap"
)

// 根据类型初始化
func NewStore(ss *StoreOptions) (st TargetStore, err error) {

	switch ss.Type {
	case STORE_TYPE_FILE:
		if ss.StoreFilePath == "" {
			msg := "NewStore.err.STORE_TYPE_FILE.must.provide.filePath"
			err = fmt.Errorf(msg)
			klog.Errorf(msg)
			return
		}

		st = &FileStore{
			FileName: ss.StoreFilePath,
		}
	case STORE_TYPE_CM:

	}

	return
}
