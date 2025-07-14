package main

import (
	"context"
	"encoding/json"
	"flag"
	esl "github.com/ning1875/errgroup-signal/signal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s-network-probe/pkg"
	"k8s-network-probe/pkg/probe"
	target_store "k8s-network-probe/pkg/target-store"
	web_handler "k8s-network-probe/pkg/web-handler"
	"k8s.io/klog/v2"
	"net/http"
	"os"
)

// 提供1 get web ：给agent 获取target 给用户上传target

// 存储：target file or k8s-configmap

var (
	httpAddr              string
	kubeConfigPath        string // 如果你采用 cm
	storeCmNameSpace      string //   cm ns
	storeCName            string //  cm name
	storeFilePath         string // 如果采用本地文件，文件位置
	storeType             string // 类型 mysql/file/cm
	reLoadIntervalSeconds int    // 多久从mem 缓存中save到存储中
	defaultTargets        = []probe.ProbeTarget{
		{
			DestAddr: "google.com",
			Func:     "dns",
			ProbeTw:  3,
		},
	}
)

func main() {

	flag.StringVar(&httpAddr, "http.addr", "0.0.0.0:8087", "The http addr ")
	flag.StringVar(&kubeConfigPath, "kubeconfig.path", "kubeconfig", "kubeconfig")
	flag.StringVar(&storeCmNameSpace, "store.cm.ns", "default", "The http addr ")
	flag.StringVar(&storeCName, "store.cm.name", "network-probe-store", "The http addr ")
	flag.StringVar(&storeFilePath, "store.file.path", "network-probe-store.json", "The http addr ")
	flag.StringVar(&storeType, "store.type", pkg.STORE_TYPE_FILE, "The http addr ")
	flag.IntVar(&reLoadIntervalSeconds, "reload.interval.second", 10, "exec tw sec")
	flag.Parse()

	switch storeType {
	case pkg.STORE_TYPE_FILE:

		// 做准备工作
		// 如果本地没有这个文件，我就创建一个json文件作为store
		// 给他加上默认探测的google

		_, err := os.Stat(storeFilePath) //os.Stat获取文件信息
		if err != nil {
			// 没有的话就生成一个默认的json文件
			data, err := json.Marshal(defaultTargets)
			if err != nil {
				klog.Errorf("defaultTargets.json.marshal.err.err:%v", err)
				return
			}
			err = os.WriteFile(storeFilePath, data, 0644)
			if err != nil {
				klog.Errorf("os.os.WriteFile.err:%v", err)
				return
			}

		} else {
			// 如果有这个文件，读取内容，报错返回
			content, err := os.ReadFile(storeFilePath)
			if err != nil {
				return
			}
			if len(content) == 0 {
				// 没有的话就生成一个默认的json文件
				data, err := json.Marshal(defaultTargets)
				if err != nil {
					klog.Errorf("defaultTargets.json.marshal.err.err:%v", err)
					return
				}
				err = os.WriteFile(storeFilePath, data, 0644)
				//_, err = os.Create(storeFilePath)
				if err != nil {
					klog.Errorf("os.os.WriteFile.err:%v", err)
					return
				}
			}
		}
	default:
		return
	}

	// 初始化store 的options
	storeOptions := &target_store.StoreOptions{
		StoreFilePath:  storeFilePath,
		KubeconfigPath: kubeConfigPath,
		ConfigMapName:  storeCName,
		ConfigMapNs:    storeCmNameSpace,
		Type:           storeType,
	}

	store, err := target_store.NewStore(storeOptions)
	if err != nil {
		klog.Errorf("target_store.NewStore.err:%v", err)
		return
	}
	// 注册metric
	prometheus.MustRegister(store)

	group, stopChan := esl.SetupStopSignalContext()
	ctxAll, cancelAll := context.WithCancel(context.Background())

	group.Go(func() error {
		klog.Infof("[stop chan watch start backend]")
		for {
			select {
			case <-stopChan:
				klog.Infof("[stop chan receive quite signal exit]")
				cancelAll()
				return nil
			}

		}
	})

	// target manager 定时从存储中重载targets到 mem-cache
	group.Go(func() error {
		klog.Infof("[RefreshTargetManager start backend]")
		err := store.ReLoadTargetManager(ctxAll)
		if err != nil {
			klog.Errorf("[RefreshTargetManager.error][err:%v]", err)

		}
		return err
	})

	group.Go(func() error {

		klog.Infof("[metrics web start backend]")

		webRunFunc := func() error {
			apiHandler, _ := web_handler.NewHandler(store)
			http.Handle("/api/", apiHandler)

			http.Handle("/metrics", promhttp.Handler())
			srv := http.Server{Addr: httpAddr}
			err = srv.ListenAndServe()
			if err != nil {
				klog.Errorf("[metrics.web.error][err:%v]", err)

			}
			return err

		}
		errChan := make(chan error, 1)
		go func() {
			errChan <- webRunFunc()
		}()
		select {
		case err = <-errChan:
			klog.Errorf("[web.server.error][err:%v]", err)
			return err
		case <-ctxAll.Done():
			klog.Info("receive.quit.signal.web.server.exit")
			return nil
		}

	})

	if err = group.Wait(); err != nil {
		klog.Fatal(err)
	}
}
