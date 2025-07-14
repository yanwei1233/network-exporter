#!/usr/bin/env bash


CGO_ENABLED=0  GOOS=linux GOARCH=amd64 go build -o k8s-network-probe-server cmd/main.go

docker build -t registry.cn-beijing.aliyuncs.com/ning1875_haiwai_image/k8s-network-probe-server:v1.0 -f   deploy/server.Dockerfile .

docker build -t registry.cn-beijing.aliyuncs.com/ning1875_haiwai_image/k8s-network-probe-agent:v1.0 -f   deploy/agent.Dockerfile .

docker push registry.cn-beijing.aliyuncs.com/ning1875_haiwai_image/k8s-network-probe-server:v1.0
docker push registry.cn-beijing.aliyuncs.com/ning1875_haiwai_image/k8s-network-probe-agent:v1.0

kubectl create ns k8s-network-probe
kubectl delete -f dep.yaml
kubectl apply -f dep.yaml
kubectl apply -f ds.yaml


