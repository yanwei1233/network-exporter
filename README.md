# 简介
k8s-network-probe网络探测,用于探测k8s内各项服务指标是否正常，包含tcp、udp、ping、dns、http探测,采用serviceMonitor方式管理，
集成grafana可视化监控指标，可自定义alermanagerRule配合webhook进行告警。
### 注意事项
1、/bin/build.sh 请在docker环境下运行，且修改文件内容，修改包括harbor地址等等。

2、/bin/run.sh 请在k8s环境下运行

3、/alertmanager/alertmanager.yaml 修改文件内容，改webhook的具体部署地址和token。

4、/webhook/hook_project 程序建议在docker下运行，非k8s环境，能连接外网，目前url是企业微信的webhook，只需传入token即可。运行启动docker-compose up -d
### 编译
```shell
chmod +x ./bin/build.sh
./build.sh
```
### 启动
```shell
chmod+x ./bin/build.sh
./run.sh
```

### 代码结构
```api
├─alertmanager //alertmanager文件
├─bin          //项目启动文件
├─cmd          //network-exporter的client和server文件
│  ├─agent     
│  └─server
├─deploy       //network-exporterdockerfile和k8s.yaml文件
├─grafana      //grafana的json文件
├─pkg          //network-exporter项目文件
│  ├─probe     
│  ├─target-store
│  ├─utils
│  └─web-handler
└─webhook      //webhook告警代码文件
    ├─bin
    └─hook_project
```
### 项目架构
![img.png](img.png)

### 展示
（1）prometheus-target
![img_1.png](img_1.png)
（2）grafana
![img_2.png](img_2.png)
![img_3.png](img_3.png)