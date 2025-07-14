package main

import (
	"context"
	"flag"
	esl "github.com/ning1875/errgroup-signal/signal"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s-network-probe/pkg"
	"k8s-network-probe/pkg/probe"
	"k8s.io/klog/v2"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
)

var (
	serverAddr string

	httpAddr                  string
	nodeName                  string
	localIp                   string
	runType                   string
	probeGlobalTimeoutSeconds int
	refreshIntervalSeconds    int
)

func main() {

	flag.StringVar(&httpAddr, "http.addr", "0.0.0.0:8088", "The http addr ")
	flag.StringVar(&serverAddr, "server.addr", "http://localhost:8087", "The server addr ")
	flag.StringVar(&localIp, "local.ip", "", "container ip")
	flag.StringVar(&runType, "run.type", pkg.RPOBE_RUN_TYPE_HOST, "host or container")
	flag.IntVar(&probeGlobalTimeoutSeconds, "probe.global.timeout.second", 3, "exec tw sec")
	flag.IntVar(&refreshIntervalSeconds, "refresh.interval.second", 10, "exec tw sec")
	flag.Parse()
	// 根据探针类型设置 这几个值
	localIp = GetLocalIp()

	// 如果是主机的，主机名就是 os.HostName
	switch runType {
	case pkg.RPOBE_RUN_TYPE_HOST:
		nodeName, _ = os.Hostname()
	default:
		// 从环境变量中读取
		nodeName = os.Getenv("MY_NODE_NAME")
	}
	if nodeName == "" {
		nodeName, _ = os.Hostname()
	}
	//fmt.Println(runType, nodeName, localIp)
	// 自定义exporter核心实现prometheus的collector
	pm := probe.NewProbeManager(serverAddr, nodeName, localIp, runType, probeGlobalTimeoutSeconds, refreshIntervalSeconds)
	prometheus.Register(pm)

	// grouting优雅退出
	group, stopChan := esl.SetupStopSignalContext()
	ctxAll, cancelAll := context.WithCancel(context.Background())

	// group.Go 里面的唯一func 参数 含义 ：要是一个一直运行的函数，一单退出返回错误，其他人也要退出
	group.Go(func() error {
		// 监听退出信号的
		klog.Infof("[stop chan watch start backend]")
		for {
			select {
			case <-stopChan: // 有人kill了这个进程 ，通过上面的ctx 传播通知到其他goroutine
				klog.Infof("[stop chan receive quite signal exit]")
				cancelAll()
				return nil
			}

		}
	})
	//开始监听metrics接口信息
	group.Go(func() error {

		klog.Infof("[metrics web start backend]")

		webRunFunc := func() error {
			http.Handle("/metrics", promhttp.Handler())
			srv := http.Server{Addr: httpAddr}
			err := srv.ListenAndServe()
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
		case err := <-errChan:
			klog.Errorf("[web.server.error][err:%v]", err)
			return err
		case <-ctxAll.Done():
			klog.Info("receive.quit.signal.web.server.exit")
			return nil
		}

	})
	//热更新target配置，获取server的target数据
	group.Go(func() error {
		klog.Infof("[RefreshTargetManager start backend]")
		err := pm.RefreshTargetManager(ctxAll)
		if err != nil {
			klog.Errorf("[RefreshTargetManager.error][err:%v]", err)

		}
		return err
	})

	if err := group.Wait(); err != nil {
		klog.Fatal(err)
	}

}

func GetLocalIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		log.Printf("get local addr err:%v", err)
		return ""
	} else {
		localIp = strings.Split(conn.LocalAddr().String(), ":")[0]
		conn.Close()
		return localIp
	}
}
